package db

import (
	"context"
	"log"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"
)

//Starting fresh!

//NewClient : creates a dgraph client we can use to make
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
