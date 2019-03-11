package main

import (
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"
)

type testCase struct {
	name       string
	requests1  []string
	requests2  []string
	responses1 []string
	responses2 []string
	order      []byte //1121111221111121...
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
	srv := initserverData(254, 128, ":8080", []string{"room1", "room2", "room3", "room4"})
	go srv.Run()
	time.Sleep(time.Millisecond)
	conn1, err := net.Dial("tcp", ":8080")
	if err != nil {
		t.Error("Can`t connect to server\n")
	} else {
		fmt.Println("Connection test1 passed")
	}
	defer conn1.Close()
	conn2, err := net.Dial("tcp", ":8080")
	if err != nil {
		t.Error("Can`t connect to server\n")
	} else {
		fmt.Println("Connection test2 passed")
	}
	defer conn2.Close()

	//////////////////////////////////
	time.Sleep(time.Second)

	buf := make([]byte, len([]byte("Welcome!\n")))
	conn1.Read(buf)
	if string(buf) != "Welcome!\n" {
		t.Errorf("expected: %+v; got: %+v\n", "Welcome!\n", string(buf))
	} else {
		fmt.Println("Recieving test1 passed")
	}
	buf = make([]byte, len([]byte("Welcome!\n")))
	conn2.Read(buf)
	if string(buf) != "Welcome!\n" {
		t.Errorf("expected: %+v; got: %+v\n", "Welcome!\n", string(buf))
	} else {
		fmt.Println("Recieving test2 passed")
	}
	/*
		///////////////////////////////////
		tests := []testCase{
			testCase{
				name: "casual",
				requests1: []string{
					"qweqwe",
					"123eq",
				},
				requests2: []string{},
				responses1: []string{
					"Unknown command\n",
					"Unknown command\n",
				},
				responses2: []string{},
				order:      []byte{1, 1},
			},
			// testCase{
			// 	name:       "",
			// 	requests1:  []string{},
			// 	requests2:  []string{},
			// 	responses1: []string{},
			// 	responses2: []string{},
			// 	order:      []byte{},
			// },
			// testCase{
			// 	name:       "",
			// 	requests1:  []string{},
			// 	requests2:  []string{},
			// 	responses1: []string{},
			// 	responses2: []string{},
			// 	order:      []byte{},
			// },
			// testCase{
			// 	name:       "",
			// 	requests1:  []string{},
			// 	requests2:  []string{},
			// 	responses1: []string{},
			// 	responses2: []string{},
			// 	order:      []byte{},
			// },
		}

		for _, tcase := range tests {
			ind1 := 0
			ind2 := 0
			condition := true
			for _, clientN := range tcase.order {
				if clientN == 1 {
					conn1.Write([]byte(tcase.requests1[ind1]))
					buf := make([]byte, len(tcase.responses1[ind1]))
					time.Sleep(5 * time.Millisecond)
					fmt.Println("1")
					conn1.Read(buf)
					fmt.Println("2")
					if string(buf) != tcase.responses1[ind1] {
						condition = false
						t.Errorf("in test case: %s, conn1 i=%d expected %s got %s\n", tcase.name, ind1, tcase.responses1[ind1], string(buf))
					}
					ind1++
				} else {
					conn2.Write([]byte(tcase.requests2[ind2]))
					buf := make([]byte, len(tcase.responses2[ind2]))
					time.Sleep(5 * time.Millisecond)
					conn1.Read(buf)
					if string(buf) != tcase.responses1[ind2] {
						condition = false
						t.Errorf("in test case: %s, conn1 i=%d expected %s got %s\n", tcase.name, ind2, tcase.responses2[ind2], string(buf))
					}
					ind2++
				}
			}
			if condition {
				fmt.Println("test case", tcase.name, "passed")
			}
		}

	*/

	// +overbunden test

}
