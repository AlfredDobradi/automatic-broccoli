package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/alfreddobradi/rumour-mill/internal/avro"
	"github.com/alfreddobradi/rumour-mill/internal/message"
	"github.com/alfreddobradi/rumour-mill/internal/stdout"
	"github.com/alfreddobradi/rumour-mill/internal/timescale"
	"github.com/alfreddobradi/rumour-mill/internal/types"
)

const backendType = "stdout"
const uri = "postgresql://postgres@127.0.0.1:5432/tutorial?sslmode=disable"

var host = flag.String("host", "127.0.0.1", "Host to listen on")
var port = flag.String("port", "9001", "Port to listen on")

func getConnection() (types.Persister, error) {
	if backendType == "timescale" {
		conn, err := timescale.New(uri)
		return &conn, err
	}

	conn, err := stdout.New("")
	return &conn, err
}

type Client struct {
	Conn net.Conn
	Nick string
}

func main() {

	flag.Parse()

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", *host, *port))
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	backend, err := getConnection()

	if err != nil {
		log.Fatalf("Error connecting to backend: %v", err)
	}

	defer listener.Close()

	clients := make(map[string]Client, 1)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		ip := conn.RemoteAddr().String()
		log.Printf("Client connection from %s", ip)

		handshake := make([]byte, 1024)
		_, err = conn.Read(handshake)
		if err != nil {
			log.Printf("Handshake from %s failed:\n%+v", ip, err)
			continue
		}

		h, err := avro.Decode(handshake)
		if err != nil {
			log.Printf("Failed decoding handshake from %s:\n%+v", ip, err)
			continue
		}

		exists := false
		for _, c := range clients {
			if c.Nick == h.User {
				exists = true
				break
			}
		}

		if exists {
			log.Printf("User %s already exists", h.User)
			m := message.Message{
				Type:    "system",
				Message: fmt.Sprintf("Nick %s already connected", h.User),
			}
			ma, err := avro.Encode(m)
			if err != nil {
				log.Printf("Avro decode error:\n%+v", err)
				continue
			}

			conn.Write(ma)
			conn.Close()
			continue
		}

		clients[ip] = Client{
			conn,
			h.User,
		}

		log.Printf("Handshake from %s with nick %s was successful", ip, h.User)

		go handleRequest(conn, backend, clients)
	}

}

func handleRequest(conn net.Conn, db types.Persister, clients map[string]Client) {
	msg := make([]byte, 1024)
	self := conn.RemoteAddr().String()

	for {
		_, err := conn.Read(msg)
		if err != nil {
			if err == io.EOF {
				log.Printf("Client %s disconnected\n", self)
				break
			}
			log.Printf("Error: %v\n", err)
			continue
		}

		var m message.Message
		m, err = avro.Decode(msg)
		if err != nil {
			log.Printf("Avro error: %v\n", err)
			continue
		}

		if clients[self].Nick != m.User {
			log.Printf("User sent message with different nick. Closing connection.")
			continue
		}

		m.Time = time.Now().UnixNano()

		db.Persist(&m)

		for _, c := range clients {
			c.Conn.Write(msg)
		}
	}
}
