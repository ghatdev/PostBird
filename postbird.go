package postbird

import (
	"encoding/json"
	"errors"
	"fmt"
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

// Client struct
// 서버모드에서 연결된 클라이언트를 저장할 구조
// socket.io를 사용할 경우에는 Connection 값이 비고, TCP를 사용할 경우 Socket 값이 빈다.
type Client struct {
	Socket     socketio.Socket
	Connection net.Conn
	ClientID   string
}

// Any struct
// 모든 형식의 값을 다 받기위한 interface
type Any interface{}

// CallEvent struct
// TCP에서 함수 호출시 사용하는 이벤트 구조
type CallEvent struct {
	FunctionName string
	Params       []Any
}

// 따로 set함수들을 호출하지 않으면 밑의 값들을 이용한다
// DefaultPort: 기본으로 연결, 바인딩될 포트
// DefaultBindAddress: 서버모드에서 기본으로 바인딩할 IP
// DefaultRemoteAddress: 클라이언트에서 기본으로 연결을 시도할 서버의 주소
// DefaultProtocol: 사용할 Protocol
const (
	DefaultPort          uint   = 8787
	DefaultBindAddress   string = "127.0.0.1"
	DefaultRemoteAddress string = "127.0.0.1"
	DefaultProtocol      uint   = SocketIO
)

// 서버로 사용하고싶으면 0, 클라이언트로 사용하고싶으면 1
const (
	ServerMode = 0
	ClientMode = 1
)

// TCP를 통해 연결하고 싶으면 0, socket.io로 연결하고 싶으면 1
const (
	TCP      = 0
	SocketIO = 1
)

// 라이브러리에서 사용될 값 저장할 공간
var info Info

// ServerConnection : 클라이언트 모드에서 연결된 서버의 연결객체
var ServerConnection net.Conn

// 서버가 연결되었는지 판단할 bool 값
var isConnected bool

// Clients : 클라이언트가 연결되면 Client 형식으로 저장한다.
var Clients []Client

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
// 연결할 서버의 주소를 설정하는 함수. 호출하지 않으면 DefaultRemoteAddress인 127.0.0.1이 설정된다
func SetRemoteAddress(ServerAddress string) {
	info.RemoteAddress = ServerAddress
}

// SetRemotePort func
// 연결할 서버의 포트를 설정하는 함수. 호출하지 않으면 DefaultPort인 8787이 설정된다
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
// 프로그램을 서버역할로 사용하려면 이 함수를 호출
// 시작되면 Binder 함수를 비동기로 호출하여 비동기로 tcp Listen
// 혹은 socket.io 를 사용할 수 있음
// 이 함수가 호출되면 무조건 Mode가 ServerMode 로 바뀐다
// Protocol로 TCP를 사용할 것인지 socket.io 를 사용할 것인지 정해야 한다 0은 TCP, 1은 socket.io
func StartServer(Protocol uint) {
	var wg sync.WaitGroup // 고루틴을 위한 WaitGroup 생성

	info.Mode = ServerMode // 서버를 시작한 것임으로 무조건 Mode 는 ServerMode
	info.Protocol = Protocol

	switch Protocol {
	case TCP:
		wg.Add(1)
		go Binder(&wg, info.BindAddress, info.BindPort) // TCP를 사용할 경우
	case SocketIO:
		wg.Add(1)
		go Listener(&wg, info.BindAddress, info.BindPort) // socket.io를 사용할 경우
	default:
		log.Println("Protocol not match. 0 for TCP, 1 for Socket.io.")
	}
	wg.Wait()
}

// Listener func
// ServerMode 일때 tcp대신 socket.io 사용
// 구현 덜됨;;
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
// ServerMode에서 TCP를 바인딩하여 요청을 requestHandler로 전달해주는 함수
func Binder(wg *sync.WaitGroup, BindAddr string, Port uint) {
	defer wg.Done()
	info.Protocol = TCP // Binder는 TCP 모드용  함수다

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

		ClientID := RandStringRunes(17)
		Clients = append(Clients, Client{nil, conn, ClientID})
		WaitHandler.Add(1)
		go requestHandler(&WaitHandler, conn) // 비동기로 requestHandler 호출
	}

	WaitHandler.Wait()
}

// requestHandler func
// tcp 연결되었을때 request 핸들러
func requestHandler(wg *sync.WaitGroup, c net.Conn) {
	defer wg.Done()
	data := json.NewDecoder(c)

	var FuncWaiter sync.WaitGroup
	var event CallEvent

	for {
		err := data.Decode(&event)
		if err != nil {
			log.Println("Invalid json format")
			return
		}

		FuncWaiter.Add(1)
		go CallLocalFunc(&FuncWaiter, event.FunctionName, event.Params...) // 비동기로 등록된 함수 실행
	}

	FuncWaiter.Wait()
}

// ConnectToRemote func
// 클라이언트로 사용할경우 서버에 연결할때 사용하는 함수
// 무슨 Protocol로 연결할지 정해야한다
func ConnectToRemote(Protocol uint) {
	info.Mode = ClientMode
	info.Protocol = Protocol

	switch Protocol {
	case TCP:
		client, err := net.Dial("tcp", info.RemoteAddress+":"+fmt.Sprint(info.RemotePort))
		if err != nil {
			fmt.Println(err)
			return
		}

		ServerConnection = client

		break
	case SocketIO:
		break
	default:
		log.Println("Undefined Protocol")
	}

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

	Event := CallEvent{FunctionName, args}

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

		if info.Mode == ServerMode { //서버모드면 다중 클라이언트일 가능성이 있음으로
			if Clients[0].ClientID != "" { //클라이언트가 하나라도 연결되 있으면
				for i = 0; i < len(Clients); i++ { // 모든 클라이언트에
					encoder := json.NewEncoder(Clients[i].Connection)
					encoder.Encode(Event) // 해당 이벤트를 보낸다
				}
			}
		} else { // 서버모드가 아니라면 (클라이언트 모드라면)
			encoder := json.NewEncoder(ServerConnection) // 연결된 서버는 1개임으로 연결된 서버에
			encoder.Encode(Event)                        // 이벤트를 인코딩해서 보낸다 (encoder 가 알아서 json으로 변경해서 보내준다)
		}

		break
	case SocketIO:
		break
	default:
		log.Fatal("Protocol Type not match")
	}
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
