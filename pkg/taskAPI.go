package pkg

import (
	"log"
	"time"

	taskV1 "google.golang.org/api/tasks/v1"
	"google.golang.org/protobuf/proto"
)

func ListTaskLists(tasksService *taskV1.Service) {
	taskLists, err := tasksService.Tasklists.List().Do()
	if err != nil {
		log.Fatalf("Unable to retrieve task lists: %v", err)
	}
	for _, taskList := range taskLists.Items {
		log.Printf("%s) %s", taskList.Id, taskList.Title)
	}
}

func ListAllTasks(tasksService *taskV1.Service) {
	taskLists, err := tasksService.Tasklists.List().Do()
	if err != nil {
		log.Fatalf("Unable to retrieve task lists: %v", err)
	}
	for _, taskList := range taskLists.Items {
		tasks, err := tasksService.Tasks.List(taskList.Id).Do()
		if err != nil {
			log.Fatalf("Unable to retrieve tasks: %v", err)
		}
		for _, task := range tasks.Items {
			log.Printf("%s) Title: %s Note: %s --- tasklist.Id: %s", task.Id, task.Title, task.Notes, taskList.Id)
		}
	}
}

func ListTasks(tasksService *taskV1.Service, taskList string) []*taskV1.Task {
	tasks, err := tasksService.Tasks.List(taskList).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve tasks: %v", err)
	}
	var retTasks []*taskV1.Task
	for _, task := range tasks.Items {
		if task.Completed == nil {
			continue
		}
		retTasks = append(retTasks, task)
		log.Printf("%s) Title: %s Note: %s", task.Id, task.Title, task.Notes)
	}

	return retTasks
}

func GetTask(tasksService *taskV1.Service, taskList string, taskID string) *taskV1.Task {
	taskListItem, err := tasksService.Tasklists.Get(taskList).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve task lists: %v", err)
	}
	task, err := tasksService.Tasks.Get(taskListItem.Id, taskID).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve task: %v", err)
	}

	log.Printf("Retrieved task: %s", task.Title)

	return task
}

func CreateTask(tasksService *taskV1.Service, taskList string, title string, note string) *taskV1.Task {
	taskListItem, err := tasksService.Tasklists.Get(taskList).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve task lists: %v", err)
	}
	task := &taskV1.Task{
		Title: title,
		Notes: note,
	}
	createdTask, err := tasksService.Tasks.Insert(taskListItem.Id, task).Do()
	if err != nil {
		log.Fatalf("Unable to create task: %v", err)
	}

	log.Printf("Created task: %s", createdTask.Title)

	return createdTask
}

func UpdateTask(tasksService *taskV1.Service, taskList string, taskID string, title string, note string, completed bool) *taskV1.Task {
	taskListItem, err := tasksService.Tasklists.Get(taskList).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve task lists: %v", err)
	}
	task, err := tasksService.Tasks.Get(taskListItem.Id, taskID).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve task: %v", err)
	}
	var completedDate *string
	status := "needsAction"
	hidden := false

	if completed {
		completedDate = proto.String(time.Now().Local().Format(time.RFC3339))
		status = "completed"
		hidden = true
	}
	if title == "" {
		title = task.Title
	}
	if note == "" {
		note = task.Notes
	}
	updatedTask, err := tasksService.Tasks.Update(taskListItem.Id, taskID, &taskV1.Task{
		Id:        taskID,
		Title:     title,
		Notes:     note,
		Completed: completedDate,
		Status:    status,
		Hidden:    hidden,
	}).Do()
	if err != nil {
		log.Fatalf("Unable to update task: %v", err)
	}

	log.Printf("Updated task: %s", updatedTask.Title)

	return updatedTask
}

func DeleteTask(tasksService *taskV1.Service, taskList string, taskID string) {
	taskListItem, err := tasksService.Tasklists.Get(taskList).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve task lists: %v", err)
	}

	err = tasksService.Tasks.Delete(taskListItem.Id, taskID).Do()
	if err != nil {
		log.Fatalf("Unable to delete task: %v", err)
	}

	log.Printf("Deleted task: %s", taskID)
}
