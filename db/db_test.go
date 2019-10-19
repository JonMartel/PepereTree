package db

import (
	"testing"
)

func TestGettingConnection(t *testing.T) {
	conn, err := NewConnection("peperetree")
	if err != nil {
		t.Error("Failed to get a Connection!")
	}

	if conn == nil {
		t.Error("Connection is nil!")
	}
}

func TestValidatingUser(t *testing.T) {
	conn, err := NewConnection("peperetree")

	if err != nil {
		t.Error("Failed to get connection!")
	} else {
		valid, err := conn.ValidateUser("test", "nottherightpass")
		if err != nil {
			t.Error("Error validating user")
		}
		if valid {
			t.Error("Accepted the wrong password!")
		}
		valid, err = conn.ValidateUser("test", "testpass")
		if err != nil {
			t.Error("Validated ")
		}
		if !valid {
			t.Error("Did not accept correct password!")
		}
	}
}
