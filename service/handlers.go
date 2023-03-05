package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
	filter := bson.D{{"login", bson.D{{"$in", logins}}}}
	count, err := s.DbClient.Database("db").Collection("users").CountDocuments(context.TODO(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if count != int64(len(logins)) {
		http.Error(w, fmt.Sprintf("invalid owner or invitees %v/%v", count, len(logins)), http.StatusBadRequest)
		return
	}

	meeting.StartTime = meeting.StartTime.Truncate(60 * time.Second)
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
	startTime, err := time.Parse("2006-01-02T15:04:05Z07:00", mux.Vars(r)["startTime"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	endTime, err := time.Parse("2006-01-02T15:04:05Z07:00", mux.Vars(r)["endTime"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	startTime = startTime.Truncate(60 * time.Second)
	endTime = endTime.Truncate(60 * time.Second)

	meetings, err := s.meetingsForUsers([]string{login}, startTime, endTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(meetings)
	w.WriteHeader(http.StatusOK)
}

func (s *Service) FindSlot(w http.ResponseWriter, r *http.Request) {
	login := mux.Vars(r)["login"]
	meetings := []Meeting{}
	filter := bson.D{{"invited.invitee", login}}
	cursor, err := s.DbClient.Database("db").Collection("meetings").Find(context.TODO(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = cursor.All(context.TODO(), &meetings); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(meetings)
	w.WriteHeader(http.StatusOK)
}

func (s *Service) meetingsForUsers(logins []string, startTime, endTime time.Time) ([]Meeting, error) {
	meetings := []Meeting{}
	filter := bson.D{
		{"$and",
			bson.A{
				bson.D{{"$or", bson.A{
					bson.D{{"owner", bson.D{{"$in", logins}}}},
					bson.D{{"invited.invitee", bson.D{{"$in", logins}}}},
				}}},
				bson.D{{"$or", bson.A{
					bson.D{{"$and", bson.A{
						bson.D{{"startTime", bson.D{{"$gte", startTime}}}},
						bson.D{{"startTime", bson.D{{"$lt", endTime}}}},
					}}},
					bson.D{{"$and", bson.A{
						bson.D{{"endTime", bson.D{{"$gt", startTime}}}},
						bson.D{{"endTime", bson.D{{"$lte", endTime}}}},
					}}},
				}}},
			}},
	}
	cursor, err := s.DbClient.Database("db").Collection("meetings").Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(context.TODO(), &meetings); err != nil {
		return nil, err
	}
	return meetings, nil
}
