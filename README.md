[![Build Status](https://travis-ci.org/ghatdev/PostBird.svg?branch=master)](https://travis-ci.org/ghatdev/PostBird)

# PostBird
Easy communication library for Golang

## Usage
### Quick Start
Run this command in bash or cmd or etc.. 
```shell
go get github.com/ghatdev/postbird
```
And, add this line 
```go
import "github.com/ghatdev/PostBird"
``` 
in your code

### How to use	
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
		postbird.CallRemoteFunc("test", "Hello World!")
}
```  

  - Result (Server Prompt):  
    Hello World!
