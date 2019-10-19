package db

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"log"

	"golang.org/x/crypto/scrypt"
)

var connections = make(map[string]*Connection)
var passwordSaltBytes = 32
var passwordHashBytes = 64

//Connection : Represents a db connection
type Connection struct {
	//replace me with whatever
}

//NewConnection : Creates a new DB connection
func NewConnection(database string) (*Connection, error) {
	var retval *Connection = nil

	if conn, ok := connections[database]; ok {
		return conn, nil
	}

	return retval, nil
}

//GetUser : for the given username, retrieves the real name, hash, and the salt
func (conn *Connection) GetUser(user string) (string, []byte, []byte) {
	var (
		name string
		hash []byte
		salt []byte
	)

	// ...

	return name, hash, salt
}

//ValidateUser : Validates the user with the specified password
func (conn *Connection) ValidateUser(user string, plainPass string) (bool, error) {
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
func (conn *Connection) AddUser(user string, username string, pass string) {
	salt := make([]byte, passwordSaltBytes)
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		log.Fatal(err)
	}

	//hash, err :=
	_, err = scrypt.Key([]byte(pass), salt, 1<<14, 8, 1, passwordHashBytes)
	if err != nil {
		log.Fatal(err)
	}

	if err == nil {
		fmt.Println("Attempted login")
	}
}

/*
	Id       int
	Father   int
	Mother   int
	ChildIds []int
	Notes    []string
*/

//How can we handle relationships for those in families? maybe instead of using a family structure, we create direct relationships?
//Then, we can easily track changes on an individual level
//A 'Stepson' can still be 1/2 biological for a family, adopted would be 0/2, weird cases like Joel's are just weird (2.5? parents?)
//Having individual relationships between people would make things easier!
//Relationships can be directed sideways (marriage, partner, etc), up (parent, stepparent), down(child, stepchild)
//for completeness, lets 'duplicate' relationships - that is, if A is married to B, have both A->B and B->A
//this means we don't need a convention (husband->wife, for example) since it'll work for those weird cases).
