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
)

const (
	configFileName string = "serverConfig.json"
	maxMessageSize int    = 254
	maxMessagesNum int    = 128
)

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
		cfg = &Config{
			Address: ":8080",
			Rooms:   []string{"defaultRoom"},
		}
	}

	data := initServerData(cfg.Rooms)

	listner, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		panic(err)
	}
	defer listner.Close()

	go acceptConnections(&listner, data)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		// можно добавить управление сервером, например, добавление комнат или бан пользователей
	}
}

func acceptConnections(listner *net.Listener, data *serverData) {
	for {
		conn, err := (*listner).Accept()
		if err != nil {
			fmt.Printf("Connection error: %s\n", err.Error())
		} else {
			go handleConnection(&conn, data)
			fmt.Printf("%s connected\n", conn.RemoteAddr().String())
		}
	}
}

func handleConnection(conn *net.Conn, data *serverData) {
	name := (*conn).RemoteAddr().String()
	defer func() {
		fmt.Println(name, "disconnected")
		(*conn).Close()
	}()
	data.initUser(conn, name)

	publPattern := regexp.MustCompile("^publish ([0-9A-Za-z_]+)[\t\n\r\f ]*:[\t\n\r\f ]*(.*)$") // \s и \w не работает?
	subPattern := regexp.MustCompile("^subscribe ([0-9A-Za-z_]+)[\t\n\r\f ]*:[\t\n\r\f ]*([0-9A-Za-z_]+)$")
	scanner := bufio.NewScanner(*conn)
	for scanner.Scan() {
		cmd := scanner.Text()
		fmt.Println("Recieved", cmd)
		if cmd == "exit" {
			data.cleenUserData(name)
			(*conn).Write([]byte("Bye\n"))
			break
		} else if publPattern.MatchString(cmd) {
			params := publPattern.FindStringSubmatch(cmd)
			if warn, nick, ok := data.verifyPublish(params[1], params[2], name); ok {
				data.addMsg(params[1], params[2], nick)
				data.sendOutMessage(params[1], params[2], name, nick)
				(*conn).Write([]byte(fmt.Sprintf("You successfully published the message\n")))
			} else {
				(*conn).Write([]byte(warn))
			}
		} else if subPattern.MatchString(cmd) {
			params := subPattern.FindStringSubmatch(cmd)
			if warn, ok := data.registerInRoom(params[1], params[2], name); ok {
				(*conn).Write([]byte(fmt.Sprintf("You successfully connected to the room %s\nMessage history of the room:\n", params[1])))
				data.sendRoomHistory(conn, params[1])
			} else {
				(*conn).Write([]byte(warn))
			}
		} else {
			(*conn).Write([]byte("Unknown command\n"))
		}
	}
}

func getMessages(conn *net.Conn, ch chan string, data *serverData) {
	for msg := range ch {
		if strings.HasPrefix(msg, "New message") {
			(*conn).Write([]byte(msg))
		} else {
			//обработка каких-то команд из main(), например, принудительное отключение пользователя
		}
	}
}
