package service

import (
	"container/heap"
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Schedule struct {
	meetingsQueue PriorityQueue
	cursor        *mongo.Cursor
	nextInCursor  *Meeting
	endTime       *time.Time
	err           error
}

func MakeSchedule(coll *mongo.Collection, logins []string, startTime, endTime *time.Time) (*Schedule, error) {
	participantFilter := bson.D{{"$or", bson.A{
		bson.D{{"owner", bson.D{{"$in", logins}}}},
		bson.D{{"invited.invitee", bson.D{{"$in", logins}}}},
	}}}
	meetingsQueue := PriorityQueue{}
	cursor, err := coll.Find(context.TODO(), bson.D{
		{"$and",
			bson.A{
				participantFilter,
				bson.D{{"startTime", bson.D{{"$lt", startTime}}}},
				bson.D{{"reoccurance", bson.D{{"$ne", NoReoccurence}}}},
			}},
	})
	if err = cursor.All(context.TODO(), &meetingsQueue); err != nil {
		return nil, err
	}
	for i := 0; i < len(meetingsQueue); i++ {
		meetingsQueue[i] = meetingsQueue[i].NextOccurence(startTime)
	}
	heap.Init(&meetingsQueue)
	if endTime != nil {
		cursor, err = coll.Find(context.TODO(), bson.D{
			{"$and",
				bson.A{
					participantFilter,
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
		})
	} else {
		cursor, err = coll.Find(context.TODO(), bson.D{
			{"$and",
				bson.A{
					participantFilter,
					bson.D{{"$or", bson.A{
						bson.D{{"$and", bson.A{
							bson.D{{"startTime", bson.D{{"$gte", startTime}}}},
						}}},
						bson.D{{"$and", bson.A{
							bson.D{{"endTime", bson.D{{"$gt", startTime}}}},
						}}},
					}}},
				}},
		})
	}
	if err != nil {
		return nil, err
	}
	return &Schedule{
		meetingsQueue: meetingsQueue,
		cursor:        cursor,
		endTime:       endTime,
	}, nil
}

func (s *Schedule) HasNext() bool {
	hadNext := true
	if s.nextInCursor == nil {
		hadNext = s.cursor.Next(context.TODO())
		if hadNext {
			s.nextInCursor = &Meeting{}
			s.err = s.cursor.Decode(s.nextInCursor)
		}
	}
	return hadNext || len(s.meetingsQueue) != 0
}

func (s *Schedule) Next() (*Meeting, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.nextInCursor == nil || (len(s.meetingsQueue) != 0 && s.meetingsQueue[0].StartTime.Before(s.nextInCursor.StartTime)) {
		nextMeeting := heap.Pop(&s.meetingsQueue).(*Meeting)
		log.Printf("nextMeeting: %v", nextMeeting.StartTime)
		if s.endTime != nil {
			log.Printf("endTime: %v", *s.endTime)
		}
		nextOccurance := nextMeeting.NextOccurence(nil)
		if s.endTime == nil || nextOccurance.StartTime.Before(*s.endTime) {
			heap.Push(&s.meetingsQueue, nextOccurance)
		}
		return nextMeeting, nil
	}
	nextMeeting := s.nextInCursor
	if s.nextInCursor.Reoccurance != NoReoccurence {
		heap.Push(&s.meetingsQueue, s.nextInCursor.NextOccurence(nil))
	}
	s.nextInCursor = nil
	return nextMeeting, nil
}

func (m *Meeting) NextOccurence(startingFrom *time.Time) *Meeting {
	switch m.Reoccurance {
	case Daily:
		next := *m
		next.Id = ""
		if startingFrom == nil {
			next.StartTime = next.StartTime.Add(24 * time.Hour)
			next.EndTime = next.EndTime.Add(24 * time.Hour)
		} else {
			// todo: meeting can start and end in different days
			next.StartTime = time.Date(
				startingFrom.Year(),
				startingFrom.Month(),
				startingFrom.Day(),
				next.StartTime.Hour(),
				next.StartTime.Minute(),
				0,
				0,
				time.UTC)
			next.EndTime = time.Date(
				startingFrom.Year(),
				startingFrom.Month(),
				startingFrom.Day(),
				next.EndTime.Hour(),
				next.EndTime.Minute(),
				0,
				0,
				time.UTC)
			if next.StartTime.Before(*startingFrom) {
				next.StartTime = next.StartTime.Add(24 * time.Hour)
				next.EndTime = next.EndTime.Add(24 * time.Hour)
			}
		}
		return &next
	default:
		log.Fatalf("invalid reoccurance")
	}
	return nil
}

type PriorityQueue []*Meeting

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].StartTime.Before(pq[j].StartTime)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x any) {
	item := x.(*Meeting)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*pq = old[0 : n-1]
	return item
}
