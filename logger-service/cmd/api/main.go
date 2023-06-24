package main

import (
	"context"
	"fmt"
	"log"
	"log-service/data"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	webPort  = "80"
	rpcPort  = "5001"
	mongoURL = "mongodb://mongo:27017"
	gRpcPort = "50001"
)

var client *mongo.Client

type Config struct {
	Models data.Models
}

func main() {
	// connect to mongo
	mongoClient, err := connectToMongo()
	if err != nil {
		log.Panic(err)
	}

	client = mongoClient

	// Create a context in order to disconnect
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// close the connection
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	// set up config
	app := Config{
		Models: data.New(client),
	}

	// start web server
	// go app.serve()
	// set up Web Server
	log.Println("Starting service on port ", webPort)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	// start Web Server
	err = srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

// New function to start up webserver to make things easier when using gRPC, could do this within main but this is cleaner
// func (app *Config) serve() {
// 	// set up Web Server
// 	srv := &http.Server{
// 		Addr:    fmt.Sprintf(":%s", webPort),
// 		Handler: app.routes(),
// 	}

// 	// start Web Server
// 	err := srv.ListenAndServe()
// 	if err != nil {
// 		log.Panic(err)
// 	}

// }

func connectToMongo() (*mongo.Client, error) {
	// create connection options
	clientOptions := options.Client().ApplyURI(mongoURL)
	clientOptions.SetAuth(options.Credential{
		// In production, get these from environment variables
		Username: "admin",
		Password: "password",
	})

	// connect to mongo
	connection, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Println("Error connecting to mongo: ", err)
		return nil, err
	}

	return connection, nil
}
