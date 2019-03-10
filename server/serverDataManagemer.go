package main

import (
	"fmt"
	"net"
	"sync"
)

// канал + перчень комнат и имен в них
type user struct {
	channel    chan string
	nameInRoom map[string]string
}

type serverData struct {
	mu          sync.Mutex
	maxMsgSize  int
	maxMsgsNum  int
	users       map[string]user              // ключ - conn.RemoteAddr().String()
	messages    map[string][]string          //ключ - комната
	subscribers map[string]map[string]string // ключ - комната, значение - map: ключ - conn.RemoteAddr().String(), значение - ник в этой комнате
}

func initServerData(rooms []string) *serverData {
	data := &serverData{
		mu:          sync.Mutex{},
		maxMsgSize:  maxMessageSize,
		maxMsgsNum:  maxMessagesNum,
		users:       make(map[string]user),
		messages:    make(map[string][]string),
		subscribers: make(map[string]map[string]string),
	}

	for _, room := range rooms {
		data.messages[room] = []string{}
		data.subscribers[room] = make(map[string]string)
	}

	return data
}

func (data *serverData) initUser(conn *net.Conn, name string) {
	myChan := make(chan string)
	data.mu.Lock()
	data.users[name] = user{
		channel:    myChan,
		nameInRoom: make(map[string]string),
	}
	data.mu.Unlock()
	go getMessages(conn, myChan, data)
	(*conn).Write([]byte("Welcome!\n"))
}

func (data *serverData) cleenUserData(name string) {
	data.mu.Lock()
	defer data.mu.Unlock()
	close(data.users[name].channel)
	for key := range data.users[name].nameInRoom { //комнаты, на которые подписан user
		delete(data.subscribers[key], name)
	}
	delete(data.users, name)
}

//return client warning message, nick in room and condition
func (data *serverData) verifyPublish(room, msg, name string) (string, string, bool) {
	data.mu.Lock()
	defer data.mu.Unlock()
	if len(msg) > data.maxMsgSize {
		return fmt.Sprintf("Sorry, but your message is too big (%d bytes). Maximum size is %d bytes\n", len(msg), data.maxMsgSize), "", false
	}
	if roomMap, ok := data.subscribers[room]; ok {
		if nick, ok := roomMap[name]; ok {
			return "", nick, true
		} else {
			return fmt.Sprintf("You should subscribe room %s before sending messages\n", room), "", false
		}
	} else {
		return fmt.Sprintf("Room with name %s doesn`t exist\n", room), "", false
	}
}

func (data *serverData) addMsg(room, msg, nick string) {
	data.mu.Lock()
	defer data.mu.Unlock()
	if len(data.messages[room]) > data.maxMsgsNum {
		data.messages[room] = data.messages[room][1:]
	}
	data.messages[room] = append(data.messages[room], nick+": "+msg)
}

func (data *serverData) sendOutMessage(room, msg, name, nick string) {
	data.mu.Lock()
	defer data.mu.Unlock()
	for key := range data.subscribers[room] {
		if key != name {
			data.users[key].channel <- "New message in room " + room + " by " + nick + ":\n" + msg + "\n"
		}
	}
}

// returns client warning and condition
func (data *serverData) registerInRoom(room, nick, name string) (string, bool) {
	data.mu.Lock()
	defer data.mu.Unlock()
	if roomMap, ok := data.subscribers[room]; ok {
		if _, ok := roomMap[name]; ok {
			return fmt.Sprintf("You are already subscribed on this room\n"), false
		} else {
			for _, usedNick := range data.subscribers[room] {
				if usedNick == nick {
					return fmt.Sprintf("Nickname %s is already used in this room\n", usedNick), false
				}
			}
			data.subscribers[room][name] = nick
			return "", true
		}
	} else {
		return fmt.Sprintf("Room with name %s doesn`t exist\n", room), false
	}
}

func (data *serverData) sendRoomHistory(conn *net.Conn, room string) {
	data.mu.Lock()
	defer data.mu.Unlock()
	for _, m := range data.messages[room] {
		(*conn).Write([]byte(m + "\n"))
	}
}
