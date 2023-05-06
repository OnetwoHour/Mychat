package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

type LoginInfo struct {
	ID       string `json:"id"`
	PW       string `json:"pw"`
	USERNAME string `json:"username"`
}

type Message struct {
	ID  string `json:"id"`
	MSG string `json:"msg"`
}

var addr = flag.String("addr", "server:3000", "http service address")
var kbReader = bufio.NewReader(os.Stdin)
var loginInfo LoginInfo

func main() {
	println("ID/PW : root/1234, guest/1234")

	login()
	channelConnect()
	ws := inputChannel()
	defer ws.Close()

	go receivemsg(ws)
	input(ws)
}

func connectws(path string) *websocket.Conn {
	u := url.URL{Scheme: "ws", Host: *addr, Path: path}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal(err)
	}
	return c
}

func login() {
	ws := connectws("/login")
	defer ws.Close()

login:
	fmt.Print("ID : ")
	fmt.Scan(&loginInfo.ID)
	fmt.Print("PW : ")
	fmt.Scan(&loginInfo.PW)

	err := ws.WriteJSON(loginInfo)
	if err != nil {
		log.Fatal(err)
	}

	_, res, _ := ws.ReadMessage()
	if string(res) == "B" {
		println("Wrong ID/PW!")
		goto login
	} else if string(res) == "D" {
		println("Already logined ID!")
		goto login
	}

	print("Login Successed!\nType your instant username to use : ")
	fmt.Scan(&loginInfo.USERNAME)
	println(loginInfo.USERNAME)
	err = ws.WriteJSON(loginInfo)
	if err != nil {
		log.Fatal(err)
	}
}

func channelConnect() {
	ws := connectws("/channel")
	defer ws.Close()

	println("\n==========CHANNEL LIST==========")
	println(" CHANNEL NAME : NUMBER OF USERS")
	println("================================")
	receivemsg(ws)
	println("================================")
}

func input(w *websocket.Conn) {
	var msg Message
	msg.ID = loginInfo.ID

	for {
		input, err := kbReader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		input = strings.TrimSpace(input)

		if string(input) == "/exit" {
			break
		}

		msg.MSG = input
		w.WriteJSON(msg)
	}
}

func receivemsg(w *websocket.Conn) {
	for {
		_, msg, err := w.ReadMessage()
		if err != nil {
			log.Fatal(err)
		}

		if string(msg) == "/exit" {
			break
		}
		println(string(msg))
	}
}

func inputChannel() *websocket.Conn {
	var channel string

	print("Channel to join : ")
	fmt.Scan(&channel)

	ws := joinChannel(channel)

	return ws
}

func joinChannel(channalName string) *websocket.Conn {

	ws := connectws("/channel/" + channalName)

	_, msg, err := ws.ReadMessage()
	if err != nil {
		log.Println(err)
	}

	println(string(msg))

	return ws
}
