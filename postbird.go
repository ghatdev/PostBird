package postbird

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/googollee/go-socket.io"
)

// Info struct
// PostBird 에서 사용될 값들
type Info struct {
	BindPort      uint
	BindAddress   string
	RemotePort    uint
	RemoteAddress string
	Mode          uint
	Protocol      uint
}

type Client struct {
	Socket     socketio.Socket
	Connection net.Conn
	ClientID   string
}

type Any interface{}

type CallEvent struct {
	FunctionName string
	Params       []Any
}

type CalledEvent struct {
	FunctionName string
	Params       json.RawMessage
}

const DefaultPort uint = 8787                   // Default Bind Port
const DefaultBindAddress string = "127.0.0.1"   // Default Bind Address
const DefaultRemoteAddress string = "127.0.0.1" // Defualt Server Address
const DefaultProtocol uint = SocketIO

const (
	ServerMode = 0
	ClientMode = 1
)

const (
	TCP      = 0
	SocketIO = 1
)

var info Info
var ServerConnection net.Conn

var isConnected bool
var Clients []Client = make([]Client, 5)

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

func SetProtocol(Protocol uint) {
	info.Protocol = Protocol
}

func init() {

	if info.BindAddress == "" {
		info.BindAddress = DefaultBindAddress
	}

	if info.BindPort == 0 {
		info.BindPort = DefaultPort
	}

	if info.RemoteAddress == "" {
		info.RemoteAddress = DefaultRemoteAddress
	}

	if info.RemotePort == 0 {
		info.RemotePort = DefaultPort
	}

	if info.Protocol == 0 {
		info.Protocol = DefaultProtocol
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
func StartServer(Protocol uint) {
	var wg sync.WaitGroup

	info.Mode = ServerMode
	info.Protocol = Protocol

	switch Protocol {
	case TCP:
		wg.Add(1)
		go Binder(&wg, info.BindAddress, info.BindPort)
	case SocketIO:
		wg.Add(1)
		go Listener(&wg, info.BindAddress, info.BindPort)
	default:
		log.Println("Protocol not match. 0 for TCP, 1 for Socket.io.")
	}
	wg.Wait()
}

// Listener func
// ServerMode 일때 tcp대신 socket.io 사용
func Listener(wg *sync.WaitGroup, BindAddr string, Port uint) {
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	server.On("connection", func(so socketio.Socket) {
		Clients = append(Clients, Client{so, nil, so.Id()})

		so.On("call", func(FunctionName string, args ...string) {
			//go CallLocalFunc(FunctionName, args)
		})

		var i int

		so.On("disconnection", func() {
			for i = 0; i < len(Clients); i++ {
				if so.Id() == Clients[i].ClientID {
					break
				}
			}

			if !(i > len(Clients)) {
				copy(Clients[i:], Clients[i+1:])
				Clients[len(Clients)-1] = Client{}
				Clients = Clients[:len(Clients)-1]
			}

		})
	})

	http.Handle("/socket.io/", server)
	http.Handle("/", http.FileServer(http.Dir("./asset")))
	http.ListenAndServe(BindAddr+":"+string(Port), nil)
}

// Binder func
// ServerMode일때 main func
func Binder(wg *sync.WaitGroup, BindAddr string, Port uint) {
	defer wg.Done()
	info.Protocol = TCP

	var WaitHandler sync.WaitGroup

	Addr, err := net.ResolveTCPAddr("tcp", BindAddr+":"+fmt.Sprint(Port))
	if err != nil {
		log.Println(err)
		return
	}

	ln, err := net.ListenTCP("tcp", Addr) // 전달받은 BindAddr:Port 에 TCP로 바인딩
	if err != nil {
		log.Println(err)
		return
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			panic(err)
		}
		defer conn.Close()

		rand.Seed(time.Now().UnixNano())

		ClientId := RandStringRunes(17)
		Clients = append(Clients, Client{nil, conn, ClientId})
		WaitHandler.Add(1)
		go requestHandler(&WaitHandler, conn)
	}

	WaitHandler.Wait()
}

// requestHandler func
// tcp 연결되었을때 request 핸들러
func requestHandler(wg *sync.WaitGroup, c net.Conn) {
	defer wg.Done()
	data := json.NewDecoder(c)

	var FuncWaiter sync.WaitGroup
	var event CalledEvent

	for {
		err := data.Decode(&event)
		if err != nil {
			log.Println("Invalid json format")
			return
		}
		FuncWaiter.Add(1)
		go CallLocalFunc(&FuncWaiter, event.FunctionName, event.Params)
	}

	FuncWaiter.Wait()
}

func ConnectToRemote(Protocol uint) {
	info.Mode = ClientMode
	info.Protocol = Protocol

	client, err := net.Dial("tcp", info.RemoteAddress+":"+fmt.Sprint(info.RemotePort))
	if err != nil {
		fmt.Println(err)
		return
	}
	ServerConnection = client
}

// CallLocalFunc func
// RegisterFunc 로 등록된 함수가 원격에서 함수를 호출했을때
// 이함수를 통해 실행된다
func CallLocalFunc(wg *sync.WaitGroup, name string, params ...Any) (result []reflect.Value, err error) {
	defer wg.Done()
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
func CallRemoteFunc(FunctionName string, args ...Any) {
	var i int

	Event := CallEvent{"test", []Any{}}

	switch info.Protocol {
	case TCP:

		/*
			tmpStr := "{\"funcname\":\"" + FunctionName + "\"" + "\"args\":["
			for i = 0; i < len(args)-1; i++ {
				switch args[i].(type) {
				case string:
					tmpStr += "\"" + args[i].(string) + "\","
					break
				default:
					tmpStr += args[i].(string) + ","
				}
			}

			switch args[i].(type) {
			case string:
				tmpStr += "\"" + args[i].(string) + "\""
				break
			default:
				tmpStr += args[i].(string)
			}

			tmpStr += "]}"
		*/
		//call, _ := json.Marshal(Event)

		//fmt.Println(string(call))

		if info.Mode == ServerMode {
			if Clients[0].ClientID != "" {
				for i = 0; i < len(Clients); i++ {
					encoder := json.NewEncoder(Clients[i].Connection)
					encoder.Encode(Event)
				}
			}
		} else {
			encoder := json.NewEncoder(ServerConnection)
			encoder.Encode(Event)
		}

		break
	case SocketIO:
		break
	default:
		log.Fatal("Protocol Type not match")
	}
}

func readFully(conn net.Conn) ([]byte, error) {
	result := bytes.NewBuffer(nil)
	var buf [512]byte
	for {
		n, err := conn.Read(buf[0:])
		result.Write(buf[0:n])
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
	}
	return result.Bytes(), nil
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// RandStringRunes func
// 랜덤 문자열 생성 함수
func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
