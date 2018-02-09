package main

import (
    "flag"
    "fmt"
    "log"
    "net"
    "os"

    "brvy.win/alfreddobradi/chat/internal/avro"
    "brvy.win/alfreddobradi/chat/internal/types"
)

var host = flag.String("host", "127.0.0.1", "Host to listen on")
var port = flag.String("port", "9001", "Port to listen on")

func main() {

    clients := make([]net.Conn, 0)

    flag.Parse()

    listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", *host, *port))
    if err != nil {
        log.Fatalf("Error: %v", err)
        os.Exit(1)
    }

    defer listener.Close()

    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Fatalf("Error: %v", err)
            os.Exit(1)
        }

        clients = append(clients, conn)

        var bc uint8
        for _, client := range clients {
            if client != conn {
                bc++
                client.Write([]byte("New client connected"))
            }
        }

        log.Printf("Broadcasted for %d clients", bc)

        go handleRequest(conn)
    }

}

func handleRequest(conn net.Conn) {
    message := make([]byte, 1024)

    _, err := conn.Read(message)
    if err != nil {
        log.Fatalf("Error: %v", err)
        return
    }

    var m types.Message
    m, err = avro.Decode(message)
    if err != nil {
        log.Printf("Avro error: %v", err)
    }

    conn.Write([]byte("OK"))
    log.Printf("Message received: %v %T\n", m)
    conn.Close()
}
