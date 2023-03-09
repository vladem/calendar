package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vladem/calendar/service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	url        = "http://api:8080"
	dbClient   *mongo.Client
	setupError error
	client     *CalendarClient
)

func TestMain(m *testing.M) {
	dbClient, setupError = service.ConnectDb()
	client = &CalendarClient{url}
	for i := 0; i < 10 && client.Ping() != nil; i++ {
		time.Sleep(time.Second)
	}
	os.Exit(m.Run())
}

func cleanup(t *testing.T) {
	require.Empty(t, setupError)
	database := dbClient.Database("db")
	_, err := database.Collection("users").DeleteMany(context.TODO(), bson.M{})
	require.Empty(t, err)
	_, err = database.Collection("meetings").DeleteMany(context.TODO(), bson.M{})
	require.Empty(t, err)
}

func createMeetings(t *testing.T) {
	require.Empty(t, client.PostUser("bob"))
	require.Empty(t, client.PostUser("alice"))
	meeting := service.Meeting{
		Owner:       "bob",
		Invited:     []service.Invitation{{Invitee: "alice"}},
		StartTime:   parseTimeNoError(t, "2023-03-07T16:20:00.000Z"),
		EndTime:     parseTimeNoError(t, "2023-03-07T16:40:00.000Z"),
		Reoccurance: 1,
	}
	_, err := client.PostMeeting(meeting)
	require.Empty(t, err)
	meeting.StartTime = parseTimeNoError(t, "2023-03-07T17:00:00.000Z")
	meeting.EndTime = parseTimeNoError(t, "2023-03-07T17:30:00.000Z")
	meeting.Reoccurance = 0
	_, err = client.PostMeeting(meeting)
	require.Empty(t, err)
	meeting.StartTime = parseTimeNoError(t, "2023-03-07T20:00:00.000Z")
	meeting.EndTime = parseTimeNoError(t, "2023-03-07T20:30:00.000Z")
	meeting.Reoccurance = 0
	_, err = client.PostMeeting(meeting)
	require.Empty(t, err)
}

func TestListMeetings(t *testing.T) {
	cleanup(t)
	createMeetings(t)
	meetings, err := client.ListMeetings("alice", "2023-03-07T16:00:00.000Z", "2023-03-07T19:00:00.000Z")
	require.Empty(t, err)
	require.Equal(t, 2, len(meetings))
	meetings, err = client.ListMeetings("bob", "2023-03-08T16:00:00.000Z", "2023-03-08T16:30:00.000Z")
	require.Empty(t, err)
	require.Equal(t, 1, len(meetings))
	meetings, err = client.ListMeetings("bob", "2023-03-08T15:00:00.000Z", "2023-03-08T16:00:00.000Z")
	require.Empty(t, err)
	require.Equal(t, 0, len(meetings))
}

func TestFindSlot(t *testing.T) {
	cleanup(t)
	createMeetings(t)
	slotStartTime, err := client.FindSlot([]string{"alice", "bob"}, "2023-03-08T16:00:00.000Z", 30)
	require.Empty(t, err)
	require.Equal(t, "2023-03-08T16:40:00Z", slotStartTime)
	slotStartTime, err = client.FindSlot([]string{"alice", "bob"}, "2023-03-07T15:50:00.000Z", 30)
	require.Empty(t, err)
	require.Equal(t, "2023-03-07T15:50:00Z", slotStartTime)
	slotStartTime, err = client.FindSlot([]string{"alice", "bob"}, "2023-03-07T15:51:00.000Z", 30)
	require.Empty(t, err)
	require.Equal(t, "2023-03-07T17:30:00Z", slotStartTime)
}
