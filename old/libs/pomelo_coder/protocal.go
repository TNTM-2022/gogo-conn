package coder

import (
	"bytes"
	"fmt"
	"log"
)

// consts
const (
	MsgFlagBytes      = 1
	MsgRouteLenBytes  = 1
	MsgRouteCodeBytes = 2
	PkgHeadBytes      = 4
	MsgIDMaxBytes     = 5

	MsgRouteCodeMax = 0xffff

	MsgCompressRouteMask      = 0x1
	MsgCompressGzipMask       = 0x1
	MsgCompressGzipEncodeMask = 1 << 4
	MsgTypeMask               = 0x7
)

// Package var
var Package = map[string]int{
	"TYPE_HANDSHAKE":     1,
	"TYPE_HANDSHAKE_ACK": 2,
	"TYPE_HEARTBEAT":     3,
	"TYPE_DATA":          4,
	"TYPE_KICK":          5,
}

// Message var
var Message = map[string]int{
	"TYPE_REQUEST":  0,
	"TYPE_NOTIFY":   1,
	"TYPE_RESPONSE": 2,
	"TYPE_PUSH":     3,
}

/*
ProtocolStrEncode func
pomele client encode
id message id;
route message route
msg message body
socketio current support string
*/
func ProtocolStrEncode(s string) []byte {
	// buffer := bytes.NewBufferString(s)
	// return buffer.Bytes()
	return []byte(s)
}

/*
ProtocolStrDecode func
client decode
msg String data
return Message Object
*/
func ProtocolStrDecode(b []byte) string {
	// buffer := bytes.NewBuffer(b)
	// return buffer.String()
	return string(b)
}

/*
PackageEncode func
 * Package protocol encode.
 *
 * Pomelo package format:
 * +------+-------------+------------------+
 * | type | body length |       body       |
 * +------+-------------+------------------+
 *
 * Head: 4bytes
 *   0: package type,
 *      1 - handshake,
 *      2 - handshake ack,
 *      3 - heartbeat,
 *      4 - data
 *      5 - kick
 *   1 - 3: big-endian body length
 * Body: body length bytes
 *
 * @param  {Number}    type   package type
 * @param  {ByteArray} body   body content in bytes
 * @return {ByteArray}        new byte array that contains encode result
*/
func PackageEncode(mType int, body []byte) []byte {
	length := len(body)
	buffer := make([]byte, PkgHeadBytes)
	index := 0
	buffer[index] = uint8(mType & 0xff)
	index++
	buffer[index] = uint8((length >> 16) & 0xff)
	index++
	buffer[index] = uint8((length >> 8) & 0xff)
	index++
	buffer[index] = uint8(length & 0xff)
	if body != nil {
		l := [][]byte{buffer, body}
		return bytes.Join(l, []byte(""))
	}
	return buffer
}

/*
PackageDecode func
 * Package protocol decode.
 * See encode for package format.
 *
 * @param  {ByteArray} buffer byte array containing package content
 * @return {Object}           {type: package type, buffer: body byte array}
*/
//func PackageDecode(b []byte) []MType {
func PackageDecode(b []byte) []MType {
	offset := 0
	rs := []MType{}
	for offset < len(b) {
		mType := b[offset]
		offset++
		l1 := (b[offset]) << 16
		offset++
		l2 := (b[offset]) << 8
		offset++
		l3 := b[offset]
		offset++
		length := int((l1 | l2 | l3) >> 0)
		// length := (uint(l1|l2|l3) >> 0)
		var body []byte
		if length > 0 {
			body = make([]byte, length)

			copyArray(body, 0, b, offset, length)
		}
		offset = offset + length
		rs = append(rs, MType{Type: mType, Body: body})
	}

	return rs
	// return rs.length === 1 ? rs[0] : rs;
}

/*
MessageEncode func
 * Message protocol encode.
 *
 * @param  {Number} id            message id
 * @param  {Number} type          message type
 * @param  {Number} compressRoute whether compress route
 * @param  {Number|String} route  route code or route string
 * @param  {Buffer} msg           message body bytes
 * @return {Buffer}               encode result
*/
func MessageEncode(id uint64, mtype int, compressRoute int, route string, msg []byte, compressGzip bool) []byte {
	// caculate message max length
	var idBytes int
	if msgHasID(mtype) {
		idBytes = caculateMsgIDBytes(id)
	}
	msgLen := MsgFlagBytes + idBytes
	if msgHasRoute(mtype) {
		msgLen += MsgRouteLenBytes
		if route != "" {
			r := ProtocolStrEncode(route)
			if len(r) > 255 {
				fmt.Println("route maxlength is overflow")
				return nil
			}
			msgLen += len(r)
		}
	}
	if len(msg) > 0 {
		msgLen += len(msg)
	}
	buffer := make([]byte, msgLen)
	offset := 0

	// add flag
	offset = encodeMsgFlag(mtype, compressRoute, buffer, offset, compressGzip)
	// add message id

	if msgHasID(mtype) {
		offset = encodeMsgID(id, buffer, offset)
	}
	// add route
	if msgHasRoute(mtype) {
		offset, buffer = encodeMsgRoute(compressRoute, route, buffer, offset)
	}
	// add body
	if len(msg) > 0 {
		offset, buffer = encodeMsgBody(msg, buffer, offset)
	}

	return buffer
}

