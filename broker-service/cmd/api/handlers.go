package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type RequestPayload struct {
	Action   string          `json:"action"`
	Register RegisterPayload `json:"register,omitempty"`
	Login    LoginPayload    `json:"login,omitempty"`
}

type RegisterPayload struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type LoginPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Hello from the broker!",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)

	// can remove this iwth the helper function above
	// out, _ := json.MarshalIndent(payload, "", "\t")
	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusAccepted)
	// w.Write(out)
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	switch requestPayload.Action {
	case "register":
		app.Register(w, requestPayload.Register)
	case "login":
		app.Login(w, requestPayload.Login)
	default:
		app.errorJSON(w, errors.New("unknown action"))
	}
}

func (app *Config) Register(w http.ResponseWriter, registerPayload RegisterPayload) {
	// Create some json we'll send to the auth microservice
	jsonData, _ := json.MarshalIndent(registerPayload, "", "\t")

	// Call the service
	request, err := http.NewRequest("POST", "http://authentication-service/register", bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	log.Default().Println("MADE IT THIS FAR")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	// make sure we get back the correct status code
	if response.StatusCode == http.StatusConflict {
		app.errorJSON(w, errors.New("email address already registered"))
		return
	}

	// create a variable we'll read the response.Body into
	var jsonFromService jsonResponse

	// decode the json from the auth service
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		app.errorJSON(w, err, http.StatusUnauthorized)
		return
	}

	// We have valid Register if we reach here. Send the response back to the client
	var payload jsonResponse
	payload.Error = false
	payload.Message = "Registered!"
	payload.Data = jsonFromService.Data

	app.writeJSON(w, http.StatusAccepted, payload)

}

func (app *Config) Login(w http.ResponseWriter, loginPayload LoginPayload) {
	// Create some json we'll send to the auth microservice
	jsonData, _ := json.MarshalIndent(loginPayload, "", "\t")

	// Call the service
	request, err := http.NewRequest("POST", "http://authentication-service/login", bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()
	log.Default().Println("STATUS CODE: ", response.StatusCode)
	// make sure we get back the correct status code
	if response.StatusCode == http.StatusUnauthorized {
		app.errorJSON(w, errors.New("invalid credentials"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error calling authentication service"))
		return
	}

	// create a variable we'll read the response.Body into
	var jsonFromService jsonResponse

	// decode the json from the auth service
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		app.errorJSON(w, err, http.StatusUnauthorized)
		return
	}

	// We have valid login if we reach here. Send the response back to the client
	var payload jsonResponse
	payload.Error = false
	payload.Message = "Authenticated!"
	payload.Data = jsonFromService.Data

	app.writeJSON(w, http.StatusAccepted, payload)
	// log.Default().Println(payload.Data)
}
