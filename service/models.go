package service

import "time"

type AcceptedChoice uint8

var (
	NotReviewed AcceptedChoice = 0
	Accepted    AcceptedChoice = 1
	Declined    AcceptedChoice = 2
)

type ReoccureanceChoice uint8

var (
	NoReoccurence ReoccureanceChoice = 0
	Daily         ReoccureanceChoice = 1
	WorkingDays   ReoccureanceChoice = 2
	Weekly        ReoccureanceChoice = 3
	Monthly       ReoccureanceChoice = 4
	Yearly        ReoccureanceChoice = 5
)

type Invitation struct {
	Invitee  string         `json:"invitee" bson:"invitee"` // todo: index
	Accepted AcceptedChoice `json:"accepted" bson:"accepted"`
}

type Meeting struct {
	Id          string             `json:"id,omitempty" bson:"_id,omitempty"`
	Owner       string             `json:"owner" bson:"owner"`
	Invited     []Invitation       `json:"invited" bson:"invited"`
	StartTime   time.Time          `json:"startTime" bson:"startTime"`
	EndTime     time.Time          `json:"endTime" bson:"endTime"`
	Reoccurance ReoccureanceChoice `json:"reoccurance" bson:"reoccurance"`
	Description string             `json:"description" bson:"description"`
}

type User struct {
	Id    string `json:"id,omitempty" bson:"_id,omitempty"`
	Login string `json:"login" bson:"login"` // todo: unique index
}

type AcceptMeetingRequest struct {
	MeetingId string `json:"meetingId"`
	Login     string `json:"login"`
	Decline   bool   `json:"decline"`
}
