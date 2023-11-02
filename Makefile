run:
	go run cmd/timetable/*.go

build:
	export CGO_ENABLED=0
	go build -o bin/ttable cmd/timetable/*.go
