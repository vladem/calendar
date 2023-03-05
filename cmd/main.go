package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/vladem/calendar/service"
)

func main() {
	service := service.Service{}
	err := service.ServeHttp()
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	interrupted := make(chan os.Signal, 1)
	signal.Notify(interrupted, os.Interrupt)
	<-interrupted
	log.Println("received interrupt stopping")
	if err := service.StopServing(); err != nil {
		log.Fatalf("error on shutdown: %v", err)
	}
}
