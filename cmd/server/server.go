package main

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/alfreddobradi/rumour-mill/internal/avro"
	"github.com/alfreddobradi/rumour-mill/internal/config"
	"github.com/alfreddobradi/rumour-mill/internal/logger"
	"github.com/alfreddobradi/rumour-mill/internal/message"
	"github.com/alfreddobradi/rumour-mill/internal/stdout"
	"github.com/alfreddobradi/rumour-mill/internal/timescale"
	"github.com/alfreddobradi/rumour-mill/internal/types"
)

func getConnection(cfg config.Options) (types.Persister, error) {
	if cfg.Backend == "timescale" {
		conn, err := timescale.New(cfg.Timescale.URI)
		return &conn, err
	}

	conn, err := stdout.New("")
	return &conn, err
}

type client struct {
	Conn *tls.Conn
	Nick string
}

var log = logger.New()

func main() {

	options, err := config.Load("../../config.sample.json")
	if err != nil {
		log.Panicf("server: config: %+v", err)
	}

	absPath := "../.."
	cert, err := tls.LoadX509KeyPair(fmt.Sprintf("%s/%s", absPath, options.TLS.Cert), fmt.Sprintf("%s/%s", absPath, options.TLS.Key))
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
		os.Exit(1)
	}

	tlsCfg := tls.Config{Certificates: []tls.Certificate{cert}, ClientAuth: tls.RequireAnyClientCert}
	tlsCfg.Rand = rand.Reader

	listener, err := tls.Listen("tcp", fmt.Sprintf("%s:%d", options.Host, options.Port), &tlsCfg)
	if err != nil {
		log.Fatalf("server: listen: %v", err)
	}

	log.Debugf("%s:%d", options.Host, options.Port)

	backend, err := getConnection(options)
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
		if !ok {
			log.Panicf("server: tls: can't assert connection variable")
		}

		handshake := make([]byte, 1024)
		_, err = tlsc.Read(handshake)
		if err != nil {
			log.Warningf("server: handshake: %s - %+v", ip, err)
			continue
		}

		state := tlsc.ConnectionState()
		for _, v := range state.PeerCertificates {
			if k, err := x509.MarshalPKIXPublicKey(v.PublicKey); err == nil {
				log.Debugf("%s", k)
			} else {
				log.Warningf("%+v", err)
			}
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

			tlsc.Write(ma)
			tlsc.Close()
			continue
		}

		clients[ip] = client{
			tlsc,
			h.User,
		}

		log.Debugf("server: handshake: %s - %s", ip, h.User)

		go handleRequest(tlsc, backend, clients)
	}

}

func handleRequest(conn net.Conn, db types.Persister, clients map[string]client) {
	msg := make([]byte, 1024)
	self := conn.RemoteAddr().String()
	client := clients[self]

	for {
		_, err := conn.Read(msg)
		if err != nil {
			if err == io.EOF {
				log.Infof("server: disconnect: [%s] %s", self, client.Nick)
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
			if len(m.Recipient) > 0 && m.Recipient == c.Nick || len(m.Recipient) == 0 || m.User == c.Nick {
				c.Conn.Write(msg)
			}
		}
	}
}
