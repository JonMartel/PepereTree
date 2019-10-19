package util

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"log"

	"golang.org/x/crypto/scrypt"
)

var passwordSaltBytes = 32
var passwordHashBytes = 64

//ValidateUser : Validates the user with the specified password
func ValidateUser(user string, plainPass string) (bool, error) {
	var (
		//username string
		//name     string
		hash []byte
		salt []byte
		//disabled bool
	)

	// ...

	//Now that we have the salt/hash, let's compare!
	//pass salt ? ? ? pwhashbytes
	var valid bool = false
	dbhash, err := scrypt.Key([]byte(plainPass), salt, 1<<14, 8, 1, 64)
	if err == nil {
		valid = bytes.Equal(dbhash, hash)
	}

	return valid, err
}

//AddUser : Adds a new user to the db
func AddUser(user string, username string, pass string) {
	salt := make([]byte, passwordSaltBytes)
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		log.Fatal(err)
	}

	hash, err := scrypt.Key([]byte(pass), salt, 1<<14, 8, 1, passwordHashBytes)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("User: %s Hash: %x Salt: %x", user, hash, salt)
}
