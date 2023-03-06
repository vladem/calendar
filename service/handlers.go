package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var dateLayout = "2006-01-02T15:04:05Z07:00"

func (s *Service) AddUser(w http.ResponseWriter, r *http.Request) {
	coll := s.DbClient.Database("db").Collection("users")
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res, err := coll.InsertOne(context.TODO(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		user.Id = oid.Hex()
	}
	json.NewEncoder(w).Encode(user)
	w.WriteHeader(http.StatusOK)
}

func (s *Service) AddMeeting(w http.ResponseWriter, r *http.Request) {
	var meeting Meeting
	err := json.NewDecoder(r.Body).Decode(&meeting)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logins := []string{meeting.Owner}
	for _, invite := range meeting.Invited {
		logins = append(logins, invite.Invitee)
	}
	if !s.checkUsersExist(logins, w) {
		return
	}
	meeting.StartTime = meeting.StartTime.Truncate(60 * time.Second)
	// todo: check that there is no intersection
	res, err := s.DbClient.Database("db").Collection("meetings").InsertOne(context.TODO(), meeting)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		meeting.Id = oid.Hex()
	}
	json.NewEncoder(w).Encode(meeting)
	w.WriteHeader(http.StatusOK)
}

func (s *Service) GetMeeting(w http.ResponseWriter, r *http.Request) {
	meetingId := mux.Vars(r)["id"]
	var meeting Meeting
	objectId, err := primitive.ObjectIDFromHex(meetingId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	filter := bson.D{{"_id", objectId}}
	err = s.DbClient.Database("db").Collection("meetings").FindOne(context.TODO(), filter).Decode(&meeting)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(meeting)
	w.WriteHeader(http.StatusOK)
}

func (s *Service) ListMeetings(w http.ResponseWriter, r *http.Request) {
	login := mux.Vars(r)["login"]
	startTime, err := time.Parse(dateLayout, mux.Vars(r)["startTime"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	endTime, err := time.Parse(dateLayout, mux.Vars(r)["endTime"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	startTime = startTime.Truncate(60 * time.Second)
	endTime = endTime.Truncate(60 * time.Second)

	schedule, err := MakeSchedule(s.DbClient.Database("db").Collection("meetings"), []string{login}, &startTime, &endTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	meetings := []Meeting{}
	for schedule.HasNext() {
		meeting, err := schedule.Next()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		meetings = append(meetings, *meeting)
	}
	json.NewEncoder(w).Encode(meetings)
	w.WriteHeader(http.StatusOK)
}

func (s *Service) FindSlot(w http.ResponseWriter, r *http.Request) {
	duration, err := strconv.Atoi(mux.Vars(r)["durationMinutes"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logins := strings.Split(mux.Vars(r)["logins"], ",")
	if !s.checkUsersExist(logins, w) {
		return
	}
	startTime, err := time.Parse(dateLayout, mux.Vars(r)["startTime"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	schedule, err := MakeSchedule(s.DbClient.Database("db").Collection("meetings"), logins, &startTime, nil) // todo: can retrieve less data from db, use projection
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	prevMeetingEnd := startTime
	for schedule.HasNext() {
		meeting, err := schedule.Next()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if prevMeetingEnd.Before(meeting.StartTime) && meeting.StartTime.Sub(prevMeetingEnd) >= time.Duration(duration)*time.Minute {
			break
		}
		if prevMeetingEnd.Before(meeting.EndTime) {
			prevMeetingEnd = meeting.EndTime
		}
	}
	json.NewEncoder(w).Encode(map[string]string{
		"startTime": prevMeetingEnd.Format(dateLayout),
	})
	w.WriteHeader(http.StatusOK)
}

func (s *Service) checkUsersExist(logins []string, w http.ResponseWriter) bool {
	filter := bson.D{{"login", bson.D{{"$in", logins}}}}
	count, err := s.DbClient.Database("db").Collection("users").CountDocuments(context.TODO(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}
	if count != int64(len(logins)) {
		http.Error(w, fmt.Sprintf("invalid owner or invitees %v/%v", count, len(logins)), http.StatusBadRequest)
		return false
	}
	return true
}
