package peperedb

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"fmt"
	"io"
	"log"
	"time"

	"golang.org/x/crypto/scrypt"

	_ "github.com/go-sql-driver/mysql"

	"gedcom"
)

var connections = make(map[string]*Connection)
var PW_SALT_BYTES = 32
var PW_HASH_BYTES = 64

type Connection struct {
	db *sql.DB
}

func NewConnection(database string) (*Connection, error) {
	var retval *Connection = nil

	if conn, ok := connections[database]; ok {
		return conn, nil
	}

	db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/"+database)
	if err == nil {
		retval = new(Connection)
		retval.db = db
		connections[database] = retval
	}
	//defer db.Close()
	return retval, err
}

//Methods for getting data populated!
func (conn *Connection) GetUser(user string) (string, []byte, []byte) {
	var (
		name string
		hash []byte
		salt []byte
	)

	stmt, err := conn.db.Prepare("select username, hash, salt from users where username = ?")
	if err != nil {
		log.Fatal("Couldn't prepare Get User query", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(user)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&name, &hash, &salt)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(name)
	}

	//Very important, apparently!
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return name, hash, salt
}

func (conn *Connection) ValidateUser(user string, plainPass string) (bool, error) {
	var (
		username string
		name     string
		hash     []byte
		salt     []byte
		disabled bool
	)

	stmt, err := conn.db.Prepare("select username, fullname, hash, salt, disabled from users where username = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(user)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&username, &name, &hash, &salt, &disabled)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(name)
	}

	//Now that we have the salt/hash, let's compare!
	//pass salt ? ? ? pwhashbytes
	var valid bool = false
	dbhash, err := scrypt.Key([]byte(plainPass), salt, 1<<14, 8, 1, 64)
	if err == nil {

		//Very important, apparently!
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}

		valid = bytes.Equal(dbhash, hash)
	}

	return valid, err
}

func (conn *Connection) AddUser(user string, username string, pass string) {
	salt := make([]byte, PW_SALT_BYTES)
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		log.Fatal(err)
	}

	hash, err := scrypt.Key([]byte(pass), salt, 1<<14, 8, 1, PW_HASH_BYTES)
	if err != nil {
		log.Fatal(err)
	}

	if err == nil {
		stmt, err := conn.db.Prepare("replace into users values (?, ?, ?, ?, false)")
		defer stmt.Close()
		if err != nil {
			log.Fatal("Failed to prepare statement for adding user")
		} else {
			_, err = stmt.Exec(user, username, hash, salt)
			if err != nil {
				log.Fatal("Failed to add user")
			}
		}
	} else {
		log.Fatal("Could not get db connection")
	}
}

