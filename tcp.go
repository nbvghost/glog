package glog

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"net"
)

type TCP struct {
	ServerAddr *net.TCPAddr
	ServerConn *net.TCPConn
	reConnectChan chan bool
	status chan<- bool
}
func (tcp *TCP)StartTCP(address string,status chan<- bool)  {
	tcp.status =status
	tcp.reConnectChan =make(chan bool,10)
	var err error
	tcp.ServerAddr,err=net.ResolveTCPAddr("tcp",address)
	if err!=nil{
		log.Panicln(err)
		return
	}

	for   {
		select {
			case <-tcp.reConnectChan:
				if tcp.ServerConn==nil{
					tcp.connect()
				}
		}
	}



}
func (tcp *TCP)connect()  {
	var err error
	tcp.ServerConn,err=net.DialTCP("tcp",nil,tcp.ServerAddr)
	if err!=nil{
		Debug(err)
		return
	}
	tcp.ServerConn.SetKeepAlive(true)
	tcp.status<-true

}
func (tcp *TCP)Write(log string) error  {
	var bodyLen int32

	bodyBuffer := bytes.NewBuffer([]byte{})
	bodyBuffer.WriteString(log)

	body:=bodyBuffer.Bytes()
	bodyLen=int32(len(body))

	packageBuffer := bytes.NewBuffer([]byte{})
	binary.Write(packageBuffer, binary.BigEndian, &bodyLen)
	binary.Write(packageBuffer, binary.BigEndian, body)

	if tcp.ServerConn!=nil{
		_,err:=tcp.ServerConn.Write(packageBuffer.Bytes())
		if err!=nil{
			tcp.ServerConn.Close()
			tcp.ServerConn=nil
			tcp.reConnectChan<-true
			tcp.status<-false
			return err
		}else {
			return nil
		}

	}else{
		tcp.reConnectChan<-true
		tcp.status<-false
		return errors.New("网络错误，在重连")
	}


}