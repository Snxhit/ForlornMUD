package main

import (
	"net"
	"os"
	"fmt"
	"bytes"
	"strings"
)

// networking
const (
	CONN_HOST = "0.0.0.0"
	CONN_PORT = "8899"
	CONN_TYPE = "tcp"
)

// state variables
var ConnectionStore []ConnectionData = []ConnectionData{}

// type structs
type World struct {

}

type ConnectionData struct {
	store net.Conn
	session *Session
}

type Session struct {
	authorized *bool
	id *int
	username *string
	password *string
	character *Character
}

type Character struct {
	hp int
	inventory []string
}

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

		FieldOne := false
		FieldTwo := 0
		FieldThree := ""
		FieldFour := ""
		sesh := Session{&FieldOne, &FieldTwo, &FieldThree, &FieldFour, nil}
		conn := ConnectionData{new_conn, &sesh}
		ConnectionStore = append(ConnectionStore, conn)
		fmt.Println(conn)
		go HandleNewClient(&conn)
	}
}

func HandleNewClient(connection *ConnectionData) {
	stream := connection.store
	stream.Write([]byte("What is your name?\n> "))
	for {
		buf := make([]byte, 2048)
		_, err := stream.Read(buf)
		if err != nil {
			fmt.Println("Error reading input: ", err.Error())
		}

		buf_cmd := bytes.Trim(buf, string([]byte{0}))
		str_cmd := strings.ToLower(string(buf_cmd[0:len(buf_cmd) - 1]))
		cmd_tokens := strings.Split(str_cmd, " ")

		if !*connection.session.authorized {
			print()
			if len(cmd_tokens) != 1 && len(*connection.session.username) == 0 {
				stream.Write([]byte("Names can't have spaces, try again!"))
			} else if len(cmd_tokens) == 1 && len(*connection.session.username) == 0 {
				*connection.session.username = cmd_tokens[0]
				stream.Write([]byte("Is " + cmd_tokens[0] + " your name? Then, enter a password."))
			} else if len(cmd_tokens) != 1 && len(*connection.session.username) != 0 {
				stream.Write([]byte("Passwords can't have spaces, try again!"))
			} else if len(cmd_tokens) == 1 && len(*connection.session.username) != 0 {
				*connection.session.password = cmd_tokens[0]
				*connection.session.authorized = true
				stream.Write([]byte("Welcome to the MUD!"))
				fmt.Println(*connection.session)
			}

			stream.Write([]byte("\n> "))
		} else {
			switch len(cmd_tokens) {
			case 1:
				switch cmd_tokens[0] {
				case "exit":
					stream.Write([]byte("Exiting game."))
					stream.Close()
					return
				default:
					stream.Write([]byte("Command not found!"))
				}

			case 2:
				switch cmd_tokens[0] {
				case "echo":
					stream.Write([]byte(cmd_tokens[1]))
				default:
					stream.Write([]byte("Command not found!"))
				}
			}

			stream.Write([]byte("\n> "))
		}
	}
}
