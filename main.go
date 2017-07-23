package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/danielrs/corkscrew/command"
	"github.com/danielrs/corkscrew/response"
)

func main() {
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

		// Reads message from user.
		line := scanner.Text()
		command := command.NewCommand(line)
		serialized := command.Serialize()
		if serialized.Len() <= 0 {
			continue
		}

		// Sends message to the server.
		server.Write(serialized.Bytes())
		if res, err := response.Lex(server); err == nil {
			fmt.Println(res)
			if command.IsQuit() && len(res) > 0 && res[0].IsOk() {
				return
			}
		} else if err == io.EOF {
			fmt.Println("Lost connection to server...")
		} else {
			fmt.Println(err)
		}
	}
}
