package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Debianov/calc-ya-go-24/backend"
	"io"
	"log"
	"net/http"
)

type Agent struct {
	ServerURL string // запись данных полей производится один раз, редактирование не предусматривается. Синхронизация
	// горутин не требуется.
	getEndpoint  string
	sendEndpoint string
}

func (a *Agent) get() (result *backend.Task, ok bool) {
	var err error
	resp, err := http.Get(a.ServerURL + a.getEndpoint)
	if err != nil {
		log.Panic(err)
	}
	if resp.StatusCode != http.StatusOK {
		ok = false
		return
	} else {
		ok = true
	}
	var (
		reqBuf = make([]byte, resp.ContentLength)
	)
	_, err = resp.Body.Read(reqBuf)
	if err != nil && err != io.EOF {
		log.Panic(err)
	}
	var structInResp backend.TaskToSend
	err = json.Unmarshal(reqBuf, &structInResp)
	if err != nil {
		log.Panic(err)
	}
	result = structInResp.Task
	return
}

func (a *Agent) calc(task backend.Task) (agentResult backend.AgentResult, err error) {
	var result float64
	agentResult = backend.AgentResult{
		ID: task.PairID,
	}
	switch task.Operation {
	case "+":
		result = task.Arg1.(float64) + task.Arg2.(float64)
	case "-":
		result = task.Arg1.(float64) - task.Arg2.(float64)
	case "*":
		result = task.Arg1.(float64) * task.Arg2.(float64)
	case "/":
		result = task.Arg1.(float64) / task.Arg2.(float64)
	default:
		err = errors.New("неизвестная операция")
		return
	}
	agentResult.Result = int64(result)
	return
}

func (a *Agent) send(agentResult backend.AgentResult) (err error) {
	reqBuf, err := json.Marshal(agentResult)
	if err != nil {
		return
	}
	resp, err := http.Post(a.ServerURL+a.sendEndpoint, "application/json", bytes.NewReader(reqBuf))
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("результат ID %d не записан, код: %d", agentResult.ID, resp.StatusCode)
		return
	}
	return
}
