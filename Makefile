


gen-fakes:
	go get github.com/maxbrunsfeld/counterfeiter/v6
	go generate ./...


test:
	go test -v ./...


