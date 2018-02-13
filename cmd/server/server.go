package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/alfreddobradi/rumour-mill/internal/avro"
	"github.com/alfreddobradi/rumour-mill/internal/types"
)

var host = flag.String("host", "127.0.0.1", "Host to listen on")
var port = flag.String("port", "9001", "Port to listen on")

func main() {

	flag.Parse()

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", *host, *port))
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	defer listener.Close()

	clients := make(map[string]net.Conn, 1)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		ip := conn.RemoteAddr()
		clients[ip.String()] = conn
		log.Printf("Client connection from %s", ip.String())

		go handleRequest(conn, clients)
	}

}

func handleRequest(conn net.Conn, clients map[string]net.Conn) {
	message := make([]byte, 1024)
	self := conn.RemoteAddr().String()

	for {
		_, err := conn.Read(message)
		if err != nil {
			if err == io.EOF {
				log.Printf("Client disconnected\n")
				break
			}
			log.Printf("Error: %v\n", err)
			continue
		}

		var m types.Message
		m, err = avro.Decode(message)
		if err != nil {
			log.Printf("Avro error: %v\n", err)
			continue
		}

		conn.Write([]byte("OK"))
		log.Printf("Message received: %v %T\n", m, m)

		for a, c := range clients {
			if a != self {
				fmt.Fprintf(c, "%s said: %s", m.User, m.Message)
			}
		}

	}
}
