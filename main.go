package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
)

const (
	ColorReset = "\033[0m"
	ColorGreen = "\033[32m"
	ColorRed   = "\033[31m"
)

type Task struct {
	Id          int
	Description string
	Done        bool
}

type TasksList struct {
	Tasks []Task
}

func (tl *TasksList) NextID() int {
	if len(tl.Tasks) == 0 {
		return 1
	}

	maxID := 0

	for _, task := range tl.Tasks {
		if task.Id > maxID {
			maxID = task.Id
		}
	}

	return maxID + 1
}

func (tl *TasksList) Append(task Task) {
	tl.Tasks = append(tl.Tasks, task)
}

func (tl *TasksList) Persist() {
	jsonData, err := json.MarshalIndent(tl, "", "  ")

	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	file, err := os.Create("tasks.json")

	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}

	defer file.Close()

	_, err = file.Write(jsonData)

	if err != nil {
		fmt.Println("Error writing to file:", err)
	}

	fmt.Println("Changes saved!")
}
func (tl *TasksList) FindIndex(id int) (int, error) {
	for idx, task := range tl.Tasks {
		if task.Id == id {
			return idx, nil
		}
	}

	return 0, errors.New("invalid id")
}

func (tl *TasksList) Find(id int) (*Task, error) {
	idx, err := tl.FindIndex(id)
	if err != nil {
		return nil, err
	}

	return &tl.Tasks[idx], nil
}

func (tl *TasksList) Remove(id int) error {
	idx, err := tl.FindIndex(id)
	if err != nil {
		return err
	}

	tl.Tasks = append(tl.Tasks[:idx], tl.Tasks[idx+1:]...)

	return nil
}

func NewTasksList() TasksList {
	return TasksList{Tasks: make([]Task, 0)}
}

func InitTasks() TasksList {
	jsonFile, err := os.Open("tasks.json")

	if err != nil {
		return TasksList{Tasks: make([]Task, 0)}
	}

	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)

	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return NewTasksList()
	}

	var tasks TasksList

	err = json.Unmarshal(byteValue, &tasks)

	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)

		return NewTasksList()
	}

	return tasks
}

type Executable interface {
	Execute(args []string, tasks *TasksList)
}

type AddCommand struct{}

func (add AddCommand) Execute(args []string, tasks *TasksList) {
	if len(args) < 1 {
		fmt.Println("Missing task description")
		os.Exit(1)
	}

	task := Task{Description: args[0], Done: false, Id: tasks.NextID()}

	tasks.Append(task)

	tasks.Persist()
}

type ListCommand struct{}

func (list ListCommand) Execute(_ []string, tasks *TasksList) {
	if len(tasks.Tasks) == 0 {
		fmt.Println("No tasks found!")
		return
	}

	for _, task := range tasks.Tasks {
		if task.Done {
			fmt.Printf("%s%v: %v - ✓ Completed%s\n",
				ColorGreen, task.Id, task.Description, ColorReset)
		} else {
			fmt.Printf("%s%v: %v - ⏳ Pending%s\n",
				ColorRed, task.Id, task.Description, ColorReset)
		}
	}
}

type CompleteCommand struct{}

func (complete CompleteCommand) Execute(args []string, tasks *TasksList) {
	if len(args) < 1 {
		fmt.Println("Missing task id")
		os.Exit(1)
	}

	id, err := strconv.Atoi(args[0])

	if err != nil {
		fmt.Printf("Invalid task ID: %s (must be a number)\n", args[0])
		os.Exit(1)
	}

	task, err := tasks.Find(id)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	task.Done = true

	tasks.Persist()
}

type RemoveCommand struct{}

func (remove RemoveCommand) Execute(args []string, tasks *TasksList) {
	if len(args) < 1 {
		fmt.Println("Missing task ID")
		os.Exit(1)
	}

	id, err := strconv.Atoi(args[0])

	if err != nil {
		fmt.Printf("Invalid task ID: %s (must be a number)\n", args[0])
		os.Exit(1)
	}

	err = tasks.Remove(id)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	tasks.Persist()
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: tasks <command> [args]")
		fmt.Println("Commands: add, list, complete, remove")
		os.Exit(1)
	}

	commands := map[string]Executable{
		"add":      AddCommand{},
		"list":     ListCommand{},
		"complete": CompleteCommand{},
		"remove":   RemoveCommand{},
	}

	command, exists := commands[os.Args[1]]

	if !exists {
		fmt.Println("invalid command")
		os.Exit(1)
	}

	tasks := InitTasks()

	command.Execute(os.Args[2:], &tasks)
}
