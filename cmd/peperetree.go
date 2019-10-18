package main

import (
	"flag"
	"log"
	"strconv"

	"gedcom"
	"peperedb"
	"webserver"
)

func main() {
	//Reference:
	// ./main -mode=parse <gedcom>
	// ./main -mode=server
	// ./main -mode=makeuser user pass
	// ./main -mode=debug <gedcom> <indi id>
	mode := flag.String("mode", "help", "Mode to run. Options are:\n parse - Parse gedcom to db\n makeuser - Creates a user\n server - Run webserver [default]")
	flag.Parse()

	switch *mode {
	case "parse":
		parseGedcom(flag.Args())
	case "makeuser":
		makeUser(flag.Args())
	case "server":
		runServer(flag.Args())
	case "debug":
		debug(flag.Args())
	case "help":
		fallthrough
	default:
		usage()
	}
}

func usage() {
	Println("Usage: main -mode=<mode> arg1 arg2 arg3")
	Println("Available modes:")
	Println("main -mode=parse <gedcom file>                 #converts gedcom and imports to mysql")
	Println("main -mode=server                              #starts webserver")
	Println("main -mode=makeuser <user> <fullname> <pass>   #creates a user with a password and inserts into db")
	Println("main -mode=debug <gedcom file> <individual id> #parses gedcom and prints out individual info")
}

func debug(args []string) {
	if len(args) == 2 {
		gedcom.Parse(args[0])

		id, err := strconv.ParseInt(args[1], 10, 32)
		if err == nil {
			//Println("Person is:")
			//p := gedcom.Individuals[int(id)]
			//Println(p)

			martels := gedcom.Families[int(id)]
			Println(martels)
			Println(gedcom.Individuals[martels.Father])
			Println(gedcom.Individuals[martels.Mother])
			for _, child := range martels.ChildIds {
				Println(gedcom.Individuals[child])

				p := gedcom.Individuals[child]
				for _, ev := range p.Events {
					Println(ev)
				}
			}
		}
	} else {
		Println("-mode=debug <gedcom> <id>")
	}
}

func parseGedcom(args []string) {
	if len(args) == 1 {
		Println("Parsing: ", args[0])
		gedcom.Parse(args[0])

		//Now that it is parsed, let's populate the database with the data we have!
		conn, err := peperedb.NewConnection("peperetree")
		if err == nil {
			err = conn.InitializeDatabase()
			if err == nil {
				conn.ImportGedcomData()
			} else {
				Println("Error initializing database schema: ", err)
			}
		}
	} else {
		Println("Incorrect # of arguments: ", len(args))
		Println("Usage: -mode=parse <gedcom>")
	}
}

func makeUser(args []string) {
	if len(args) == 3 {
		conn, err := peperedb.NewConnection("peperetree")
		if err == nil {
			conn.AddUser(args[0], args[1], args[2])
		} else {
			log.Fatal("Failed to add user!", err)
		}
	} else {
		Println("Specify user, full name, and a password to add a user")
	}
}

func runServer(args []string) {
	webserver.RunServer()
}
