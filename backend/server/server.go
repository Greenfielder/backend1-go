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
	go servMsg()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}

// func servMsg - функция для отправки сообщений клиентам
func servMsg() {
	for {
		reader := bufio.NewReader(os.Stdin)
		srvmsg, err := reader.ReadString('\n')
		if err != nil {
			log.Println(err)
		}
		messages <- srvmsg
	}
}

func broadcaster() {
	clients := make(map[client]bool)
	for {

		select {
		case msg := <-messages:
			for cli := range clients {
				fmt.Println(msg)
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
	fmt.Println(who)
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
