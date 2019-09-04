package glog

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"net"
	"time"
)

type GlogTCP struct {
	ServerAddr *net.TCPAddr
	ServerConn *net.TCPConn
	ServerAble bool
	status     chan<- bool
	isClose    bool
}

func (tcp *GlogTCP) Close() {
	tcp.isClose = true
	if tcp.ServerAddr != nil {
		tcp.ServerConn.Close()
	}

}
func (tcp *GlogTCP) StartTCP(address string, status chan<- bool) {
	tcp.status = status
	//tcp.reConnectChan =make(chan bool,10)
	var err error
	tcp.ServerAddr, err = net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.Panicln(err)
		return
	}

	for {
		if tcp.isClose == false {
			err := tcp.connect()
			if err != nil {
				log.Println(err)
			}
			time.Sleep(time.Second)
		} else {
			break
		}

	}

}
func (tcp *GlogTCP) connect() error {
	var err error
	tcp.ServerConn, err = net.DialTCP("tcp", nil, tcp.ServerAddr)
	if err != nil {
		//Debug(err)
		return err
	}
	tcp.ServerConn.SetKeepAlive(true)
	tcp.ServerAble = true
	tcp.status <- true

	defer func() {
		tcp.status <- false
		tcp.ServerAble = false
	}()

	b := make([]byte, 0)

	for {
		_, err := tcp.ServerConn.Read(b)
		if err != nil {
			return err
		}
		time.Sleep(time.Second)

	}

}
func (tcp *GlogTCP) Write(log string) error {

	if tcp.ServerAble == false {
		//tcp.status<-false
		return errors.New("网络错误，在重连")
	}

	var bodyLen int32

	bodyBuffer := bytes.NewBuffer([]byte{})
	bodyBuffer.WriteString(log)

	body := bodyBuffer.Bytes()
	bodyLen = int32(len(body))

	packageBuffer := bytes.NewBuffer([]byte{})
	binary.Write(packageBuffer, binary.BigEndian, &bodyLen)
	binary.Write(packageBuffer, binary.BigEndian, body)

	wb := packageBuffer.Bytes()

	if tcp.ServerConn != nil && len(wb) > 0 {
		_, err := tcp.ServerConn.Write(wb)
		if err != nil {
			tcp.ServerConn.Close()
			tcp.ServerConn = nil
			//tcp.reConnectChan<-true
			tcp.status <- false
			return err
		} else {
			return nil
		}

	} else {
		//tcp.reConnectChan<-true
		tcp.status <- false
		return errors.New("网络错误，在重连")
	}

}
