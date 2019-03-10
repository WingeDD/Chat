package main

import (
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"
)

type testCase struct {
	requests1  []string
	requests2  []string
	responses1 []string
	responses2 []string
	order      []byte //1121111221111121...
}

func TestGetConfig(t *testing.T) {
	rooms := []string{"room1", "room2", "room3", "room4"}
	cfg, err := getConfig(configFileName)
	if cfg == nil {
		t.Errorf("Cfg is nil using right config file")
	} else if err != nil {
		t.Errorf("Error: %s while using right config file", err.Error())
	} else if !reflect.DeepEqual(cfg.Rooms, rooms) {
		t.Errorf("Room sets are not corresponding")
	} else if cfg.Address != ":8080" {
		t.Errorf("Addresses are not corresponding")
	}
	badPath := "wqe"
	cfg, err = getConfig(badPath)
	if cfg != nil {
		t.Errorf("Cfg is not nil using not existing config file")
	} else if err == nil {
		t.Errorf("Error is nil while using not existing config file")
	}
	cfg, err = getConfig("server_test.go")
	if cfg != nil {
		t.Errorf("Cfg is not nil using bad config file")
	} else if err == nil {
		t.Errorf("Error is nil while using bad config file")
	}
}

func TestMain(t *testing.T) {
	go main()
	time.Sleep(time.Second)
	cfg, _ := getConfig(configFileName)
	conn1, err := net.Dial("tcp", cfg.Address)
	if err != nil {
		t.Error("Can`t connect to server")
	}
	// conn2, err := net.Dial("tcp", cfg.Address)
	// if err != nil {
	// 	t.Error("Can`t connect to server")
	// }
	t.Log("TEST 2.1 passed")
	resp := make([]byte, 1024)
	conn1.Read(resp)
	fmt.Println(string(resp))
	conn1.Write([]byte("qwweqwe"))
	time.Sleep(time.Millisecond)
	conn1.Read(resp)
	fmt.Println(string(resp))
	fmt.Println("end")
}
