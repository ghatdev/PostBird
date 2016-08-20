package postbird

import (
	"fmt"
	"log"
	"net"
)

type Info struct {
	BindPort      uint
	BindAddress   string
	ServerPort    uint
	ServerAddress string
	Mode          uint
}

const DefaultPort uint = 8787                   // Default Bind Port
const DefaultBindAddress string = "127.0.0.1"   // Default Bind Address
const DefaultServerAddress string = "127.0.0.1" // Defualt Server Address
const DefaultMode uint = ServerMode             // Default Mode : ServerMode (Tcp Listen)

const (
	ServerMode = 0
	ClientMode = 1
)

var info Info

func PostBird() {

}

func (c Info) SetMode(Mode uint) {
	c.Mode = Mode
}

func (c Info) SetBindAddress(BindAddress string) {
	c.BindAddress = BindAddress
}

func (c Info) SetBindPort(BindPort uint) {
	c.BindPort = BindPort
}

func (c Info) SetServerAddress(ServerAddress string) {
	c.ServerAddress = ServerAddress
}

func (c Info) SetServerPort(ServerPort uint) {
	c.ServerPort = ServerPort
}

func (c Info) init() {
	if c.Mode == ServerMode {

		if c.BindAddress == "" {
			c.BindAddress = DefaultBindAddress
		}

		if c.BindPort == 0 {
			c.BindPort = DefaultPort
		}

	} else if c.Mode == ClientMode {
		if c.ServerAddress == "" {
			c.ServerAddress = DefaultBindAddress
		}

		if c.ServerPort == 0 {
			c.ServerPort = DefaultPort
		}

	} else {
		log.Println("Mode is not defined")
	}
}

// Binder func
// ServerMode일때 main func
func Binder(BindAddr string, Port uint) {
	ln, err := net.Listen("tcp", BindAddr+":"+string(Port)) // 전달받은 BindAddr:Port 에 TCP로 바인딩
	if err != nil {
		log.Println(err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		defer conn.Close()

		go requestHandler(conn)
	}
}

func requestHandler(c net.Conn) {
	data := make([]byte, 4096) // 4096 크기의 바이트 슬라이스 생성

	for {
		n, err := c.Read(data) // 클라이언트에서 받은 데이터를 읽음
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(string(data[:n])) // 데이터 출력

		_, err = c.Write(data[:n]) // 클라이언트로 데이터를 보냄
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func CallRemoteFunc() {

}

func CallLocalFunc() {

}
