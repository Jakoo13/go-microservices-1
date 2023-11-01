package main

import (
	// "broker/event"
	"broker/logs"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/rpc"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RequestPayload struct {
	Action   string          `json:"action"`
	Register RegisterPayload `json:"register,omitempty"`
	Login    LoginPayload    `json:"login,omitempty"`
	Log      LogPayload      `json:"log,omitempty"`
	Mail     MailPayload     `json:"mail,omitempty"`
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

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
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
	case "log":
		// app.logItem(w, requestPayload.Log)
		// app.logEventViaRabbit(w, requestPayload.Log)
		app.logItemViaRPC(w, requestPayload.Log)
	case "register":
		app.register(w, requestPayload.Register)
	case "login":
		app.login(w, requestPayload.Login)
	case "getAllUsers":
		app.getAllUsers(w)
	case "mail":
		app.sendMail(w, requestPayload.Mail)
	default:
		app.errorJSON(w, errors.New("unknown action"))
	}
}

// func (app *Config) logItem(w http.ResponseWriter, logPayload LogPayload) {
// 	// TODO: in prod dont use MarshalIndent, use Marshal
// 	jsonData, _ := json.MarshalIndent(logPayload, "", "\t")

// 	logServiceURL := "http://logger-service/log"

// 	// Build request
// 	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
// 	if err != nil {
// 		app.errorJSON(w, err)
// 		return
// 	}

// 	// Set Headers
// 	request.Header.Set("Content-Type", "application/json")

// 	// Send request
// 	client := &http.Client{}
// 	response, err := client.Do(request)
// 	if err != nil {
// 		app.errorJSON(w, err)
// 		return
// 	}
// 	defer response.Body.Close()

// 	// make sure we get back the correct status code
// 	if response.StatusCode != http.StatusAccepted {
// 		app.errorJSON(w, errors.New("error calling logger service"))
// 		return
// 	}

// 	var payload jsonResponse
// 	payload.Error = false
// 	payload.Message = "Logged!"

// 	app.writeJSON(w, http.StatusAccepted, payload)
// }

func (app *Config) register(w http.ResponseWriter, registerPayload RegisterPayload) {
	// Create some json we'll send to the auth microservice
	jsonData, _ := json.MarshalIndent(registerPayload, "", "\t")

	// Call the service
	request, err := http.NewRequest("POST", "http://authentication-service/register", bytes.NewBuffer(jsonData))
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

func (app *Config) login(w http.ResponseWriter, loginPayload LoginPayload) {
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

func (app *Config) getAllUsers(w http.ResponseWriter) {
	// Call the service
	request, err := http.NewRequest("GET", "http://authentication-service/user", nil)
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

	// make sure we get back the correct status code
	if response.StatusCode != http.StatusAccepted {
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
}

func (app *Config) sendMail(w http.ResponseWriter, mailPayload MailPayload) {
	// Create some json we'll send to the auth microservice
	jsonData, _ := json.MarshalIndent(mailPayload, "", "\t")

	// Call the service
	// url defined in docker compose first line
	mailServiceUrl := "http://mailer-service/send"
	request, err := http.NewRequest("POST", mailServiceUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	// make sure we get back the correct status code
	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error calling mail service"))
		return
	}

	// send back JSON
	var payload jsonResponse
	payload.Error = false
	payload.Message = "Mail sent to " + mailPayload.To

	app.writeJSON(w, http.StatusAccepted, payload)

}

// func (app *Config) logEventViaRabbit(w http.ResponseWriter, logPayload LogPayload) {
// 	err := app.pushToQueue(logPayload.Name, logPayload.Data)
// 	if err != nil {
// 		app.errorJSON(w, err)
// 		return
// 	}

// 	var payload jsonResponse
// 	payload.Error = false
// 	payload.Message = "Logged via RabbitMQ!"

// 	app.writeJSON(w, http.StatusAccepted, payload)
// }

// func (app *Config) pushToQueue(name, msg string) error {
// 	emitter, err := event.NewEventEmitter(app.Rabbit)
// 	if err != nil {
// 		return err
// 	}

// 	payload := LogPayload{
// 		Name: name,
// 		Data: msg,
// 	}

// 	jsonData, _ := json.MarshalIndent(payload, "", "\t")
// 	err = emitter.Push(string(jsonData), "log.INFO")
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// has to exactly match the server side type
type RPCPayload struct {
	Name string
	Data string
}

func (app *Config) logItemViaRPC(w http.ResponseWriter, logPayload LogPayload) {
	client, err := rpc.Dial("tcp", "logger-service:5001")
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	logP := LogPayload{
		Name: logPayload.Name,
		Data: logPayload.Data,
	}

	rpcPayload := RPCPayload(logP)

	// Populated by the remote rpc call response
	var result string
	err = client.Call("RPCServer.LogInfo", rpcPayload, &result)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	// if everything went well, we'll write some JSON back to the user
	payload := jsonResponse{
		Error:   false,
		Message: result,
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) LogViaGRPC(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	// Specify the credentials here with second param
	conn, err := grpc.Dial("logger-service:50001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer conn.Close()

	// create a client instance
	client := logs.NewLogServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = client.WriteLog(ctx, &logs.LogRequest{
		LogEntry: &logs.Log{
			Name: requestPayload.Log.Name,
			Data: requestPayload.Log.Data,
		},
	})
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	// if everything went well, we'll write some JSON back to the user
	var payload jsonResponse
	payload.Error = false
	payload.Message = "Logged via GRPC!"

	app.writeJSON(w, http.StatusAccepted, payload)

}
