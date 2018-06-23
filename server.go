/*
 * Copyright (c) 2018, Ryan Westlund.
 * This code is under the BSD 3-Clause license.
 */

package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

// Messages are the JSON objects used for most communication between clients and the server. In the case of a chat message, Content will be used and Command will be left blank. In the context of a special message, Command will be used and Content will be left blank.
type Message struct {
	Username string `json:"username"`
	Content  string `json:"message"`
	Command  string `json:"command"`
}

// The two channels in this struct are for the player sending commands to the server and for the server sending gamestate updates to the player's computer.
type User struct {
	Name             string
	Ready            bool
	InGame           bool
	BattleInputChan  chan Message
	BattleUpdateChan chan Update
}

// ConnInfo models the communication channel between a user's client and the
// server.
type ConnInfo struct {
	Inbound  chan Message
	Outbound chan interface{}
}

// MessageInfo wraps a Message with a reference to the User that sent it.
type MessageInfo struct {
	Message Message
	User    *User
}

func main() {
	// When new clients arrive, their IO channels will be sent through here.
	var newClients = make(chan ConnInfo)
	go dispatcher(newClients)
	fs := http.FileServer(http.Dir("./"))
	http.Handle("/", fs)
	// handleConnection actually returns an anonymous function that handles connections.
	http.Handle("/ws", handleConnection(newClients))
	port := ":8000"
	log.Println("http server starting on port", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// dispatcher takes a channel to receive new clients on and coordinates
// high-level message passing. It alone has the list of all connected clients,
// so no mutex is needed. Because it only takes in ConnInfos, it doesn't care
// how the clients are connected.
func dispatcher(newClients <-chan ConnInfo) {
	// The list of clients never leaves this scope.
	var clients = make(map[*ConnInfo]*User)
	// All incoming messages will be merged into this channel.
	var messages = make(chan MessageInfo)
	// This is used for clients that disconnect, so they can be removed.
	var leaving = make(chan *ConnInfo)
	for {
		select {
		// When a new connection is established.
		case newConn := <-newClients:
			// Add them to the list.
			user := User{"", false, false, make(chan Message), make(chan Update)}
			clients[&newConn] = &user

			// Merge their Messages ino the single messages channel.
			go func(sink chan<- MessageInfo, conn *ConnInfo, user *User,
				leaving chan<- *ConnInfo) {
				for m := range conn.Inbound {
					// Associate the Message with the User so we can tell who
					// sent it later.
					sink <- MessageInfo{Message: m, User: user}
				}
				// Let dispatch know that they're gone before we exit.
				leaving <- conn
			}(messages, &newConn, &user, leaving)

		// Delete clients when they disconnect.
		case oldConn := <-leaving:
			delete(clients, oldConn)

		// When a Message is received from anyone.
		case msg := <-messages:
			// If they're in a game, forward all messages there.
			if msg.User.InGame {
				if msg.Message.Command == "END MATCH" {
					msg.User.InGame = false
				} else {
					msg.User.BattleInputChan <- msg.Message
				}

				// Handle lobby command messages.
			} else if msg.Message.Command != "" {
				switch msg.Message.Command {
				case "READY":
					msg.User.Ready = true
					// Try to start a match.
					matchmaker(clients)
				case "UNREADY":
					msg.User.Ready = false
				case "SETNAME":
					msg.User.Name = msg.Message.Username
				case "BOT MATCH":
					// Find the ConnInfo in the clients map, because msg.User doesn't contain it.
					var conn *ConnInfo
					for conn = range clients {
						if clients[conn]==msg.User {
							break
						}
					}
					msg.User.Ready = false
					msg.User.InGame = true
					botInputChan:=make(chan Message)
					botUpdateChan:=make(chan Update)
					conn.Outbound <- Message{Username: "", Content: msg.Message.Content, Command: "START GAME"}
					switch msg.Message.Content {
					case "AttackBot":
						go AttackBot(botInputChan,botUpdateChan)
					default:
						log.Println("unrecognized bot",msg.Message.Content)
					}
					go battle(msg.User.BattleInputChan, botInputChan, msg.User.BattleUpdateChan, botUpdateChan)
					go forwardUpdates(conn.Outbound, msg.User.BattleUpdateChan)
				default:
					log.Println("got unexpected command message", msg.Message.Command, "from user", msg.Message.Username)
				}
				// Handle lobby chat messages.
			} else {
				for conn := range clients {
					conn.Outbound <- msg.Message
				}
			}
		}
	}
}

// This function is called whenever a new player readies for battle. If at least two people are ready for battle, it matches two of them.
func matchmaker(clients map[*ConnInfo]*User) {
	readyUsers := make([]*ConnInfo, 0)
	for socket, user := range clients {
		if user.Ready == true {
			readyUsers = append(readyUsers, socket)
		}
	}
	if len(readyUsers) >= 2 {
		clients[readyUsers[0]].Ready = false
		clients[readyUsers[0]].InGame = true
		clients[readyUsers[1]].Ready = false
		clients[readyUsers[1]].InGame = true
		readyUsers[0].Outbound <- Message{Username: "", Content: clients[readyUsers[1]].Name, Command: "START GAME"}
		readyUsers[1].Outbound <- Message{Username: "", Content: clients[readyUsers[0]].Name, Command: "START GAME"}
		go battle(clients[readyUsers[0]].BattleInputChan, clients[readyUsers[1]].BattleInputChan, clients[readyUsers[0]].BattleUpdateChan, clients[readyUsers[1]].BattleUpdateChan)
		go forwardUpdates(readyUsers[0].Outbound, clients[readyUsers[0]].BattleUpdateChan)
		go forwardUpdates(readyUsers[1].Outbound, clients[readyUsers[1]].BattleUpdateChan)
	}
}

// Each time a new user connects, a goroutine running the function that this one returns is created. It keeps track of the connection and sends chat data or game data back and forth.
func handleConnection(newClients chan<- ConnInfo) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Upgrade initial GET request to a websocket
		var upgrader = websocket.Upgrader{}
		socket, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Fatal(err)
		}
		defer socket.Close()
		// Send the connection info.
		var conn = ConnInfo{
			Inbound:  make(chan Message),
			Outbound: make(chan interface{}),
		}
		// This will let the consumer know that it's no longer active.
		defer close(conn.Inbound)
		defer close(conn.Outbound)
		// Signal that a new client has arrived.
		newClients <- conn

		// Connect the outbound channel to the websocket.
		go func() {
			for msg := range conn.Outbound {
				err := socket.WriteJSON(msg)
				if err != nil {
					log.Println(err)
				}
				//TODO remove them or just drop the message?
			}
		}()

		// Connect the websocket to the inbound channel.
		var msg Message
		for {
			// Read the next message from chat
			err := socket.ReadJSON(&msg)
			if err != nil {
				log.Printf("error: %v", err)
				return
			}
			conn.Inbound <- msg
		}
	})
}

// This goroutine listens for gamestate updates from battle.go and forwards them to the player.
func forwardUpdates(dest chan interface{}, src chan Update) {
	for update := range src {
		dest <- update
		// This function is here to prevent crashes when someone reloads during battle
		defer func() {
			// recover from panic caused by writing to a closed channel
			r := recover()
			if r != nil {
				log.Printf("write: error writing on channel: %v\n", r)
				return
			}
		}()
		if update.Self.Life <= 0 || update.Enemy.Life <= 0 {
			return
		}
	}
}
