package mqtt_server

import (
	"bytes"
	"errors"
	"fmt"
	"go-connector/logger"
	"go.uber.org/zap"
	"net"
	"sync"
	"sync/atomic"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

type PublishPacket = packets.PublishPacket
type PublishHandle func(*Conn, string, uint16, []byte)

type Client struct {
	ID   string
	Conn net.Conn
}
type Server struct {
	addr                net.Addr
	publishMessages     chan interface{}
	subscribeMessages   chan interface{}
	unsubscribeMessages chan interface{}
	Clients             sync.Map
	ClientsCount        int64
	Topics              map[string]interface{}
	onPublishHandle     PublishHandle
}
type Conn struct {
	conn net.Conn
}

func (c *Conn) Reply(b []byte) (err error) {
	buf := new(bytes.Buffer)
	p := packets.PublishPacket{FixedHeader: packets.FixedHeader{MessageType: packets.Publish}, Payload: b, TopicName: "rpc"}
	if err = writeConn(&p, c.conn); err != nil {
		_, err = c.conn.Write(buf.Bytes())
	}
	return
}

func (s *Server) Publish(_ string, _ string) {

}

func (s *Server) OnSubscribe(f interface{}) {
	go s.registSubscribe(f)

}

func (s *Server) OnUnSubscribe(f interface{}) {
	go s.registUnSubscribe(f)

}

func (s *Server) OnPublish(f PublishHandle) {
	s.onPublishHandle = f
	//go s.registPublish(f)

}

func (s *Server) registSubscribe(f interface{}) {

	if v, ok := f.(func(m string)); ok {
		select {
		case m := <-s.subscribeMessages:
			fmt.Println("server", m)
			v(string(m.([]byte)))
		}
	}

}

func (s *Server) registUnSubscribe(f interface{}) {

	if v, ok := f.(func(m string)); ok {
		for {
			select {
			case m := <-s.unsubscribeMessages:
				// fmt.Println("server", m)
				v(m.(string))
			}
		}
	}

}

//func (s *Server) registPublish(f interface{}) {
//
//	if v, ok := f.(func(m []byte)); ok {
//		for {
//			select {
//			case m := <-s.publishMessages:
//				v(m.([]byte))
//			}
//		}
//
//	}
//
//}

func (s *Server) Addr() net.Addr {
	return s.addr
}

func (s *Server) New(addr string) error {
	ln, err := net.Listen("tcp", addr)

	if err != nil {
		return err
		// handle error
	}
	s.subscribeMessages = make(chan interface{})
	//s.publishMessages = make(chan interface{})

	s.addr = ln.Addr()
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("err=>", err)
				// handle error
			}
			go s.handleConnection(conn)
		}
	}()
	return nil
}

func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		e := conn.Close()
		logger.DEBUG.Println("mqtt_server", "mqtt server closing", zap.Error(e))
	}()

	clientID, err := s.handleConnect(conn)
	if err != nil {
		return
	}

	for {
		p, err := packets.ReadPacket(conn)
		if err != nil {
			fmt.Printf("Error reading packet: %s", err.Error())
			return
		}
		switch messType := p.(type) {
		case *packets.ConnectPacket:
			//???? 是吗? MQTT协议规定要断开重复的
			//conn.Close()
		case *packets.SubscribePacket:
			s.handleSubscribe(clientID, p.(*packets.SubscribePacket), conn)
		case *packets.DisconnectPacket:
			s.handleDisconnect(conn)
		case *packets.PingreqPacket:
			s.handlePingreq(conn)
		case *packets.PublishPacket:
			s.handlePublish(clientID, p.(*packets.PublishPacket), conn)
		case *packets.UnsubscribePacket:
			s.handleUnSubscribe(clientID, p.(*packets.UnsubscribePacket), conn)
		default:
			fmt.Print("maybe wrong?", messType)
		}

	}
}
func (s *Server) handleSubscribe(_ string, p *packets.SubscribePacket, conn net.Conn) {

	for _, topic := range p.Topics {
		s.subscribeMessages <- topic
	}

	var ack = &packets.SubackPacket{FixedHeader: packets.FixedHeader{MessageType: packets.Suback}}
	ack.ReturnCodes = append(ack.ReturnCodes, p.Qoss...)

	_ = writeConn(ack, conn)

}
func (s *Server) handleDisconnect(conn net.Conn) {
	//todo : drop messages of public
	_ = conn.Close()
}
func (s *Server) handlePingreq(conn net.Conn) {
	var ack = &packets.SubackPacket{FixedHeader: packets.FixedHeader{MessageType: packets.Pingresp}}
	_ = writeConn(ack, conn)
}
func (s *Server) handlePublish(_ string, p *packets.PublishPacket, conn net.Conn) {
	if s.onPublishHandle != nil {
		go s.onPublishHandle(&Conn{conn}, p.TopicName, p.MessageID, p.Payload)
	}

	var ack = &packets.PubackPacket{FixedHeader: packets.FixedHeader{MessageType: packets.Puback}}
	_ = writeConn(ack, conn)
}
func (s *Server) handleUnSubscribe(_ string, p *packets.UnsubscribePacket, conn net.Conn) {

	for _, topic := range p.Topics {
		s.unsubscribeMessages <- topic
	}

	var ack = &packets.SubackPacket{FixedHeader: packets.FixedHeader{MessageType: packets.Unsuback}}

	_ = writeConn(ack, conn)

}
func (s *Server) handleConnect(conn net.Conn) (id string, error error) {

	p, err := packets.ReadPacket(conn)
	if err != nil {
		error = err
		return
	}
	cp := p.(*packets.ConnectPacket)

	var ack = &packets.ConnackPacket{FixedHeader: packets.FixedHeader{MessageType: packets.Connack}}
	ack.ReturnCode = 0x0

	if _, ok := s.Clients.Load(cp.ClientIdentifier); ok {
		ack.ReturnCode = 0x02
		error = errors.New("ID has been registered")
	}

	// todo : 验证字段 , cp.Username  cp.Password ...
	//if "验证不通过" != "验证不通过" {
	//	ack.ReturnCode = 0x04
	//	error = errors.New("auth fail")
	//}

	if err := writeConn(ack, conn); err != nil {
		error = err
		return
	}

	if ack.ReturnCode != 0 {
		return
	}
	s.Clients.Store(cp.ClientIdentifier, conn)
	atomic.AddInt64(&s.ClientsCount, 1)
	return cp.ClientIdentifier, nil

}

func writeConn(p packets.ControlPacket, conn net.Conn) error {
	buf := new(bytes.Buffer)
	if err := p.Write(buf); err != nil {
		return err
	}
	if n, e := conn.Write(buf.Bytes()); e != nil {
		logger.ERROR.Println("write to conn failed", zap.Error(e), zap.Int("num", n))
	}

	return nil
}