/*
MessageDecode func
 * Message protocol decode.
 *
 * @param  {Buffer|Uint8Array} buffer message bytes
 * @return {Object}            message object
*/
func MessageDecode(b []byte) DecodedMsg {
	bytesLen := len(b)
	offset := 0
	id := 0
	route := ""
	// parse flag
	flag := b[offset]
	offset++
	compressRoute := flag & MsgCompressRouteMask
	mtype := (flag >> 1) & MsgTypeMask
	compressGzip := (flag >> 4) & MsgCompressGzipMask

	if msgHasId(int(mtype)) {
		i := 0
		l := len(b)
		for {
			m := 0
			if offset < l {
				m = int(b[offset])
			} else {
				break
			}
			id += (m & 0x7f) << uint(7*i)
			offset++
			i++
			if m < 128 {
				break
			}
		}
	}

	// parse route
	if msgHasRoute(int(mtype)) && len(b) > offset {
		//  一定不会进行路由压缩的
		// if (compressRoute) != 0 {
		// 	t1 := (b[offset]) << 8
		// 	offset++
		// 	t2 := b[offset]
		// 	offset++
		// 	route = t1 | t2
		// } else {
		routeLen := int(b[offset])
		offset++
		if routeLen > 0 {
			routeBytes := make([]byte, routeLen)
			copyArray(routeBytes, 0, b, offset, routeLen)
			route = ProtocolStrDecode(routeBytes)
			log.Println("router", route)
		} else {
			route = ""
			log.Println("router", "no")

		}
		offset += routeLen
		// }
	}

	// parse body
	var bodyLen = bytesLen - offset
	var body = make([]byte, bodyLen)

	copyArray(body, 0, b, offset, bodyLen)
	fmt.Println("compressRoute", compressRoute, "mtype", int(mtype))
	return DecodedMsg{
		ID:            int64(id),
		Type:          mtype,
		CompressRoute: compressRoute != 0,
		Route:         route,
		Body:          body,
		CompressGzip:  compressGzip != 0,
	}
}

func ProtocalEncode() {

}


func Compose (t uint8, data []byte, id int64) []byte{
	if t == 0 && len(data) == 0 {
		fmt.Println("data should not be empty.")
		return nil;
	}

	var buf []byte
	var dataLen int64;
	if data != nil {
		dataLen ++;
		lsize := calLengthSize(dataLen)
		buf = make([]byte, lsize + dataLen)
		fillLenght(buf ,uint8(dataLen), lsize)
		buf[lsize] = t;
		var off int64 = lsize + 1;
		for  i := 0 ; i<len(data); i++ {
			buf[off + int64(i)] = data[i]
		}
	} else {
		dataLen = 1;
		lsize := calLengthSize(dataLen)
		buf = make([]byte, lsize + dataLen)
		fillLenght(buf, uint8(dataLen), lsize)
		buf[lsize] = t;
	}

	return buf;
}

func calLengthSize (length int64) int64{
	var res int64
	for length > 0 {
		length >>= 7
		res ++
	}
	return res;
}

var LEFT_SHIFT_BITS uint8 = 1 << 7
func fillLenght (buf []byte, data uint8, size int64) {
	offset := size - 1;
	var b uint8
	for ; offset >= 0; offset-- {
		b = data % LEFT_SHIFT_BITS
		if offset < size - 1 {
			b |= 0x80
		}
		buf[offset] = b;
		data >>= 7
	}
}

/* ProtocalDecode func
*  packageDecode > protocalDecode > messageDecode
*  一定会有路由
 */
// func ProtocalDecode() {
// 	msg = Message.decode(msg.body);
// 	var route = msg.route;
// 	console.log('decode', route)
// 	// decode use dictionary
// 	if(!!msg.compressRoute) {
// 	  if(!!this.connector.useDict) {
// 		var abbrs = this.dictionary.getAbbrs();
// 		if(!abbrs[route]) {
// 		  logger.error('dictionary error! no abbrs for route : %s', route);
// 		  return null;
// 		}
// 		route = msg.route = abbrs[route];
// 	  } else {
// 		logger.error('fail to uncompress route code for msg: %j, server not enable dictionary.', msg);
// 		return null;
// 	  }
// 	}
// 	  try {
// 		  // decode use protobuf
// 		  if (!!this.protobuf && !!this.protobuf.getProtos().client[route]) {
// 			  msg.body = this.protobuf.decode(route, msg.body);
// 		  } else if (!!this.decodeIO_protobuf && !!this.decodeIO_protobuf.check(Constants.RESERVED.CLIENT, route)) {
// 			  msg.body = this.decodeIO_protobuf.decode(route, msg.body);
// 		  } else {
// 			  msg.body = JSON.parse(msg.body.toString('utf8'));
// 		  }
// 	  } catch (ex) {
// 		  console.error('2route:', route, ex, 'isBuffer:', Buffer.isBuffer(msg.body), msg.body.toString());
// 		  msg.body = {};
// 	  }
// 	return msg;

// }
