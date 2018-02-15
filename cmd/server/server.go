package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/alfreddobradi/rumour-mill/internal/avro"
	"github.com/alfreddobradi/rumour-mill/internal/timescale"
	"github.com/alfreddobradi/rumour-mill/internal/types"
)

const URI = "postgresql://postgres@127.0.0.1:5432/tutorial?sslmode=disable"

var host = flag.String("host", "127.0.0.1", "Host to listen on")
var port = flag.String("port", "9001", "Port to listen on")

func main() {

	flag.Parse()

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", *host, *port))
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	timescaleConnection, err := timescale.New(URI)
	if err != nil {
		log.Fatalf("Error connecting to Timescale: %v", err)
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

		go handleRequest(conn, &timescaleConnection, clients)
	}

}

func handleRequest(conn net.Conn, db types.Persister, clients map[string]net.Conn) {
	message := make([]byte, 1024)
	self := conn.RemoteAddr().String()

	for {
		_, err := conn.Read(message)
		if err != nil {
			if err == io.EOF {
				log.Printf("Client %s disconnected\n", self)
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
		m.Time = time.Now().UnixNano()

		db.Persist(m)
		log.Printf("Message received: %v %T\n", m, m)

		for _, c := range clients {
			c.Write(message)
		}
	}
}
