package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	_ "gopkg.in/jackc/pgx.v4/stdlib"
)

type App struct {
	router *mux.Router
	db     *sql.DB
}

func (a *App) ConnectDb() {
	connectionString := "host=79.104.55.66 port=7002 user=user4 password=iN1A1 dbname=hr-test"
	var err error
	if a.db, err = sql.Open("pgx", connectionString); err != nil {
		log.Fatal(err)
	}
}

func (a *App) initializeRouter() {
	a.router = mux.NewRouter()
	a.router.HandleFunc("/users", a.GetUsersWithMinAge).Methods("GET").Queries("minAge", "{minAge}")
	a.router.HandleFunc("/users", a.GetAllUsers).Methods("GET")

	a.router.HandleFunc("/users/{id}", a.GetUserById).Methods("GET")
	a.router.HandleFunc("/users", a.GreateUser).Methods("POST")
	a.router.HandleFunc("/users/{id}", a.ChangeUser).Methods("PUT")
	a.router.HandleFunc("/users/{id}", a.DeleteUser).Methods("DELETE")
}

func (a *App) Init() {
	a.ConnectDb()
	a.initializeRouter()
}

func (a *App) Run() {
	srv := &http.Server{
		Addr:    ":8000",
		Handler: a.router,
	}
	serveChan := make(chan error, 1)
	go func() {
		serveChan <- srv.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-stop:
		fmt.Println("shutting down")
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	case err := <-serveChan:
		log.Fatal(err)
	}
}
