package main

import (
	"bufio"
	"fmt"
	"net"
	"os"

	"github.com/danielrs/corkscrew/command"
	"github.com/danielrs/corkscrew/response"
)

func main() {
	// reader := strings.NewReader("*2\r\n$3\r\none\r\n:2\r\n")
	// token, err := response.Lex(reader)
	// fmt.Println(token, err)
	// return
	args := os.Args[1:]

	var addr = "127.0.0.1"
	var port = "6679"

	if len(args) >= 2 {
		addr = args[0]
		port = args[1]
	}

	fullAddress := fmt.Sprintf("%v:%v", addr, port)
	fmt.Printf("Connecting to %q...\n", fullAddress)
	server, err := net.Dial("tcp", fullAddress)
	if err == nil {
		loop(server)
	} else {
		fmt.Println(err)
	}
}

func loop(server net.Conn) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("Command> ")
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		// Sends message.
		line := scanner.Text()
		command := command.Serialize(line)

		if command.Len() > 0 {
			server.Write(command.Bytes())
			res, _ := response.Lex(server)
			fmt.Println(res)
		}
	}
}
