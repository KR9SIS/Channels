package main

import (
	"crypto/tls"
	"log"
	"net"

	"github.com/joho/godotenv"
)

type ConnMsg struct {
	sender net.Conn
	msg    []byte
}

func broadcast(ch <-chan ConnMsg) {
	clients := make([]net.Conn, 0, 10)
	for {
		msg := <-ch
		found := false
		for _, client := range clients {
			if client == msg.sender {
				found = true
				continue
			}
			client.Write(msg.msg)
		}
		if !found {
			clients = append(clients, msg.sender)
		}
	}
}

func serve(conn net.Conn, ch chan<- ConnMsg) {
	defer conn.Close()

	const SDPMsgSize = 1024
	buf := make([]byte, SDPMsgSize)
	for {
		offset := 0
		for offset < SDPMsgSize {
			n, err := conn.Read(buf)
			if err != nil {
				log.Printf("Error reading: %#v\n", err)
				return
			}
			offset += n
		}
		msgBuf := make([]byte, SDPMsgSize)
		copy(msgBuf, buf)
		ch <- ConnMsg{
			sender: conn,
			msg:    msgBuf,
		}
	}
}

func main() {
	var env map[string]string
	env, err := godotenv.Read()
	if err != nil {
		log.Fatal(err)
	}

	// https://go-cookbook.com/snippets/networking/ssl-and-tls
	cert, err := tls.LoadX509KeyPair(env["server_cert"], env["server_key"])
	if err != nil {
		log.Fatal(err)
	}
	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	listener, err := tls.Listen("tcp", env["public_ip"]+":443", config)
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	log.Printf("Listening on %s\n", env["public_ip"])

	connChan := make(chan ConnMsg, 50)
	go broadcast(connChan)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %s\nError: %#v\n", conn.LocalAddr().String(), err)
		}
		go serve(conn, connChan)
	}
}
