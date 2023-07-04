package main

import (
	"context"
	"log"
	"log-service/data"
	"time"
)

type RPCServer struct{}

type RPCPayload struct {
	Name string
	Data string
}

func (r *RPCServer) LogInfo(payload RPCPayload, resp *string) error {
	// the json is writing to collection log_entries
	collection := client.Database("logs").Collection("rpc_logs")
	_, err := collection.InsertOne(context.TODO(), data.LogEntry{
		Name:      payload.Name,
		Data:      payload.Data,
		CreatedAt: time.Now(),
	})
	log.Println("INSERTING")
	if err != nil {
		log.Println("Error inserting log entry to mongo: ", err)
		return err
	}

	*resp = "Log entry added to mongo via RPC" + payload.Name
	return nil
}