//Initializes the database, dropping tables if needed. Allows us to start fresh whenever making changes!
func (conn *Connection) InitializeDatabase() error {

	//Drop any existing tables
	dropStmt, err := conn.db.Prepare("DROP TABLE IF EXISTS individual")
	if err == nil {
		defer dropStmt.Close()
		dropStmt.Exec()
	} else {
		return err
	}

	//Create the tables
	stmt, err := conn.db.Prepare("CREATE TABLE individual (id int, fullname text, title text, sex tinytext, occupation text, aliases text, notes mediumtext, lastupdate timestamp, PRIMARY KEY(id)) ENGINE=InnoDB DEFAULT CHARSET=utf8")
	if err == nil {
		defer stmt.Close()

		_, err = stmt.Exec()
		if err != nil {
			return err
		}
	}

	//username, name, hash, salt, disabled
	_, err = conn.db.Query("DROP TABLE IF EXISTS users")
	if err != nil {
		return err
	}
	_, err = conn.db.Query("CREATE TABLE users (username text, fullname text, hash blob, salt blob, disabled bool, KEY(username(5))) ENGINE=InnoDB DEFAULT CHARSET=utf8")
	if err != nil {
		return err
	}

	//	EventTime time.Time, EventType string, SourceId  int, Location  string
	_, err = conn.db.Query("DROP TABLE IF EXISTS event")
	if err != nil {
		return err
	}

	_, err = conn.db.Query("CREATE TABLE event (id int, sourceid int, year smallint, month smallint, day smallint, type tinytext, location text, KEY(id)) ENGINE=InnoDB DEFAULT CHARSET=utf8")
	if err != nil {
		return err
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

	_, err = conn.db.Query("DROP TABLE IF EXISTS relation")
	if err != nil {
		return err
	}

	_, err = conn.db.Query("CREATE TABLE relation (fid int, tid int, type smallint, notes mediumtext) ENGINE=InnoDB DEFAULT CHARSET=utf8")

	/*
		_, err = conn.db.Query("CREATE TABLE individual_family_head (id int, famid int, KEY(id), KEY(famid)) ENGINE=InnoDB")
		if err != nil {
			return err
		}

		_, err = conn.db.Query("CREATE TABLE individual_family_child (id int, famid int, KEY(id), KEY(famid)) ENGINE=InnoDB")
		if err != nil {
			return err
		}

		_, err = conn.db.Query("CREATE TABLE family (famid int, , PRIMARY KEY(id)) ENGINE=InnoDB")
		if err != nil {
			return err
		}
	*/

	return nil
}

//Family/Individual/Source/etc population methods
func (conn *Connection) ImportGedcomData() error {
	//fmt.Println(gedcom.Individuals[456])

	//INSERT INTO individual VALUES( id, fullname, title, sex, occupation, aliases, notes)
	//lastupdate is automatic!
	//aliases and notes need to be combined into one singular string
	//events have their own table
	//indi-to-family has own table
	//indi-to-child-fam has own table
	//indi-to-media has own table
	fmt.Println(time.Now())
	for _, indi := range gedcom.Individuals {
		//Concat the fields that require it
		var aliasBuf bytes.Buffer
		for _, alias := range indi.Aliases {
			aliasBuf.WriteString(alias)
			aliasBuf.WriteString(" ")
		}

		var notesBuf bytes.Buffer
		for _, note := range indi.Notes {
			notesBuf.WriteString(note)
			notesBuf.WriteString("\n")
		}

		_, err := conn.db.Exec("INSERT INTO individual VALUES (?, ?, ?, ?, ?, ?, ?, ?)", indi.Id, indi.Fullname, indi.Title, indi.Gender, indi.Occupation, aliasBuf.String(), notesBuf.String(), nil)
		if err != nil {
			fmt.Println("Error inserting individual:", err)
			return err
		}

		//id , sid , date , type , location
		for _, ev := range indi.Events {
			_, err := conn.db.Exec("INSERT INTO event VALUES(?, ?, ?, ?, ?, ?, ?)", indi.Id, ev.SourceId, ev.EventYear, ev.EventMonth, ev.EventDay, ev.EventType, ev.Location)
			if err != nil {
				fmt.Println("Error inserting event:", err)
				return err
			}
		}
	}

	_ = importFamilyRelationships(conn)

	//Why is this taking 5 minutes to complete?
	//Let's try improving it later!

	fmt.Println(time.Now())

	return nil
}

func importFamilyRelationships(conn *Connection) error {

	var fRel bytes.Buffer
	fRel.WriteString("Father")

	for _, fam := range gedcom.Families {

		father := fam.Father
		mother := fam.Mother

		var notesBuf bytes.Buffer
		for _, note := range fam.Notes {
			notesBuf.WriteString(note)
			notesBuf.WriteString("\n")
		}

		//Father to Mother Relation
		//Mother to Father Relation

		for _, child := range fam.ChildIds {

			//fid, tid, type, notes, syear, smonth, sday, eyear, emonth, eday
			//Father to Kid relation
			_, err := conn.db.Exec("INSERT INTO relation VALUES(?, ?, ?, ?)", father, child, 1, notesBuf.String())
			if err != nil {
				fmt.Println("Error adding father relation", err)
				return err
			}
			_, err = conn.db.Exec("INSERT INTO relation VALUES(?, ?, ?, ?)", child, father, 3, notesBuf.String())
			if err != nil {
				fmt.Println("Error adding child to father relation", err)
				return err
			}

			//Mother to Kid relation
			conn.db.Exec("INSERT INTO relation VALUES(?, ?, ?, ?)", mother, child, 2, notesBuf.String())
			if err != nil {
				fmt.Println("Error adding mother relation", err)
				return err
			}
			conn.db.Exec("INSERT INTO relation VALUES(?, ?, ?, ?)", child, mother, 3, notesBuf.String())
			if err != nil {
				fmt.Println("Error adding child to mother relation", err)
				return err
			}
			//_, err := conn.db.Exec("INSERT INTO relation VALUES(?, ?, ?, ?)")
			//_, err := conn.db.Exec("INSERT INTO relation VALUES(?, ?, ?, ?)")

			//Kid to other kid relations?
		}
	}

	return nil
}
