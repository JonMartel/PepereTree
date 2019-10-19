package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"

	"github.com/JonMartel/PepereTree/db"
	"github.com/JonMartel/PepereTree/gedcom"
	"github.com/JonMartel/PepereTree/webserver"
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
	fmt.Println("Usage: main -mode=<mode> arg1 arg2 arg3")
	fmt.Println("Available modes:")
	fmt.Println("main -mode=parse <gedcom file>                 #converts gedcom and imports to mysql")
	fmt.Println("main -mode=server                              #starts webserver")
	fmt.Println("main -mode=makeuser <user> <fullname> <pass>   #creates a user with a password and inserts into db")
	fmt.Println("main -mode=debug <gedcom file> <individual id> #parses gedcom and prints out individual info")
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
			fmt.Println(martels)
			fmt.Println(gedcom.Individuals[martels.Father])
			fmt.Println(gedcom.Individuals[martels.Mother])
			for _, child := range martels.ChildIds {
				fmt.Println(gedcom.Individuals[child])

				p := gedcom.Individuals[child]
				for _, ev := range p.Events {
					fmt.Println(ev)
				}
			}
		}
	} else {
		fmt.Println("-mode=debug <gedcom> <id>")
	}
}

func parseGedcom(args []string) {
	if len(args) == 1 {
		fmt.Println("Parsing: ", args[0])
		gedcom.Parse(args[0])

		//Now that it is parsed, let's populate the database with the data we have!
		conn, err := db.NewConnection("peperetree")
		if err == nil {
			conn.GetUser("fake")
		}
	} else {
		fmt.Println("Incorrect # of arguments: ", len(args))
		fmt.Println("Usage: -mode=parse <gedcom>")
	}
}

func makeUser(args []string) {
	if len(args) == 3 {
		conn, err := db.NewConnection("peperetree")
		if err == nil {
			conn.AddUser(args[0], args[1], args[2])
		} else {
			log.Fatal("Failed to add user!", err)
		}
	} else {
		fmt.Println("Specify user, full name, and a password to add a user")
	}
}

func runServer(args []string) {
	webserver.RunServer()
}
