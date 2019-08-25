
all: build
build:
	GOOS=windows GOARCH=amd64 go build -o ./dist/vrc_nitro_process.exe ./cli/
