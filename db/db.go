package db

import (
	"context"
	"encoding/json"
	"log"

	"github.com/JonMartel/PepereTree/gedcom"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"
)

//NewClient : creates a dgraph client
func NewClient() *dgo.Dgraph {
	conn, err := grpc.Dial("localhost:9080", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	client := api.NewDgraphClient(conn)

	return dgo.NewDgraphClient(client)
}

//Init : Initializes the database
func Init(client *dgo.Dgraph) {
	err := client.Alter(context.Background(), &api.Operation{
		Schema: `
			type Person {
				name: string
				age: int
				friend: [Person]
			}
			
			type Animal {
				name: string
			}
			
			# Define Directives and index
			
			name: string @index(term) @lang .
			age: int @index(int) .
			friend: [uid] @count .
		`,
	})

	if err != nil {
		log.Fatal(err)
	}
}

//Import imports the data stored in our gedcom maps into the provided dgraph instance
func Import(client *dgo.Dgraph) {
	people, _, _ := gedcom.GetGedcomData()

	for _, person := range people {
		trx := client.NewTxn()

		//We defer a discard - if anything bad happens, transaction is closed
		//If we complete the transacation, the defered discard is a no-op
		defer trx.Discard(context.Background())

		marshalled, err := json.Marshal(person)
		if err != nil {
			log.Fatalln("Error marshalling data: ", err)
		}

		_, err = trx.Mutate(context.Background(), &api.Mutation{
			SetJson: marshalled,
		})

		if err != nil {
			log.Fatalln("Error importing individual data: ", err)
		}
	}
}
