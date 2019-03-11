package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
)

// канал + перчень комнат и имен в них
type user struct {
	channel    chan string
	nameInRoom map[string]string
}

type server struct {
	mu          sync.Mutex
	Address     string
	maxMsgSize  int
	maxMsgsNum  int
	control     chan string
	active      bool
	users       map[string]user              // ключ - conn.RemoteAddr().String()
	messages    map[string][]string          //ключ - комната
	subscribers map[string]map[string]string // ключ - комната, значение - map: ключ - conn.RemoteAddr().String(), значение - ник в этой комнате
}

func (srv *server) Run() {
	listner, err := net.Listen("tcp", srv.Address)
	if err != nil {
		panic(err)
	}
	srv.active = true //чтобы убить вызваные сервером горутины по комманде shutdown в случе, если на этом не заканчивется горутина, в которой крутился сервер
	go srv.acceptConnections(&listner)
	go srv.checkSTDIN()

	for cmd := range srv.control {
		if cmd == "shutdown" {
			fmt.Println("server inactive")
			srv.active = false
			break
		}
		// можно добавить управление сервером, например, добавление комнат или бан пользователей
	}
}

func (srv *server) Stop() {
	srv.control <- "shutdown"
}

func (srv *server) checkSTDIN() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		srv.control <- scanner.Text()
	}
}

func (srv *server) initUser(conn *net.Conn, name string) {
	myChan := make(chan string)
	srv.mu.Lock()
	srv.users[name] = user{
		channel:    myChan,
		nameInRoom: make(map[string]string),
	}
	srv.mu.Unlock()
	go srv.getMessages(conn, myChan)
	(*conn).Write([]byte("Welcome!\n"))
}

func (srv *server) cleenUserData(name string) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	close(srv.users[name].channel)
	for key := range srv.users[name].nameInRoom { //комнаты, на которые подписан user
		delete(srv.subscribers[key], name)
	}
	delete(srv.users, name)
}

//return client warning message, nick in room and condition
func (srv *server) verifyPublish(room, msg, name string) (string, string, bool) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if len(msg) > srv.maxMsgSize {
		return fmt.Sprintf("Sorry, but your message is too long (%d bytes). Maximum size is %d bytes\n", len(msg), srv.maxMsgSize), "", false
	}
	if roomMap, ok := srv.subscribers[room]; ok {
		if nick, ok := roomMap[name]; ok {
			return "", nick, true
		}
		return fmt.Sprintf("You should subscribe room %s before sending messages\n", room), "", false
	}
	return fmt.Sprintf("Room with name %s doesn`t exist\n", room), "", false
}

func (srv *server) addMsg(room, msg, nick string) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if len(srv.messages[room]) > srv.maxMsgsNum {
		srv.messages[room] = srv.messages[room][1:]
	}
	srv.messages[room] = append(srv.messages[room], nick+": "+msg)
}

func (srv *server) sendOutMessage(room, msg, name, nick string) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	for key := range srv.subscribers[room] {
		if key != name {
			srv.users[key].channel <- "New message in room " + room + " by " + nick + ": " + msg + "\n"
		}
	}
}

// returns client warning and condition
func (srv *server) registerInRoom(room, nick, name string) (string, bool) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if roomMap, ok := srv.subscribers[room]; ok {
		if _, ok := roomMap[name]; ok {
			return fmt.Sprintf("You are already subscribed on this room\n"), false
		}
		for _, usedNick := range srv.subscribers[room] {
			if usedNick == nick {
				return fmt.Sprintf("Nickname %s is already used in this room\n", usedNick), false
			}
		}
		srv.subscribers[room][name] = nick
		srv.users[name].nameInRoom[room] = nick
		return "", true
	}
	return fmt.Sprintf("Room with name %s doesn`t exist\n", room), false
}

func (srv *server) sendRoomHistory(conn *net.Conn, room string) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	for _, m := range srv.messages[room] {
		(*conn).Write([]byte(m + "\n"))
	}
}

func (srv *server) acceptConnections(listner *net.Listener) {
	defer (*listner).Close()
	for {
		if srv.active == false {
			break
		}
		conn, err := (*listner).Accept()
		if err != nil {
			fmt.Printf("Connection error: %s\n", err.Error())
		} else {
			go srv.handleConnection(&conn)
			fmt.Printf("%s connected\n", conn.RemoteAddr().String())
		}
	}
}

func (srv *server) handleConnection(conn *net.Conn) {
	name := (*conn).RemoteAddr().String()
	defer func() {
		fmt.Println(name, "disconnected")
		(*conn).Close()
	}()
	srv.initUser(conn, name)

	publPattern := regexp.MustCompile("^publish ([0-9A-Za-z_]+)[\t\n\r\f ]*:[\t\n\r\f ]*(.*)$") // \s и \w не работает?
	subPattern := regexp.MustCompile("^subscribe ([0-9A-Za-z_]+)[\t\n\r\f ]*:[\t\n\r\f ]*([0-9A-Za-z_]+)$")
	scanner := bufio.NewScanner(*conn)
	for scanner.Scan() {
		if srv.active == false {
			break
		}
		cmd := scanner.Text()
		fmt.Println("Recieved", cmd)
		if cmd == "exit" {
			srv.cleenUserData(name)
			(*conn).Write([]byte("Bye\n"))
			break
		} else if publPattern.MatchString(cmd) {
			params := publPattern.FindStringSubmatch(cmd)
			if warn, nick, ok := srv.verifyPublish(params[1], params[2], name); ok {
				srv.addMsg(params[1], params[2], nick)
				srv.sendOutMessage(params[1], params[2], name, nick)
				(*conn).Write([]byte(fmt.Sprintf("You successfully published the message\n")))
			} else {
				(*conn).Write([]byte(warn))
			}
		} else if subPattern.MatchString(cmd) {
			params := subPattern.FindStringSubmatch(cmd)
			if warn, ok := srv.registerInRoom(params[1], params[2], name); ok {
				(*conn).Write([]byte(fmt.Sprintf("You successfully connected to the room %s. Message history:\n", params[1])))
				srv.sendRoomHistory(conn, params[1])
			} else {
				(*conn).Write([]byte(warn))
			}
		} else {
			(*conn).Write([]byte("Unknown command\n"))
		}
	}
}

func (srv *server) getMessages(conn *net.Conn, ch chan string) {
	for msg := range ch {
		if srv.active == false {
			break
		}
		if strings.HasPrefix(msg, "New message") {
			(*conn).Write([]byte(msg))
		} else {
			//обработка каких-то команд из main(), например, принудительное отключение пользователя
		}
	}
}
