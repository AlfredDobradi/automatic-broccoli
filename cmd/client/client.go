package main

import (
    "bufio"
    "fmt"
    "io"
    "log"
    "net"
    "os"
)

func main() {
    tcpAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:9001")
    if err != nil {
        log.Fatalf("Error: %v", err)
        os.Exit(1)
    }

    conn, err := net.DialTCP("tcp", nil, tcpAddr)
    if err != nil {
        log.Fatalf("Error: %v", err)
        os.Exit(1)
    }

    defer conn.Close()

    chatConn := make(chan string)
    in := make(chan string)
    quit := make(chan bool)

    go func() {
        msg := make([]byte, 1024)
        _, err := conn.Read(msg)
        if err != nil && err != io.EOF {
            log.Fatalf("Error: %v", err)
        }

        in <- string(msg)
    }()

    go func() {

        reader := bufio.NewReader(os.Stdin)

        msg, err := reader.ReadString('\n')

        if err != nil {
            fmt.Println(err)
            return
        }

        for msg != "Q\n" {
            chatConn <- msg

            msg, err = reader.ReadString('\n')
            if err != nil {
                quit <- true
            }
        }

        quit <- true
    }()

    chat(chatConn, in, quit, conn)
}

func chat(chat, in chan string, quit chan bool, conn net.Conn) {
    for {
        select {
        case x := <-chat:
            _, err := conn.Write([]byte(x))
            if err != nil {
                log.Fatalf("Error: %v", err)
                os.Exit(1)
            }
        case x := <-in:
            log.Printf("Message from the server: %s", x)
        case <-quit:
            return
        }
    }
}