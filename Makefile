build:
	mkdir -p .bin
	rm .bin/envgen.*

	GOOS=linux GOARCH=386 go build -o .bin/envgen.linux-386 .
	cd .bin && tar -czf envgen.linux-386.tar.gz envgen.linux-386
	rm .bin/envgen.linux-386
	
	GOOS=linux GOARCH=amd64 go build -o .bin/envgen.linux-amd64 .
	cd .bin && tar -czf envgen.linux-amd64.tar.gz envgen.linux-amd64
	rm .bin/envgen.linux-amd64
	
	GOOS=darwin GOARCH=386 go build -o .bin/envgen.darwin-386 .
	cd .bin && tar -czf envgen.darwin-386.tar.gz envgen.darwin-386
	rm .bin/envgen.darwin-386
	
	GOOS=darwin GOARCH=amd64 go build -o .bin/envgen.darwin-amd64 .
	cd .bin && tar -czf envgen.darwin-amd64.tar.gz envgen.darwin-amd64
	rm .bin/envgen.darwin-amd64
