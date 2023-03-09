build:
	docker build . -t local/api

run:
	docker-compose up

clean:
	rm -rf data/

test:
	docker-compose up test
