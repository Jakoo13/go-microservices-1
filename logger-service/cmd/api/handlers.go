package main

import (
	"net/http"
	"log-service/data"
)

type JSONPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) WriteLog(w http.ResponseWriter, r *http.Request) {
	// read json into our var
	var requestPayload JSONPayload 
		_ = app.readJSON(w, r, &requestPayload)
	
	// insert data
	event := data.LogEntry{
		Name: requestPayload.Name,
		Data: requestPayload.Data,
	}
	err := app.Models.LogEntry.Insert(event)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	response := jsonResponse{
		Error:   false,
		Message: "Logged",
	}

	app.writeJSON(w, http.StatusAccepted, response)
}
