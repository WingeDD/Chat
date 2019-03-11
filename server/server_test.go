package main

import (
	"bytes"
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"
)

type order struct { //запросы и ответы во времени
	client int    //number
	action string // r/w
}

type testCase struct {
	name      string
	requests  [][]string
	responses [][]string
	ord       []order
}

func TestGetConfig(t *testing.T) {
	configFileName := "serverConfig.json"
	rooms := []string{"room1", "room2", "room3", "room4"}
	cfg, err := getConfig(configFileName)
	if cfg == nil {
		t.Errorf("Cfg is nil using right config file\n")
	} else if err != nil {
		t.Errorf("Error: %s while using right config file\n", err.Error())
	} else if !reflect.DeepEqual(cfg.Rooms, rooms) {
		t.Errorf("Room sets are not corresponding\n")
	} else if cfg.Address != ":8080" {
		t.Errorf("Addresses are not corresponding\n")
	} else if cfg.MaxN != 128 {
		t.Errorf("maxN are not corresponding\n")
	} else if cfg.MaxS != 254 {
		t.Errorf("maxS are not corresponding\n")
	}
	badPath := "wqe"
	cfg, err = getConfig(badPath)
	if cfg != nil {
		t.Errorf("Cfg is not nil using not existing config file\n")
	} else if err == nil {
		t.Errorf("Error is nil while using not existing config file\n")
	}
	cfg, err = getConfig("server_test.go")
	if cfg != nil {
		t.Errorf("Cfg is not nil using bad config file\n")
	} else if err == nil {
		t.Errorf("Error is nil while using bad config file\n")
	}
}

