package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
)

const configFileName string = "clientConfig.json"

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
		fmt.Printf("Error occured while parsing config\n%s\nTrying to connect localhost port 8080\n", err.Error())
		cfg.Address = ":8080"
	}
	fmt.Println("Default available rooms:", cfg.Rooms)

	conn, err := net.Dial("tcp", cfg.Address)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	go getMessagesFromServer(&conn)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		cmd := scanner.Text()
		cmd = cmd + "\n"
		conn.Write([]byte(cmd))
	}

}

func getMessagesFromServer(conn *net.Conn) {
	scanner := bufio.NewScanner(*conn)
	for scanner.Scan() {
		fmt.Print("<- ")
		response := scanner.Text()
		fmt.Println(response)
	}
}

//publish <room> : <message>
//subscribe <room> : <nickname>
