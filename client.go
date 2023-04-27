package main

import (
	"bufio"
	"flag"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Time     string `json:"time"`
	Username string `json:"username"`
	Message  string `json:"message"`
	Connect  int    `json:"connect"`
}

var broadcast = make(chan Message)
var upgrader = websocket.Upgrader{}
var kbReader = bufio.NewReader(os.Stdin)
var addr = flag.String("addr", "localhost:8000", "http service address")
var connect = make(chan bool)

func main() {
	fplog, err := os.OpenFile("Clientlog.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer fplog.Close()
	log.SetOutput(fplog)
	flag.Parse()

	login()

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	log.Printf("Connection to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial: ", err)
	}

	print("Username : ")
	username, err := kbReader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	username = strings.TrimSpace(username)
	c.WriteMessage(1, []byte(username))

	var msg Message
	msg.Username = username
	msg.Connect = 0

	go input(msg)
	go receivemsg(c)
	sendmsg(c)
}

func login() {
	u := url.URL{Scheme: "login", Host: *addr, Path: "/"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial: ", err)
	}

	defer c.Close()

	for {
		print("ID : ")
		ID, _ := kbReader.ReadString('\n')
		ID = strings.TrimSpace(ID)
		c.WriteMessage(1, []byte(ID))

		print("PW : ")
		PW, _ := kbReader.ReadString('\n')
		PW = strings.TrimSpace(PW)
		c.WriteMessage(1, []byte(PW))

		_, success, _ := c.ReadMessage()
		if string(success) == "G" {
			println("Login Successed")
			break
		} else {
			println("Login Failed")
		}
	}
}

func sendmsg(ws *websocket.Conn) {
	for {
		msg := <-broadcast
		if strings.Compare(msg.Message, `/exit`) == 0 {
			msg.Connect = -1
		}
		err := ws.WriteJSON(msg)
		if err != nil {
			log.Printf("error: %v", err)
		}
		log.Printf("S %s %s", msg.Username, msg.Message)

		if msg.Connect == -1 {
			os.Exit(0)
		}
	}
}

func input(msg Message) {
	for {
		input, err := kbReader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		msg.Message = strings.TrimSpace(input)
		msg.Time = time.Now().String()

		broadcast <- msg
	}
}

func receivemsg(ws *websocket.Conn) {
	for {
		var msg Message
		err := ws.ReadJSON(&msg)

		if err != nil {
			log.Fatal(err)
		}
		log.Printf("R %s %s", msg.Username, msg.Message)
		println(msg.Time, " ", msg.Username, " : ", msg.Message)
	}

}
