package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
)

var channelList = make(map[string]int)
var upgrader = websocket.Upgrader{}
var logined = make(map[string]string)
var channelUser = make(map[string]map[*websocket.Conn]bool)
var message = make(map[string](chan Message))

const MAX_USER = 30

type LoginInfo struct {
	ID       string `json:"id"`
	PW       string `json:"pw"`
	USERNAME string `json:"username"`
}

type Message struct {
	MSG string `json:"msg"`
	ID  string `json:"id"`
}

func main() {
	makeChannel("official")
	makeChannel("minor")

	go http.HandleFunc("/", home)
	go http.HandleFunc("/login", login)
	go runChannel()
	go commendInput()

	http.ListenAndServe(":3000", nil)
}

func commendInput() {
	var input string
	for {
		fmt.Scan(&input)
		if input == "/exit" {
			os.Exit(0)
		}
	}

}

func home(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintln(w, "SERVER IS RUNNING!")
}

func login(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	defer ws.Close()

	var loginInfo LoginInfo

Login:
	err = ws.ReadJSON(&loginInfo)
	if err != nil {
		return
	}

	connect := requestLogin(loginInfo)
	if !connect {
		ws.WriteMessage(1, []byte("B"))
		goto Login
	}

	_, exist := logined[loginInfo.ID]
	if exist {
		ws.WriteMessage(1, []byte("D"))
	}

	ws.WriteMessage(1, []byte("G"))
	ws.ReadJSON(&loginInfo)

	logined[loginInfo.ID] = loginInfo.USERNAME
}

func runChannel() {
	go http.HandleFunc("/channel", channelShow)

	for channel := range channelList {
		go http.HandleFunc("/channel/"+channel, channelRoom)
	}
}

func channelShow(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	defer ws.Close()

	for channel := range channelList {
		ws.WriteMessage(1, []byte(channel+" : "+strconv.Itoa(channelList[channel])))
	}
	ws.WriteMessage(1, []byte("/exit"))
}

func channelRoom(w http.ResponseWriter, r *http.Request) {
	var channel = r.URL.Path
	channel = strings.Split(channel, "/")[2]

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	defer ws.Close()

	if channelList[channel] > 29 {
		ws.WriteMessage(1, []byte(channel+" channel is fulled!"))
		return
	}

	channelUser[channel][ws] = true
	channelList[channel]++
	ws.WriteMessage(1, []byte(channel+" channel connected("+strconv.Itoa(channelList[channel])+"/30)"))

	go spreadMsg(ws, channel)
	channelChat(ws, channel)
}

func requestLogin(loginInfo LoginInfo) bool {
	db, err := sql.Open("mysql", "root:1234@tcp(logindb:3306)/logindb")
	if err != nil {
		log.Print(err)
		return false
	}

	defer db.Close()
	var password string

	err = db.QueryRow("SELECT PW FROM LoginInfo WHERE ID = ?", loginInfo.ID).Scan(&password)
	if err != nil {
		if err == sql.ErrNoRows {
			password = ""
		} else {
			fmt.Println(err)
		}
	}

	if password == loginInfo.PW {
		return true
	} else {
		return false
	}

}

func channelChat(ws *websocket.Conn, channel string) {
	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Println(err)
			delete(channelUser[channel], ws)
			delete(logined, msg.ID)
			channelList[channel]--
			break
		}

		if string(msg.MSG) == "" {
			continue
		}

		message[channel] <- msg
		log.Println(msg.ID + " : " + msg.MSG)
	}
}

func spreadMsg(ws *websocket.Conn, channel string) {
	for {
		msg := <-message[channel]
		for client := range channelUser[channel] {
			client.WriteMessage(1, []byte(logined[msg.ID]+" : "+msg.MSG))
		}
	}
}

func makeChannel(name string) {
	channelList[name] = 0
	message[name] = make(chan Message)
	channelUser[name] = make(map[*websocket.Conn]bool)
}
