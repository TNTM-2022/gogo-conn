package pomelo_coder

import (
	"encoding/binary"
)

func msgHasID(mtype byte) bool {
	return mtype == Message["TYPE_REQUEST"] || mtype == Message["TYPE_RESPONSE"]
}

func msgHasRoute(mtype byte) bool {
	return mtype == Message["TYPE_REQUEST"] || mtype == Message["TYPE_NOTIFY"] ||
		mtype == Message["TYPE_PUSH"]
}
func msgHasId(mtype byte) bool {
	return mtype == Message["TYPE_REQUEST"] || mtype == Message["TYPE_RESPONSE"]
}

func caculateMsgIDBytes(id uint64) int {
	l := 0
	for {
		l++
		id >>= 7
		if id <= 0 {
			break
		}
	}
	return l
}

func encodeMsgFlag(mtype byte, compressRoute int, buffer []byte, offset int, compressGzip bool) int {
	if mtype != Message["TYPE_REQUEST"] && mtype != Message["TYPE_NOTIFY"] &&
		mtype != Message["TYPE_RESPONSE"] && mtype != Message["TYPE_PUSH"] {
		// throw new Error('unkonw message type: ' + type);
		return 0
	}
	// buffer[offset] = (mtype << 1) | (compressRoute ? 1 : 0);
	buffer[offset] = byte((mtype << 1) | 0)

	if compressGzip {
		buffer[offset] = buffer[offset] | MsgCompressGzipEncodeMask
	}

	return offset + MsgFlagBytes
}

func encodeMsgID(id uint64, buffer []byte, offset int) int {
	//fmt.Println("id", id, "offset", offset)
	//for {
	//	var tmp = byte(math.Mod(float64(id), 128))
	//	var next = math.Floor(float64(id) / 128)
	//
	//	if next != 0 {
	//		tmp = tmp + 128
	//	}
	//	fmt.Println("offset", offset, "tmp", tmp, "next", next)
	//	buffer[offset] = tmp
	//	offset++
	//
	//	id = uint64(next)
	//	if id < 1 {
	//		break
	//	}
	//}

	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, id)
	for i := 0; i < n; i++ {
		buffer[offset] = buf[i]
		offset++
	}
	return offset
}

func encodeMsgRoute(compressRoute int, route string, buffer []byte, offset int) (int, []byte) {
	if route != "" {
		buffer[offset] = byte(len(route) & 0xff)
		offset++
		copyArray(buffer, offset, []byte(route), 0, len(route))
		offset += len(route)
	} else {
		buffer[offset] = 0
		offset++
	}

	return offset, buffer
}

func encodeMsgBody(msg []byte, buffer []byte, offset int) (int, []byte) {
	copyArray(buffer, offset, msg, 0, len(msg))
	return offset + len(msg), buffer
}

func copyArray(dest []byte, doffset int, src []byte, soffset, length int) {
	for index := 0; index < length; index++ {
		dest[doffset] = src[soffset]
		doffset++
		soffset++
	}
}