func TestMain(t *testing.T) {
	const clientsNum = 2
	srv := initserverData(254, 128, ":8080", []string{"room1", "room2", "room3", "room4"})
	go srv.Run()
	time.Sleep(5 * time.Millisecond)

	connections := make([]net.Conn, clientsNum)
	for i := 0; i < clientsNum; i++ {
		con, err := net.Dial("tcp", ":8080")
		connections[i] = con
		if err != nil {
			t.Error("Can`t connect to server\n")
			return
		} else {
			fmt.Printf("Connection test%d passed\n", i)
			defer connections[i].Close()
		}
	}

	//////////////////////////////////
	time.Sleep(time.Millisecond)

	for i := 0; i < clientsNum; i++ {
		buf := make([]byte, len([]byte("Welcome!\n")))
		connections[i].Read(buf)
		if string(buf) != "Welcome!\n" {
			t.Errorf("expected: %+v; got: %+v\n", "Welcome!\n", string(buf))
		} else {
			fmt.Printf("Recieving test%d passed\n", i)
		}
	}

	///////////////////////////////////
	tests := []testCase{
		testCase{
			name: "casual",
			requests: [][]string{
				[]string{
					"qweqwe\n",
					"subscribe room1 : cl1\n",
					"subscribe room2 : cl1\n",
					"publish room1 : hey\n",
					"publish room1 : its me\n",
				},
				[]string{
					"subscribe room1 : cl2\n",
					"subscribe room2 : cl2\n",
					"publish room2 : 123\n",
				},
			},
			responses: [][]string{
				[]string{
					"Unknown command\n",
					"You successfully connected to the room room1. Message history:\n",
					"You successfully connected to the room room2. Message history:\n",
					"You successfully published the message\n",
					"You successfully published the message\n",
					"New message in room room2 by cl2: 123\n",
				},
				[]string{
					"You successfully connected to the room room1. Message history:\n",
					"cl1: hey\n",
					"cl1: its me\n",
					"You successfully connected to the room room2. Message history:\n",
					"You successfully published the message\n",
				},
			},
			ord: []order{
				order{
					client: 0,
					action: "w",
				},
				order{
					client: 0,
					action: "r",
				},
				order{
					client: 0,
					action: "w",
				},
				order{
					client: 0,
					action: "r",
				},
				order{
					client: 0,
					action: "w",
				},
				order{
					client: 0,
					action: "r",
				},
				order{
					client: 0,
					action: "w",
				},
				order{
					client: 0,
					action: "r",
				},
				order{
					client: 0,
					action: "w",
				},
				order{
					client: 0,
					action: "r",
				},
				order{
					client: 1,
					action: "w",
				},
				order{
					client: 1,
					action: "r",
				},
				order{
					client: 1,
					action: "r",
				},
				order{
					client: 1,
					action: "r",
				},
				order{
					client: 1,
					action: "w",
				},
				order{
					client: 1,
					action: "r",
				},
				order{
					client: 1,
					action: "w",
				},
				order{
					client: 1,
					action: "r",
				},
				order{
					client: 0,
					action: "r",
				},
			},
		},
		testCase{
			name: "too long message",
			requests: [][]string{
				[]string{
					"subscribe room3 : cl1\n",
					"publish room3 : qweqweqweqweqweqweqweqweqweqweqweqweqweqweqweqweqweqeqweqweqweqweqweqweqqeqweqewqeweqwqeweqweqwweqqweqweqweewqewqweeqwqeweqweqweqweqweqweqwewqweqweqweqweqweqweqewqewqweqweqweqweqweqweqweqweqweqweqweqweqweqweqweqweqwekqwemqkwmekqmwemqwkemqkwmekqwmemqwkemkqwmeqwkekqrnqweruqwequrqwjrqwjqriwqwrojiqrwji\n",
				},
				[]string{},
			},
			responses: [][]string{
				[]string{
					"You successfully connected to the room room3. Message history:\n",
					"Sorry, but your message is too long (299 bytes). Maximum size is 254 bytes\n",
				},
				[]string{},
			},
			ord: []order{
				order{
					client: 0,
					action: "w",
				},
				order{
					client: 0,
					action: "r",
				},
				order{
					client: 0,
					action: "w",
				},
				order{
					client: 0,
					action: "r",
				},
			},
		},
	}

	for _, tcase := range tests {
		currentIndexesW := make([]int, len(tcase.requests))
		currentIndexesR := make([]int, len(tcase.responses))
		condition := true
		for _, ord := range tcase.ord {
			time.Sleep(time.Millisecond)
			if ord.action == "r" {
				buf := make([]byte, len(tcase.responses[ord.client][currentIndexesR[ord.client]]))
				connections[ord.client].Read(buf)
				if !bytes.Equal(buf, []byte(tcase.responses[ord.client][currentIndexesR[ord.client]])) {
					condition = false
					t.Errorf("in test case: %s, connection%d iter=%d expected: %s got: %s\n%v\n%v\n", tcase.name, ord.client, currentIndexesR[ord.client], tcase.responses[ord.client][currentIndexesR[ord.client]], string(buf), []byte(tcase.responses[ord.client][currentIndexesR[ord.client]]), buf)
				}
				currentIndexesR[ord.client]++
			} else if ord.action == "w" {
				connections[ord.client].Write([]byte(tcase.requests[ord.client][currentIndexesW[ord.client]]))
				currentIndexesW[ord.client]++
			} else {
				fmt.Println("bad case!")
				return
			}
		}
		if condition {
			fmt.Println("test case", tcase.name, "passed")
		}
	}

	//////////////////////////////

	connections[0].Write([]byte("subscribe room4 : cl1\n"))
	connections[0].Write([]byte("publish room4 : first msg\n"))
	for i := 0; i < 129; i++ {
		connections[0].Write([]byte("publish room4 : others\n"))
	}
	time.Sleep(time.Second)
	connections[1].Write([]byte("subscribe room4 : cl2\n"))
	time.Sleep(time.Millisecond)
	buf1 := make([]byte, len("You successfully connected to the room room4. Message history:\n"))
	buf2 := make([]byte, len("cl1: others\n"))
	connections[1].Read(buf1)
	connections[1].Read(buf2)
	if string(buf1) != "You successfully connected to the room room4. Message history:\n" || string(buf2) != "cl1: others\n" {
		t.Errorf("Extra messages are not deleted\n")
	} else {
		fmt.Println("overflow test passed")
	}

	///////////////////////////
	srv.Stop()
}
