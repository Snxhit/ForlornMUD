package main

import (
	"net"
	"os"
	"fmt"
	"bytes"
	"strings"
)

const (
	CONN_HOST = "0.0.0.0"
	CONN_PORT = "8899"
	CONN_TYPE = "tcp"
)

func main() {
	l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		fmt.Println("Error listening: ", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	fmt.Println("Listening successfuly on: " + CONN_HOST + ":" + CONN_PORT)

	for {
		new_conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		new_conn.Write([]byte("> "))
		go HandleNewClient(new_conn)
	}
}

func HandleNewClient(conn net.Conn) {
	for {
		buf := make([]byte, 1024)
		_, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading input: ", err.Error())
		}

		buf_cmd := bytes.Trim(buf, string([]byte{0}))
		str_cmd := strings.ToLower(string(buf_cmd[0:len(buf_cmd) - 1]))
		cmd_tokens := strings.Split(str_cmd, " ")

		switch len(cmd_tokens) {
		case 1:
			switch cmd_tokens[0] {
			case "exit":
				conn.Write([]byte("Exiting game.\n"))
				conn.Close()
				return
			default:
				conn.Write([]byte("Command not found!"))
			}

		case 2:
			switch cmd_tokens[0] {
			case "echo":
				conn.Write([]byte(cmd_tokens[1]))
			}
		default:
			conn.Write([]byte("Command not found!"))
		}

		conn.Write([]byte("\n> "))

	}
	conn.Close()
}
