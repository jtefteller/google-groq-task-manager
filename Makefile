build:
	go build -o bin/google-groq-task-manager main.go

install:
	sh scripts/install.sh

uninstall:
	sh scripts/uninstall.sh

list-all-tasks:
	go run main.go -a all

list-all-taskLists:
	go run main.go -a listTaskList