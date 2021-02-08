package main

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/logrusorgru/aurora"
)

type Task struct {
	Job      *Job
	ResQueue chan<- Res
}

func Runner(id int, wg *sync.WaitGroup, taskQueue <-chan *Task, canExit *bool) {
	defer wg.Done()

	fmt.Printf("Runner %d created\n", id)
	for task := range taskQueue {
		task.ResQueue <- task.Job.Run()
	}
	fmt.Printf("Runner %d end\n", id)

}

type Res struct {
	Res string
}

//= Job =================================================================
type Job struct {
	ID int
}

func (j *Job) Run() Res {
	return Res{fmt.Sprintf("%s-%d done", reflect.TypeOf(j), j.ID)}
}

func (j Job) String() string {
	return aurora.Sprintf(aurora.Red("Job-%d"), j.ID)
}

//= main() ===============================================================
func main() {

	canExit := false
	taskQueue := make(chan *Task, 1000000)
	resQueue := make(chan Res)

	// Create Runners
	wgRunners := &sync.WaitGroup{}
	for i := 0; i < 32; i++ {
		wgRunners.Add(1)
		go Runner(i, wgRunners, taskQueue, &canExit)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		count := 0
		for res := range resQueue {
			count++
			fmt.Printf("Res-A : %s\n", res)

			fmt.Println("Count :", count)
		}
	}()

	func() {
		count := 0
		for i := 0; i < 1000000; i++ {
			fmt.Printf("Push job %d\n", i)

			taskQueue <- &Task{Job: &Job{ID: i}, ResQueue: resQueue}
			count++
		}
		fmt.Printf("All %d tasks created, closing channel\n", count)
		close(taskQueue)
	}()

	wgRunners.Wait()
	close(resQueue)
	wg.Wait()
}
