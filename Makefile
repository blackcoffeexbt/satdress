all: satdress satdress-cli

satdress:
	go build

satdress-cli:
	go build ./cli/satdress-cli.go

clean:
	rm satdress-cli
	go clean
