package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]string)
var broadcast = make(chan Message)
var logcast = make(chan Log)
var upgrader = websocket.Upgrader{}
var PORT = 3000

type Message struct {
	Time     string `json:"time"`
	Username string `json:"username"`
	Message  string `json:"message"`
	Connect  int    `json:"connect"`
}

type Log struct {
	Type    string `json:"type"` // E : Error S : Send, L : Link, O : Operation, F : File
	Time    string `json:"time"`
	Host    string `json:"host"`
	Guest   string `json:"guest"`
	Message string `json:"message"`
}

func main() {
	fpLog, err := os.OpenFile("logfile.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer fpLog.Close()

	multiWriter := io.MultiWriter(fpLog, os.Stdout)
	log.SetOutput(multiWriter)

	go http.HandleFunc("/", home)
	http.HandleFunc("/ws", handleConnections)
	go savelog()
	go handleMessages()

	log := []string{`O`, time.Now().String(), `SERVER`, `SERVER`, "http server started on :3000"}
	go getLogs(log...)

	err = http.ListenAndServe(":3000", nil)
	if err != nil {
		log := []string{`E`, time.Now().String(), `SERVER`, `SERVER`, err.Error()}
		go getLogs(log...)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log := []string{`E`, time.Now().String(), `SERVER`, `SERVER`, err.Error()}
		go getLogs(log...)
	}

	defer ws.Close()

	var validLogin = false

	for !validLogin {
		_, ID, err := ws.ReadMessage()
		_, PW, err := ws.ReadMessage()
		if err != nil {
			log := []string{`E`, time.Now().String(), `SERVER`, `SERVER`, err.Error()}
			go getLogs(log...)
		}

		validLogin = requestLogin(string(ID), string(PW))

		if validLogin {
			ws.WriteMessage(1, []byte(`G`))
			continue
		} else {
			ws.WriteMessage(1, []byte(`B`))
		}
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log := []string{`E`, time.Now().String(), `SERVER`, `SERVER`, err.Error()}
		go getLogs(log...)
	}

	defer ws.Close()

	_, name, err := ws.ReadMessage()
	if err != nil {
		log := []string{`E`, time.Now().String(), `SERVER`, `SERVER`, err.Error()}
		go getLogs(log...)
	}

	clients[ws] = string(name)

	log := []string{`L`, time.Now().String(), `SERVER`, clients[ws], "Connected"}
	go getLogs(log...)

	for {
		var msg Message
		err := ws.ReadJSON(&msg)

		log := []string{`S`, msg.Time, msg.Username, `SERVER`, msg.Message}
		go getLogs(log...)

		if err != nil {
			log := []string{`S`, msg.Time, msg.Username, `SERVER`, err.Error()}
			go getLogs(log...)
			delete(clients, ws)
		}
		broadcast <- msg

		if msg.Connect == -1 {
			break
		}
	}
}

func handleMessages() {

	for {
		msg := <-broadcast
		if msg.Connect == -1 {
			msg.Message = msg.Username + " Disconnected"
		}

		for client := range clients {

			if msg.Connect == -1 && clients[client] == msg.Username {
				log := []string{`L`, time.Now().String(), `SERVER`, clients[client], "Disconnected"}
				go getLogs(log...)
				client.Close()
				delete(clients, client)
				continue
			}

			err := client.WriteJSON(msg)
			if err != nil {
				log := []string{`L`, time.Now().String(), `SERVER`, clients[client], err.Error()}
				go getLogs(log...)
				client.Close()
				delete(clients, client)
			}

			log := []string{`S`, time.Now().String(), `SERVER`, clients[client], msg.Message}
			go getLogs(log...)

		}
	}
}

func getLogs(logs ...string) {
	var cast Log
	cast.Type = logs[0]
	cast.Time = logs[1]
	cast.Host = logs[2]
	cast.Guest = logs[3]
	cast.Message = logs[4]

	log.Printf("%s %s %s %s %s", cast.Type, cast.Time, cast.Host, cast.Guest, cast.Message)

	logcast <- cast
}

func savelog() {
	/*Dummy*/
}

func requestLogin(ID string, PW string) bool {
	var url = `http://login:1000`
	buffer := make([]byte, 8)
	var valid int

	connect, err := net.Dial("tcp", url)
	if err != nil {
		log := []string{`E`, time.Now().String(), `SERVER`, `SERVER`, err.Error()}
		go getLogs(log...)
	} else {
		connect.Write([]byte(ID))
		valid, _ = connect.Read(buffer)
		if valid == 0 {
			connect.Write([]byte(PW))
			valid, _ = connect.Read(buffer)
		}
	}

	if valid == 0 {
		return true
	} else {
		return false
	}
}
