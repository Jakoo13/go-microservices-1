package main

import (
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	webPort = "80"
	rpcPort = "5001"
	mongoURL = "mongodb://mongo:27017"
	gRpcPort = "50001"
)

var client *mongo.Client



func main() {

}