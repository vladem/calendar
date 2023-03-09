package service

import (
	"context"
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

type Service struct {
	DbClient *mongo.Client
	Server   *http.Server
	StopWg   *sync.WaitGroup
}

func (s *Service) ServeHttp() error {
	if s.Server != nil {
		return errors.New("already serving")
	}
	var err error
	s.DbClient, err = ConnectDb()
	if err != nil {
		return err
	}
	r := mux.NewRouter()
	r.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		s.AddUser(w, r)
	}).Methods("POST")
	r.HandleFunc("/api/meetings", func(w http.ResponseWriter, r *http.Request) {
		s.AddMeeting(w, r)
	}).Methods("POST")
	r.HandleFunc("/api/meetings/{id}", func(w http.ResponseWriter, r *http.Request) {
		s.GetMeeting(w, r)
	}).Methods("GET")
	r.HandleFunc("/api/users/{login}/meetings", func(w http.ResponseWriter, r *http.Request) {
		s.ListMeetings(w, r)
	}).Methods("GET").Queries("startTime", "{startTime}").Queries("endTime", "{endTime}")
	r.HandleFunc("/api/findSlot", func(w http.ResponseWriter, r *http.Request) {
		s.FindSlot(w, r)
	}).Methods("GET").Queries("startTime", "{startTime}").Queries("durationMinutes", "{durationMinutes}").Queries("logins", "{logins}")
	r.HandleFunc("/api/acceptMeeting", func(w http.ResponseWriter, r *http.Request) {
		s.AcceptMeeting(w, r)
	}).Methods("POST")

	s.Server = &http.Server{Addr: ":8080", Handler: r}
	s.StopWg = &sync.WaitGroup{}
	s.StopWg.Add(1)
	go func() {
		defer s.StopWg.Done()
		if err := s.Server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()
	return nil
}

func (s *Service) StopServing() error {
	if s.Server == nil {
		return errors.New("already stopped")
	}
	if err := s.Server.Shutdown(context.TODO()); err != nil {
		return err
	}
	s.StopWg.Wait()
	return s.DbClient.Disconnect(context.TODO())
}
