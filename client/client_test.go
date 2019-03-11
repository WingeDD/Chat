package main

import (
	"reflect"
	"testing"
)

func TestGetConfig(t *testing.T) {
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
	}
	badPath := "wqe"
	cfg, err = getConfig(badPath)
	if cfg != nil {
		t.Errorf("Cfg is not nil using not existing config file\n")
	} else if err == nil {
		t.Errorf("Error is nil while using not existing config file\n")
	}
	cfg, err = getConfig("client_test.go")
	if cfg != nil {
		t.Errorf("Cfg is not nil using bad config file\n")
	} else if err == nil {
		t.Errorf("Error is nil while using bad config file\n")
	}
}
