package postbird

import (
	"errors"
	"fmt"
	"log"
	"net"
	"reflect"
)

// Info struct
// PostBird 에서 사용될 내용
type Info struct {
	BindPort      uint
	BindAddress   string
	RemotePort    uint
	RemoteAddress string
	Mode          uint
}

const DefaultPort uint = 8787                   // Default Bind Port
const DefaultBindAddress string = "127.0.0.1"   // Default Bind Address
const DefaultServerAddress string = "127.0.0.1" // Defualt Server Address

const (
	ServerMode = 0
	ClientMode = 1
)

var info Info

// funcs map
// 원격에서 호출가능한 함수들을 등록해놓은 map
// RegisterFunc 함수로 이 맵에 등록한다
var funcs map[string]interface{} = make(map[string]interface{})

// SetBindAddress func
// StartServer로 ServerMode 로 실행할때 바인드될 아이피 주소. ""로 설정하면 모든 NIC에 바인딩된다.
// 이 함수를 호출하지 않으면 DefaultBindAddress인 127.0.0.1로 바인딩된다.
func SetBindAddress(BindAddress string) {
	info.BindAddress = BindAddress
}

// SetBindPort func
// StartServer로 ServerMode 로 실행할때 바인드될 포트 번호.
// 이 함수를 호출하지 않으면 DefaultPortd인 8787로 바인딩된다.
func SetBindPort(BindPort uint) {
	info.BindPort = BindPort
}

// SetRemoteAddress func
//
func SetRemoteAddress(ServerAddress string) {
	info.RemoteAddress = ServerAddress
}

func SetRemotePort(ServerPort uint) {
	info.RemotePort = ServerPort
}

func init() {

	if info.BindAddress == "" {
		info.BindAddress = DefaultBindAddress
	}

	if info.BindPort == 0 {
		info.BindPort = DefaultPort
	}

	if info.RemoteAddress == "" {
		info.RemoteAddress = DefaultBindAddress
	}

	if info.RemotePort == 0 {
		info.RemotePort = DefaultPort
	}

}

// RegisterFunc func
// CallLocalFunc 함수에 의해 실행될 수 있는, 즉 원격에서 호출가능한 함수를 등록하는 함수
// funcs 맵에 등록되며 이 함수에 등록되지 않은 함수는 원격에서 호출할 수 없다.
func RegisterFunc(FuncName string, Function interface{}) {
	funcs[FuncName] = Function
}

// StartServer func
// 프로그램을 서버역할로 사용하려면 이 함수를 호출해서 tcp 서버를 시작하면 된다.
// 시작되면 Binder 함수를 비동기로 호출하여 비동기로 tcp Listen
// 이 함수가 호출되면 무조건 Mode가 ServerMode 로 바뀐다
func StartServer() {
	info.Mode = ServerMode
	go Binder(info.BindAddress, info.BindPort)
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

// requestHandler func
// tcp 연결되었을때 request 핸들러
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

func Connect() {

}

// CallLocalFunc func
// RegisterFunc 로 등록된 함수가 원격에서 함수를 호출했을때
// 이함수를 통해 실행된다
func CallLocalFunc(name string, params ...interface{}) (result []reflect.Value, err error) {
	f := reflect.ValueOf(funcs[name])
	if len(params) != f.Type().NumIn() {
		err = errors.New("The number of params is not adapted.")
		return
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	result = f.Call(in)
	return
}

// CallRemoteFunc func
// 연결된 (서버)의 함수를 호출하고 싶을때 사용하는 함수
// json형식으로 변환해서 tcp로 서버에 전달.
func CallRemoteFunc() {

}
