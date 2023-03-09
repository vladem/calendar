build:
	docker build . -t local/api

run:
	docker-compose up

clean:
	rm -rf data/ && docker-compose down

test:
	docker-compose up test
