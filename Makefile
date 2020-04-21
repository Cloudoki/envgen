build:
	mkdir -p .bin
	GOOS=linux GOARCH=386 go build -o .bin/envgen.linux-386 .
	GOOS=linux GOARCH=amd64 go build -o .bin/envgen.linux-amd64 .
	GOOS=darwin GOARCH=386 go build -o .bin/envgen.darwin-386 .
	GOOS=darwin GOARCH=amd64 go build -o .bin/envgen.darwin-amd64 .
