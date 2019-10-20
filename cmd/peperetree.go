package main

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/JonMartel/PepereTree/db"
	"github.com/JonMartel/PepereTree/gedcom"
	"github.com/JonMartel/PepereTree/util"
	"github.com/JonMartel/PepereTree/webserver"
)

func main() {
	//Reference:
	// ./main -mode=parse <gedcom>
	// ./main -mode=server
	// ./main -mode=setup
	// ./main -mode=makeuser user pass
	// ./main -mode=debug <gedcom> <indi id>
	mode := flag.String("mode", "help", "Mode to run.")
	flag.Parse()

	fmt.Printf("Mode chosen: %s\n", *mode)

	switch *mode {
	case "parse":
		parseGedcom(flag.Args())
	case "makeuser":
		makeUser(flag.Args())
	case "setup":
		setupDb()
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
	fmt.Println("main -mode=setup                               #initializes dgraph db")
	fmt.Println("main -mode=makeuser <user> <fullname> <pass>   #creates a user with a password and inserts into db")
	fmt.Println("main -mode=debug <gedcom file> <individual id> #parses gedcom and prints out individual info")
}

func debug(args []string) {
	if len(args) == 2 {
		gedcom.Parse(args[0])

		id, err := strconv.ParseInt(args[1], 10, 64)
		if err == nil {
			gedcom.DisplayFamily(id)
		}
	} else {
		fmt.Println("-mode=debug <gedcom> <id>")
	}
}

func parseGedcom(args []string) {
	if len(args) != 1 {
		fmt.Println("Incorrect # of arguments: ", len(args))
		fmt.Println("Usage: -mode=parse <gedcom>")
		return
	}

	fmt.Println("Parsing: ", args[0])
	gedcom.Parse(args[0])

}

func setupDb() {
	client := db.NewClient()
	db.Init(client)
}

func makeUser(args []string) {
	if len(args) != 3 {
		fmt.Println("Specify user, full name, and a password to add a user")
	}

	util.AddUser(args[0], args[1], args[2])
}

func runServer(args []string) {
	webserver.Run()
}
