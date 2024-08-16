package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/jtefteller/tasks/pkg"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/tasks/v1"
)

var homePath string

func init() {
	home := os.Getenv("HOME")
	projectName := ".google-groq-task-manager"
	homePath = home + "/" + projectName
}

func main() {
	loadEnv()
	tasksMain()
}

func tasksMain() {
	tasksService := authorize()
	llm := newLLM()

	createTask, deleteTask, listTasks, updateTask, allTasks, listTaskList,
		recommendation, taskID, taskList, taskName, note, prompt, completed := handleFlags(os.Args)

	if createTask {
		pkg.CreateTask(tasksService, taskList, taskName, note)
	} else if updateTask {
		pkg.UpdateTask(tasksService, taskList, taskID, taskName, note, completed)
	} else if deleteTask {
		pkg.DeleteTask(tasksService, taskList, taskID)
	} else if listTaskList {
		pkg.ListTaskLists(tasksService)
	} else if listTasks {
		pkg.ListTasks(tasksService, taskList)
	} else if allTasks {
		pkg.ListAllTasks(tasksService)
	} else if recommendation {
		var tasks []*tasks.Task
		if taskList != "" && taskID != "" {
			t := pkg.GetTask(tasksService, taskList, taskID)
			tasks = append(tasks, t)
		} else {
			ts := pkg.ListTasks(tasksService, taskList)
			tasks = ts
		}
		recs := pkg.Recommendation(llm, tasks, prompt)

		log.Print("Enter recommendations indeces separated by commas (no spaces): ")
		// var then variable name then variable type
		var userRecs string
		// Taking input from user
		fmt.Scanln(&userRecs)
		userRecs = strings.TrimSpace(userRecs)
		for _, idx := range strings.Split(userRecs, ",") {
			// Convert string to int
			intIdx, _ := strconv.Atoi(idx)
			rec := recs.Recommendations[intIdx]
			pkg.CreateTask(tasksService, taskList, rec.Title, rec.Notes)
		}
	} else {
		log.Fatalf("No action provided")
	}
}

func handleFlags(args []string) (bool, bool, bool, bool, bool, bool, bool, string, string, string, string, string, bool) {
	var (
		createTask     bool
		deleteTask     bool
		listTasks      bool
		updateTask     bool
		allTasks       bool
		listTaskList   bool
		recommendation bool
		action         string
		taskID         string
		taskList       string
		taskName       string
		note           string
		prompt         string
		completed      bool
	)

	flag.StringVar(&taskList, "lid", "", "Task list ID: @default, @me etc...")
	flag.StringVar(&action, "a", "", "Task action: list, create, update, delete, all, listTaskList, recommendation")
	flag.StringVar(&taskID, "tid", "", "Task ID")
	flag.StringVar(&taskName, "name", "", "Task name")
	flag.StringVar(&note, "note", "", "Task note")
	flag.StringVar(&prompt, "prompt", "", "Prompt for recommendation")
	flag.BoolVar(&completed, "completed", false, "Task completed")
	flag.Parse()

	if len(args) < 2 {
		flag.PrintDefaults()
		log.Fatalf("Not enough arguments provided")
	}

	if action == "" {
		flag.PrintDefaults()
		log.Fatalf("No action provided")
	}

	switch action {
	case "listTaskList":
		listTaskList = true
	case "list":
		if taskList == "" {
			flag.PrintDefaults()
			log.Fatalf("taskList is required")
		}
		listTasks = true

	case "create":
		createTask = true
		if taskList == "" || taskName == "" {
			flag.PrintDefaults()
			log.Fatalf("taskList, taskName are required")
		}
	case "update":
		updateTask = true
		if taskList == "" || taskID == "" {
			flag.PrintDefaults()
			log.Fatalf("taskList and taskID are required")
		}
	case "delete":
		deleteTask = true
		if taskList == "" || taskID == "" {
			flag.PrintDefaults()
			log.Fatalf("taskList and taskID are required")
		}
	case "all":
		allTasks = true
	case "recommendation":
		if prompt == "" || (taskList == "" && taskID == "") {
			flag.PrintDefaults()
			log.Fatalf("prompt and taskList or taskID are required")
		}
		recommendation = true
	default:
		log.Fatalf("Invalid action provided")
	}

	return createTask, deleteTask, listTasks, updateTask, allTasks, listTaskList, recommendation, taskID, taskList, taskName, note, prompt, completed
}

func loadEnv() {
	godotenv.Load(homePath + "/.env")
}

func newLLM() pkg.LLM {
	return pkg.NewGroq(os.Getenv("GROQ_API_KEY"))
}

func authorize() *tasks.Service {
	config := oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes:       []string{"https://www.googleapis.com/auth/tasks"},
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://accounts.google.com/o/oauth2/auth",
			TokenURL:  "https://oauth2.googleapis.com/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	storer := pkg.NewStorer(homePath + "/token.json")
	b := storer.Retrieve()
	if len(b) == 0 {
		wait := make(chan struct{})
		go localhostServer(wait, storer, &config)
		authCodeUrl := config.AuthCodeURL("state-token", oauth2.ApprovalForce, oauth2.AccessTypeOffline)
		log.Print(authCodeUrl)
		<-wait
		b = storer.Retrieve()
	}

	token, err := storer.ToToken(b)
	if err != nil {
		log.Fatalf("ToToken Error: %v", err)
	}
	ctx := context.Background()
	ts := config.TokenSource(ctx, &token)
	tasksService, err := tasks.NewService(ctx, option.WithScopes(tasks.TasksScope), option.WithTokenSource(ts))
	if err != nil {
		log.Fatalf("Unable to create Tasks service: %v", err)
	}

	return tasksService
}

func localhostServer(wait chan struct{}, storer pkg.Storer, config *oauth2.Config) {
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")
	if redirectURL == "" {
		log.Fatalf("GOOGLE_REDIRECT_URL is required")
	}
	regexp := regexp.MustCompile(`http://localhost:(\d+)`)
	matches := regexp.FindStringSubmatch(redirectURL)
	var addr string
	if len(matches) < 2 {
		addr = redirectURL
	} else {
		port := matches[1]
		addr = ":" + port
	}
	server := &http.Server{Addr: addr}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		values := r.URL.Query()
		code := values.Get("code")

		token, err := config.Exchange(context.Background(), code)
		if err != nil {
			log.Fatalf("Error: %v", err)
			return
		}
		storeToken := pkg.StoreToken{
			AccessToken:  token.AccessToken,
			ExpiresIn:    token.Expiry.Second(),
			RefreshToken: token.RefreshToken,
			Expiry:       token.Expiry,
			TokenType:    token.TokenType,
			Scope:        token.Extra("scope").(string),
		}

		storer.Store(storeToken)
		close(wait)
	})

	server.Handler = mux
	go server.ListenAndServe()
	<-wait
	server.Shutdown(context.Background())
}
