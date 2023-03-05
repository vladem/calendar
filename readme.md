# Calendar

## Usage
```
make build
make up

curl -X POST http://127.0.0.1:8080/api/users -d '{"login": "bob"}' -H "Content-Type: application/json"
curl -X POST http://127.0.0.1:8080/api/users -d '{"login": "alice"}' -H "Content-Type: application/json"

data='{"owner": "bob", "invited": [{"invitee": "alice"}], "startTime": "2023-03-07T16:20:00.000Z","endTime": "2023-03-07T16:40:00.000Z","reoccurance": 0,"description": "blabla"}'
curl -X POST http://127.0.0.1:8080/api/meetings -d $data -H "Content-Type: application/json"
data='{"owner": "bob", "invited": [{"invitee": "alice"}], "startTime": "2023-03-07T17:00:00.000Z","endTime": "2023-03-07T17:30:00.000Z","reoccurance": 0,"description": "blabla"}'
curl -X POST http://127.0.0.1:8080/api/meetings -d $data -H "Content-Type: application/json"
data='{"owner": "bob", "invited": [{"invitee": "alice"}], "startTime": "2023-03-07T20:00:00.000Z","endTime": "2023-03-07T20:30:00.000Z","reoccurance": 0,"description": "blabla"}'
curl -X POST http://127.0.0.1:8080/api/meetings -d $data -H "Content-Type: application/json"

curl 'http://127.0.0.1:8080/api/users/alice/meetings?startTime=2023-03-07T16:00:00.000Z&endTime=2023-03-07T19:00:00.000Z'
curl 'http://127.0.0.1:8080/api/users/bob/meetings?startTime=2023-03-07T16:00:00.000Z&endTime=2023-03-07T19:00:00.000Z'
curl 'http://127.0.0.1:8080/api/users/bob/meetings?startTime=2023-03-07T16:00:00.000Z&endTime=2023-03-07T20:10:00.000Z'
```
