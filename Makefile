
up:
	docker compose up -d --build

down:
	docker compose down

gen:
	go get github.com/maxbrunsfeld/counterfeiter/v6
	go generate ./...

test:
	go test -v ./...


