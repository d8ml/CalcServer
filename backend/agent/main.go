package main

import (
	"github.com/Debianov/calc-ya-go-24/backend"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

func main() {
	var (
		err   error
		agent = getDefaultAgent()
		wg    sync.WaitGroup
	)

	var numberCalcGoroutinesInString = os.Getenv("COMPUTING_POWER")
	numberCalcGoroutines, err := strconv.ParseInt(numberCalcGoroutinesInString, 10, 32)

	var (
		results          = make(chan backend.AgentResult, numberCalcGoroutines)
		tasksReadyToCalc = make(chan backend.Task, numberCalcGoroutines)
	)

	for range numberCalcGoroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case task := <-tasksReadyToCalc:
					agentResult, err := agent.calc(task)
					if err != nil {
						log.Println(err, task.PairID)
					}
					results <- agentResult
				}
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-time.After(30 * time.Millisecond):
				task, ok := agent.get()
				if ok {
					tasksReadyToCalc <- *task
				}
			}
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case result := <-results:
				err = agent.send(result)
				if err != nil {
					log.Println(err, result.ID)
				}
			}
		}
	}()
	wg.Wait()
}
