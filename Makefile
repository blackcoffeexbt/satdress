satdress: $(shell find . -name "*.go") $(shell find . -name "*.html") $(shell find . -name "*.css") go.mod
	CC=$$(which musl-gcc) go build -ldflags='-s -w -linkmode external -extldflags "-static"' -o ./satdress
