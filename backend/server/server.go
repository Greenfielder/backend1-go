package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

type client chan<- string

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string)
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}

	go broadcaster()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}

		// input := bufio.NewScanner(conn)
		// messages <- input.Text()

		go handleConn(conn)
		// go servMsg(conn)
		go servMsg()

	}
}

// func servMsg(conn net.Conn) {
// 	reader := bufio.NewReader(os.Stdin)
// 	msg, err := reader.ReadString('\n')
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	messages <- msg
// 	conn.Close()
// }

// func servMsg(conn net.Conn) {
// 	input := bufio.NewScanner(os.Stdin)
// 	messages <- input.Text()
// 	conn.Close()
// }

func servMsg() {
	clients := make(map[client]bool)
	reader := bufio.NewReader(os.Stdin)
	premsg, err := reader.ReadString('\n')
	if err != nil {
		log.Println(err)
	}
	messages <- premsg

	for {
		select {
		case msg := <-messages:
			for cli := range clients {
				cli <- msg
			}
		}
	}
}

func broadcaster() {
	clients := make(map[client]bool)
	for {

		select {
		case msg := <-messages:
			for cli := range clients {
				cli <- msg
			}

		case cli := <-entering:
			clients[cli] = true

		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}

func handleConn(conn net.Conn) {
	ch := make(chan string)
	go clientWriter(conn, ch)

	who := conn.RemoteAddr().String()
	ch <- "You are " + who
	messages <- who + " has arrived"
	entering <- ch

	input := bufio.NewScanner(conn)
	for input.Scan() {
		messages <- who + ": " + input.Text()
	}

	leaving <- ch
	messages <- who + " has left"
	conn.Close()
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}
