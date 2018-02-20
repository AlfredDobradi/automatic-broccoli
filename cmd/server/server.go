package main

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/alfreddobradi/rumour-mill/internal/avro"
	"github.com/alfreddobradi/rumour-mill/internal/logger"
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

type client struct {
	Conn net.Conn
	Nick string
}

var log = logger.New()

func main() {

	flag.Parse()

	cert, err := tls.LoadX509KeyPair("../../certs/server.pem", "../../certs/server.key")
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
		os.Exit(1)
	}

	config := tls.Config{Certificates: []tls.Certificate{cert}}
	config.Rand = rand.Reader

	listener, err := tls.Listen("tcp", fmt.Sprintf("%s:%s", *host, *port), &config)
	if err != nil {
		log.Fatalf("server: listen: %v", err)
	}

	backend, err := getConnection()
	if err != nil {
		log.Fatalf("server: backend: %v", err)
	}

	defer listener.Close()

	clients := make(map[string]client, 1)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Warningf("server: accept: %v", err)
		}

		ip := conn.RemoteAddr().String()
		log.Infof("server: connection: %s", ip)

		tlsc, ok := conn.(*tls.Conn)
		if ok {
			state := tlsc.ConnectionState()
			for _, v := range state.PeerCertificates {
				if k, err := x509.MarshalPKIXPublicKey(v.PublicKey); err == nil {
					log.Debugf("%s", k)
				}
			}
		}

		handshake := make([]byte, 1024)
		_, err = conn.Read(handshake)
		if err != nil {
			log.Warningf("server: handshake: %s - %+v", ip, err)
			continue
		}

		h, err := avro.Decode(handshake)
		if err != nil {
			log.Warningf("server: handshake: %s - %+v", ip, err)
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
			log.Warningf("User %s already exists", h.User)
			m := message.Message{
				Type:    "system",
				Message: fmt.Sprintf("Nick %s already connected", h.User),
			}
			ma, err := avro.Encode(m)
			if err != nil {
				log.Errorf("Avro decode error:\n%+v", err)
				continue
			}

			conn.Write(ma)
			conn.Close()
			continue
		}

		clients[ip] = client{
			conn,
			h.User,
		}

		log.Debugf("server: handshake: %s - %s", ip, h.User)

		go handleRequest(conn, backend, clients)
	}

}

func handleRequest(conn net.Conn, db types.Persister, clients map[string]client) {
	msg := make([]byte, 1024)
	self := conn.RemoteAddr().String()

	for {
		_, err := conn.Read(msg)
		if err != nil {
			if err == io.EOF {
				log.Infof("server: disconnect: %s", self)
				delete(clients, self)
				break
			}
			log.Errorf("server: message: %v\n", err)
			continue
		}

		var m message.Message
		m, err = avro.Decode(msg)
		if err != nil {
			log.Errorf("server: avro: %v\n", err)
			continue
		}

		if clients[self].Nick != m.User {
			log.Warningf("server: invalid message: %s - %s != %s", self, clients[self].Nick, m.User)
			continue
		}

		m.Time = time.Now().UnixNano()

		db.Persist(&m)

		for _, c := range clients {
			c.Conn.Write(msg)
		}
	}
}
