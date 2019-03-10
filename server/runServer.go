package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
)

type Config struct {
	Address string   `json:"address"`
	Rooms   []string `json:"rooms"`
	MaxS    int      `json:"maxs"`
	MaxN    int      `json:"maxn"`
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

func initserverData(maxS, maxN int, addr string, rooms []string) *server {
	srv := &server{
		mu:          sync.Mutex{},
		Address:     addr,
		maxMsgSize:  maxS,
		maxMsgsNum:  maxN,
		control:     make(chan string),
		users:       make(map[string]user),
		messages:    make(map[string][]string),
		subscribers: make(map[string]map[string]string),
	}

	for _, room := range rooms {
		srv.messages[room] = []string{}
		srv.subscribers[room] = make(map[string]string)
	}

	return srv
}

func main() {
	const configFileName = "serverConfig.json"
	cfg, err := getConfig(configFileName)
	if err != nil {
		fmt.Printf("Error\n\r%s\n\rServing without rooms from config.json\n", err.Error())
		cfg = &Config{
			Address: ":8080",
			Rooms:   []string{"defaultRoom"},
			MaxS:    254,
			MaxN:    128,
		}
	}
	srv := initserverData(cfg.MaxS, cfg.MaxN, cfg.Address, cfg.Rooms)
	srv.Run()
}
