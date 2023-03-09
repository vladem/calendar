package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vladem/calendar/service"
)

type CalendarClient struct {
	Endpoint string
}

func (c *CalendarClient) PostUser(login string) error {
	uri := c.Endpoint + "/api/users"
	reqBody, _ := json.Marshal(map[string]string{
		"login": login,
	})
	request, err := http.NewRequest("POST", uri, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		body, _ := ioutil.ReadAll(response.Body)
		return fmt.Errorf("PostUser, bad response %v %v", response.Status, body)
	}
	return nil
}

func (c *CalendarClient) PostMeeting(meeting service.Meeting) (string, error) {
	uri := c.Endpoint + "/api/meetings"
	reqBody, _ := json.Marshal(meeting)
	request, err := http.NewRequest("POST", uri, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		body, _ := ioutil.ReadAll(response.Body)
		return "", fmt.Errorf("PostUser, bad response %s %s", response.Status, body)
	}
	err = json.NewDecoder(response.Body).Decode(&meeting)
	if err != nil {
		body, _ := ioutil.ReadAll(response.Body)
		return "", fmt.Errorf("PostUser, bad response %s %s", response.Status, body)
	}
	return meeting.Id, nil
}

func (c *CalendarClient) ListMeetings(login string, startTime, endTime string) ([]service.Meeting, error) {
	uri := fmt.Sprintf("%s/api/users/%s/meetings?startTime=%s&endTime=%s", c.Endpoint, login, startTime, endTime)
	response, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		body, _ := ioutil.ReadAll(response.Body)
		return nil, fmt.Errorf("ListMeetings, bad response %s %s", response.Status, body)
	}
	meetings := []service.Meeting{}
	err = json.NewDecoder(response.Body).Decode(&meetings)
	if err != nil {
		body, _ := ioutil.ReadAll(response.Body)
		return nil, fmt.Errorf("ListMeetings, bad response %s %s", response.Status, body)
	}
	return meetings, nil
}

func (c *CalendarClient) FindSlot(logins []string, startTime string, durationMinutes int) (string, error) {
	uri := fmt.Sprintf("%s/api/findSlot?startTime=%s&durationMinutes=%d&logins=%s", c.Endpoint, startTime, durationMinutes, strings.Join(logins, ","))
	response, err := http.Get(uri)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		body, _ := ioutil.ReadAll(response.Body)
		return "", fmt.Errorf("FindSlot, bad response %s %s", response.Status, body)
	}
	slot := map[string]string{}
	err = json.NewDecoder(response.Body).Decode(&slot)
	if err != nil {
		body, _ := ioutil.ReadAll(response.Body)
		return "", fmt.Errorf("FindSlot, bad response %s %s", response.Status, body)
	}
	return slot["startTime"], nil
}

func (c *CalendarClient) Ping() error {
	_, err := http.Get(c.Endpoint)
	return err
}

func parseTimeNoError(t *testing.T, timestr string) time.Time {
	dateLayout := "2006-01-02T15:04:05Z07:00"
	res, err := time.Parse(dateLayout, timestr)
	require.Empty(t, err)
	return res
}
