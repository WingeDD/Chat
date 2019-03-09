package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
)

const (
	configFileName string = "serverConfig.json"
	maxMsgSize     int    = 254
	maxMsgsNum     int    = 128
)

// канал + перчень комнат и имен в них
type user struct {
	channel    chan string
	nameInRoom map[string]string
}

// ключ - conn.RemoteAddr().String()
type users struct {
	set map[string]user
	mu  *sync.Mutex
}

//ключ - комната
//можно было бы использовать массивы размерности 128, но, вероятно, в некоторых комнатах будет мало сообщений и мы займем лишнее место
type messages struct {
	set map[string][]string
	mu  *sync.Mutex
}

// ключ - комната, значение - map: ключ - conn.RemoteAddr().String(), значение - ник в этой комнате
type subscribers struct {
	set map[string]map[string]string
	mu  *sync.Mutex
}

type Config struct {
	Address string   `json:"address"`
	Rooms   []string `json:"rooms"`
}

func getConfig(path string) (*Config, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("in function getRoomsFromConfig: %s", err.Error())
	}
	cfg := &Config{}
	err = json.Unmarshal(bytes, cfg)
	if err != nil {
		return nil, fmt.Errorf("in function getRoomsFromConfig: %s", err.Error())
	}
	return cfg, nil
}

func main() {
	cfg, err := getConfig(configFileName)
	if err != nil {
		fmt.Printf("Error\n\r%s\n\rServing without rooms from config.json\n", err.Error())
		cfg.Address = ":8080"
	}

	usersSet := users{
		set: make(map[string]user),
		mu:  &sync.Mutex{},
	}
	messageSet := messages{
		set: make(map[string][]string),
		mu:  &sync.Mutex{},
	}
	subscribersSet := subscribers{
		set: make(map[string]map[string]string),
		mu:  &sync.Mutex{},
	}
	for _, room := range cfg.Rooms {
		messageSet.set[room] = []string{}
		subscribersSet.set[room] = make(map[string]string)
	}

	listner, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		panic(err)
	}
	defer listner.Close()

	go acceptConnections(&listner, &messageSet, &subscribersSet, &usersSet)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		// можно добавить управление сервером, например, добавление комнат или бан пользователей
	}
}

func acceptConnections(listner *net.Listener, messageSet *messages, subscribersSet *subscribers, usersSet *users) {
	for {
		conn, err := (*listner).Accept()
		if err != nil {
			fmt.Printf("Connection error: %s\n", err.Error())
		} else {
			go handleConnection(&conn, messageSet, subscribersSet, usersSet)
			fmt.Printf("%s connected\n", conn.RemoteAddr().String())
		}
	}
}

func handleConnection(conn *net.Conn, messageSet *messages, subscribersSet *subscribers, usersSet *users) {
	defer (*conn).Close()
	name := (*conn).RemoteAddr().String()
	usersSet.mu.Lock()
	usersSet.set[name] = user{
		channel:    make(chan string),
		nameInRoom: make(map[string]string),
	}
	usersSet.mu.Unlock()
	(*conn).Write([]byte("Welcome!\n"))
	go getMessages(conn, usersSet, name)

	publPattern := regexp.MustCompile("^publish ([0-9A-Za-z_]+)[\t\n\r\f ]*:[\t\n\r\f ]*(.*)$") // \s и \w не работает?
	subPattern := regexp.MustCompile("^subscribe ([0-9A-Za-z_]+)[\t\n\r\f ]*:[\t\n\r\f ]*([0-9A-Za-z_]+)$")
	scanner := bufio.NewScanner(*conn)
	for scanner.Scan() {
		cmd := scanner.Text()
		fmt.Print(cmd)
		if cmd == "exit" {
			(*conn).Write([]byte("Bye\n"))
			fmt.Println((*conn).RemoteAddr().String(), "disconnected")
			close(usersSet.set[name].channel)
			subscribersSet.mu.Lock()
			for key := range usersSet.set[name].nameInRoom {
				delete(subscribersSet.set[key], name)
			}
			subscribersSet.mu.Unlock()
			usersSet.mu.Lock()
			delete(usersSet.set, name)
			usersSet.mu.Unlock()
			break
		} else if publPattern.MatchString(cmd) {
			params := publPattern.FindStringSubmatch(cmd)
			if len(params[2]) > maxMsgSize {
				(*conn).Write([]byte(fmt.Sprintf("Sorry, but your message is too big (%d bytes). Maximum size is %d bytes\n", len(params[2]), maxMsgSize)))
				continue
			}
			nick := ""
			subscribersSet.mu.Lock()
			if roomMap, ok := subscribersSet.set[params[1]]; ok {
				if n, ok := roomMap[name]; ok {
					nick = n
				} else {
					(*conn).Write([]byte(fmt.Sprintf("You should subscribe room %s before sending messages\n", params[1])))
					continue
				}
			} else {
				(*conn).Write([]byte(fmt.Sprintf("Room with name %s doesn`t exist\n", params[1])))
				continue
			}
			subscribersSet.mu.Unlock()
			messageSet.mu.Lock()
			if len(messageSet.set[params[1]]) > maxMsgsNum {
				messageSet.set[params[1]] = messageSet.set[params[1]][1:]
			}
			messageSet.set[params[1]] = append(messageSet.set[params[1]], nick+": "+params[2])
			messageSet.mu.Unlock()
			subscribersSet.mu.Lock()
			for key := range subscribersSet.set[params[1]] {
				if key != name {
					usersSet.mu.Lock()
					usersSet.set[key].channel <- "New message in room " + params[1] + " by " + nick + ":\n" + params[2] + "\n"
					usersSet.mu.Unlock()
				}
			}
			subscribersSet.mu.Unlock()
			(*conn).Write([]byte(fmt.Sprintf("You successfully published the message\n")))
		} else if subPattern.MatchString(cmd) {
			params := subPattern.FindStringSubmatch(cmd)
			subscribersSet.mu.Lock()
			if roomMap, ok := subscribersSet.set[params[1]]; ok {
				if _, ok := roomMap[name]; ok {
					(*conn).Write([]byte(fmt.Sprintf("You are already subscribed on this room\n")))
					continue
				} else {
					for _, usedNick := range subscribersSet.set[params[1]] {
						if usedNick == params[2] {
							(*conn).Write([]byte(fmt.Sprintf("Nickname %s is already used in this room\n", usedNick)))
						}
					}
					subscribersSet.set[params[1]][name] = params[2]
				}
			} else {
				(*conn).Write([]byte(fmt.Sprintf("Room with name %s doesn`t exist\n", params[1])))
				continue
			}
			subscribersSet.mu.Unlock()
			(*conn).Write([]byte(fmt.Sprintf("You successfully connected to the room %s\nMessage history of the room:\n", params[1])))
			messageSet.mu.Lock()
			for _, m := range messageSet.set[params[1]] {
				(*conn).Write([]byte(m + "\n"))
			}
			messageSet.mu.Unlock()
		} else {
			(*conn).Write([]byte("Unknown command\n"))
		}
	}
}

func getMessages(conn *net.Conn, usersSet *users, name string) {
	for msg := range usersSet.set[name].channel {
		if strings.HasPrefix(msg, "New message") {
			(*conn).Write([]byte(msg))
		} else {
			//обработка каких-то команд из main(), например, принудительное отключение пользователя
		}
	}
}
