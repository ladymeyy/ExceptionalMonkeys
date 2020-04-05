package main

import (
	_ "encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Define our message object
type actionMessage struct {
	X string // "0" || "10" ||  "-10"
	Y string // "0" || "10" ||  "-10"
}

type screenWH struct {
	Width int64 // "0" || "10" ||  "-10"
	Height int64 // "0" || "10" ||  "-10"
}

type point struct {
	X int64 `json:"x"`
	Y int64 `json:"y"`
}

type Player struct {
	Id            uuid.UUID `json"id"`
	ExceptionType string    `json:"exceptionType"`
	Color         [3]int    `json:"color"`
	P             point     `json:"p"`
	Show          bool      `json:"show"`
	windowH 	 int64
	windowW      int64
	Collision   string      `json:"collision"`  // border || player || exception
	Score       int   `json:"score"`
}

type Exception struct {
	Id            uuid.UUID `json"id"`
	ExceptionType string    `json:"exceptionType"`
	Show          bool      `json:"show"`
	P             point     `json:"p"`
}

var exceptionsMap = struct {
	sync.RWMutex
	m map[point]Exception
}{m: make(map[point]Exception)}

var s = rand.NewSource(time.Now().UnixNano())
var exceptionsTypes = [3]string{"NullPointerException", "DivideByZeroException", "IOException"}
var clients = make(map[*websocket.Conn]*Player) // connected clients
var broadcastPlayers = make(chan Player)
var broadcastException = make(chan Exception)
var upgrader = websocket.Upgrader{}

func handleNewPlayer(ws *websocket.Conn) {
	r2 := rand.New(s)
	player := Player{Id: uuid.New(), P: point{X: int64(r2.Intn(300)), Y: 0}, Score: 0, Show: true, ExceptionType: exceptionsTypes[rand.Intn(3)], Color: [3]int{r2.Intn(256), r2.Intn(256), r2.Intn(256)}, Collision: "" }

	fmt.Println("new player")
	fmt.Println(player)
	ws.WriteJSON(player)

	//send to new player all current players
	for key := range clients {
		err := ws.WriteJSON(*clients[key])
		if err != nil {
			log.Printf( "76 error: %v", err)
			ws.Close()
			delete(clients, ws)
		}
	}

	//TODO send to new player all current exceptions
	clients[ws] = &player
	broadcastPlayers <- player //broadcast new player

	var msg screenWH
	err := ws.ReadJSON(&msg)
	if err != nil {
		log.Printf(" 89 error: %v", err)
		var plyrMsg = clients[ws]
		plyrMsg.Show = false
		broadcastPlayers <- *plyrMsg
		delete(clients, ws)
	}
	clients[ws].windowH = msg.Height
	clients[ws].windowW = msg.Width

}

func handlePlayerMovement(ws *websocket.Conn, newX int64, newY int64){

	x := int64(clients[ws].P.X) + newX
	y := int64(clients[ws].P.Y) + newY
	player := *clients[ws]

	if y < 0  || x < 0 || x >= clients[ws].windowW || y >= clients[ws].windowH{
		player.Collision = "border"
	}else{

		//check for collision with other players
		for key := range clients {
			client := *clients[key]
			if client.P.X == x ||  client.P.Y == y{
				player.Collision = "player"
			}
		}

		//TODO add check collisions with exceptions
		player.P.X = x
		clients[ws].P.X = x
		clients[ws].P.Y = y
		player.P.Y = y
		fmt.Println("player with new values ")
		fmt.Println(*clients[ws])
	}

	broadcastPlayers <- player
}


func handleConnections(w http.ResponseWriter, r *http.Request) {

	upgrader.CheckOrigin = func(r *http.Request) bool { return true } //no cors

	ws, err := upgrader.Upgrade(w, r, nil) // Upgrade initial GET request to a websocket
	if err != nil {
		log.Fatal(err)
	}

	handleNewPlayer(ws)

	defer ws.Close() // Make sure we close the connection when the function returns

	for {
		var msg actionMessage
		// Read in a new message as JSON and map it to a incomingMessage object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf(" 149 error: %v", err)
			var plyrMsg = clients[ws]
			plyrMsg.Show = false
			broadcastPlayers <- *plyrMsg
			delete(clients, ws)
			break
		}
		//INSERT HERE  call to check collision function
		newX, _ := strconv.ParseInt(msg.X, 10, 64)
		newY, _ := strconv.ParseInt(msg.Y, 10, 64)
		handlePlayerMovement(ws, newX , newY)
	}
}
func broadcastMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcastPlayers

		// Send it out to every client that is currently connected
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("171 error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}

	}
}

func exceptiosMapHandler() {
	//Thread.
	// every 5 sec:  create new ex ->add to eXarr -> broadcast to users
	//exery 10 sec:  choose rand ex ->remove from exArr ->broadcast to users

	var r = rand.New(s)
	addExTicker := time.NewTicker(5 * time.Second)
	go func() {
		for t := range addExTicker.C {
			_ = t // we don't print the ticker time, so assign this `t` variable to underscore `_` to avoid error
			var newEx = Exception{Id: uuid.New(), ExceptionType: exceptionsTypes[r.Intn(3)], P: point{X: int64(r.Intn(255)), Y: int64(r.Intn(255))}, Show: true}
			exceptionsMap.Lock() //take the write lock
			exceptionsMap.m[newEx.P] = newEx
			exceptionsMap.Unlock() // release the write lock
			fmt.Println(newEx)
			broadcastException <- newEx
			//todo add broadcast to user here
		}
	}()

	removeExTicker := time.NewTicker(10 * time.Second)
	go func() {
		for t := range removeExTicker.C {
			_ = t // we don't print the ticker time, so assign this `t` variable to underscore `_` to avoid error
			var value Exception
			exceptionsMap.RLock()
			for key := range exceptionsMap.m {
				value = exceptionsMap.m[key]
				exceptionsMap.RUnlock()
				break
			}
			value.Show = false
			broadcastException <- value
			//exceptionsMap.Lock() //take the write lock
			//delete(exceptionsMap, value.p)
			//exceptionsMap.Unlock() //take the write lock
			fmt.Println(value)

		}
	}()

	// wait for 10 seconds
	//time.Sleep(20 *time.Second)
	//	ticker.Stop()

}

func main() {
	fmt.Println("websockets project")
	// Configure websocket route

	http.HandleFunc("/ws", handleConnections)
	go broadcastMessages()
	//go exceptiosMapHandler()
	// Start the server on localhost port 8080 and log any errors
	log.Println("http server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
