# PostBird
Easy communication library for Golang

## Usage
### Quick Start
run 
```shell
go get github.com/ghatdev/postbird
```
Add 
```go
import "github.com/ghatdev/PostBird"
``` 
in your code
	
- Server Mode:  
  - Call RegisterFunc() to register functions.  
  - Call StartServer() to start the server.  
  
- Client Mode:   
  - Call ConnectToRemote() to connect to server.  
  - Call CallRemoteFunc() to call registered functions.  
  - call RegisterFunc() to register functions..  
  
## Example
  - As Server :
  ```go
  package main

import (
	"PostBird"
	"fmt"
)

func main() {
	postbird.RegisterFunc("test", test)
	postbird.SetBindAddress("0.0.0.0")
	postbird.StartServer(0)

}

func test(a string) {
	fmt.Println(a)
}
```  
  - As Client :  
  ```go
  package main

import "PostBird"

func main() {
	postbird.ConnectToRemote(0)
	postbird.CallRemoteFunc("test", "abcd")
}
```  

  - Result (Server Prompt):  
    abcd
