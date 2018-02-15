package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/alfreddobradi/rumour-mill/internal/avro"
	"github.com/alfreddobradi/rumour-mill/internal/types"
)

var nick = flag.String("nick", "", "Your nick")

func main() {

	flag.Parse()

	if len(*nick) == 0 {
		fmt.Println("You have to choose a nick")
		os.Exit(1)
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:9001")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	defer conn.Close()

	chatConn := make(chan []byte)
	in := make(chan string)
	quit := make(chan bool)

	go func() {
		for {
			msg := make([]byte, 1024)
			_, err := conn.Read(msg)
			if err != nil {
				if err == io.EOF {
					log.Fatalf("Server sent EOF. Bye!")
				}
				log.Fatalf("Error: %v", err)
			}

			in <- string(msg)
		}
	}()

	go func() {

		reader := bufio.NewReader(os.Stdin)

		for msg, err := reader.ReadString('\n'); msg != "Q\n"; {

			if err != nil {
				log.Printf("Error reading message: %v\n", err)
				continue
			}

			msg = strings.TrimSpace(msg)

			var m types.Message
			m.User = *nick
			m.Message = msg

			avro, err := avro.Encode(m)
			if err != nil {
				log.Printf("Avro encode failed: %v\n", err)
				quit <- true
			}

			chatConn <- avro

			msg, err = reader.ReadString('\n')
			if err != nil {
				quit <- true
			}
		}

		quit <- true
	}()

	chat(chatConn, in, quit, conn)
}

func chat(chat chan []uint8, in chan string, quit chan bool, conn net.Conn) {
	for {
		select {
		case x := <-chat:
			_, err := conn.Write(x)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}
		case x := <-in:
			if m, err := avro.Decode([]byte(x)); err == nil {
				if m.User == *nick {
					log.Printf("You said: %s", m.Message)
				} else {
					log.Printf("%s said: %s", m.User, m.Message)
				}
			} else {
				log.Printf("Error decoding incoming message: %v", err)
			}
		case <-quit:
			conn.Close()
			return
		}
	}
}
