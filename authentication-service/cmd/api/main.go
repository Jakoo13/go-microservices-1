package main

import (
	"authentication/data"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/jackc/pgconn"
)

const webPort = "80"

var counts int64

type Config struct {
	DB *sql.DB
	Models data.Models
}

func main(){
	log.Println("starting authentication service")

	// connect to DB
	conn := connectToDB()
	if conn == nil {
		log.Panic("could not connect to db")
	}

	// set up config
	app := Config{
		DB: conn,
		Models: data.New(conn),
	}

	// set up Web Server
	srv := &http.Server{
		Addr: fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	// start Web Server
	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func openDB(dsn string) (*sql.DB, error){
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	// check connection
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Connect to DB
func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")

	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Println("could not connect to db:", err)
			counts ++
		} else {
			log.Println("connected to Postgres")
			return connection
		}

		if counts > 10 {
			log.Println("could not connect to db after 10 tries")
			return nil
		}

		log.Println("Backing off for 2 seconds")
		time.Sleep(2 * time.Second)
		continue
	}
}
