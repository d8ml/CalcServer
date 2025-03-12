package backend

import (
	"encoding/json"
	"errors"
	"github.com/Debianov/calc-ya-go-24/pkg"
	"go/types"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

type JsonPayload interface {
	Marshal() (result []byte, err error)
}

type RequestJson struct {
	Expression string `json:"expression"`
}

func (r RequestJson) Marshal() (result []byte, err error) {
	result, err = json.Marshal(&r)
	return
}

/*
RequestNilJson изначально нужен для передачи nil и вызова Internal Server Error. Мы передаём nil, затем
он извлекается через Expression для создания Reader, а этот Reader запихивается в http.Request и передаётся
дальше в функцию. Далее, функция вызовет панику, паника перехватится PanicMiddleware, и далее по списку.

Используется в тесте TestBadGetHandler.
*/
type RequestNilJson struct {
	Expression types.Type `json:"expression"`
}

func (r RequestNilJson) Marshal() (result []byte, err error) {
	return nil, nil
}

type OKJson struct {
	Result float64 `json:"result"`
}

func (o OKJson) Marshal() (result []byte, err error) {
	result, err = json.Marshal(&o)
	return
}

type ErrorJson struct {
	Error string `json:"error"`
}

func (e ErrorJson) Marshal() (result []byte, err error) {
	result, err = json.Marshal(&e)
	return
}

type EmptyJson struct {
}

func (e EmptyJson) Marshal() (result []byte, err error) {
	return
}

type ExprStatus string

const (
	Ready        ExprStatus = "Есть готовые задачи"
	NoReadyTasks            = "Нет готовых задач"
	Completed               = "Выполнено"
	Cancelled               = "Отменено"
)

const (
	TIME_ADDITION_MS        string = "TIME_ADDITION_MS"
	TIME_SUBTRACTION_MS            = "TIME_SUBTRACTION_MS"
	TIME_MULTIPLICATIONS_MS        = "TIME_MULTIPLICATIONS_MS"
	TIME_DIVISIONS_MS              = "TIME_DIVISIONS_MS"
)

type TaskToSend struct {
	Task              *Task `json:"task"`
	timeAtSendingTask time.Time
}

func (t *TaskToSend) Marshal() (result []byte, err error) {
	result, err = json.Marshal(&t)
	return
}

type Expression struct {
	postfix      []string
	ID           int        `json:"id"`
	Status       ExprStatus `json:"status"`
	Result       int64      `json:"result"`
	tasksHandler *Tasks
	mut          sync.Mutex
}

func (e *Expression) DivideIntoTasks() {
	var (
		operatorCount int
		stack         = pkg.StackFabric[int64]()
	)
	for _, r := range e.postfix { // TODO: сделать структуру в постфиксе, уже распарсенную. нам останется пройтись
		// TODO по ней слева направо и записать всё в порядке <оператор, операнд, операнд>.
		if pkg.IsNumber(r) {
			operandInInt, err := strconv.ParseInt(r, 10, 64)
			if err != nil {
				log.Panic(err)
			}
			stack.Push(operandInInt)
		} else if pkg.IsOperator(r) {
			var (
				newId   = e.generateId(operatorCount)
				newTask *Task
			)
			if stack.Len() >= 2 {
				newTask = &Task{PairID: newId, Arg2: stack.Pop(), Arg1: stack.Pop(),
					Operation: r, OperationTime: e.getOperationTime(r), Status: ReadyToCalc}
			} else if stack.Len() == 1 {
				newTask = &Task{PairID: newId, Arg2: stack.Pop(), Operation: r,
					OperationTime: e.getOperationTime(r), Status: WaitingOtherTasks}
			} else {
				newTask = &Task{PairID: newId, Operation: r, OperationTime: e.getOperationTime(r),
					Status: WaitingOtherTasks}
			}
			e.tasksHandler.add(newTask)
			operatorCount++
		}
	}
	return
}

func (e *Expression) generateId(operatorCount int) int {
	return pkg.Pair(e.ID, operatorCount)
}

func (e *Expression) getOperationTime(currentOperator string) (result time.Duration) {
	var (
		operatorAndEnvNamePairs = map[string]string{"+": TIME_ADDITION_MS, "-": TIME_SUBTRACTION_MS,
			"*": TIME_MULTIPLICATIONS_MS, "/": TIME_DIVISIONS_MS}
		maybeDuration string
		err           error
	)
	for operator, envName := range operatorAndEnvNamePairs {
		if currentOperator == operator {
			maybeDuration = os.Getenv(envName)
			if maybeDuration == "" {
				log.Printf("WARNING: переменная %s не обнаружена", envName)
			}
			result, err = time.ParseDuration(maybeDuration)
			if err != nil {
				log.Panic(err)
			}
		}
	}
	return
}

func (e *Expression) FabricReadyExprSendTask() TaskToSend {
	maybeReadyTask := e.tasksHandler.registerFirst()
	if maybeReadyTask.IsReadyToCalc() {
		if e.tasksHandler.Len() == 1 {
			e.changeStatus(NoReadyTasks)
		} else {
			e.changeStatus(Ready)
		}
		taskToSend := e.tasksHandler.TaskToSendFabricAdd(maybeReadyTask, time.Now())
		return taskToSend
	} else {
		e.changeStatus(NoReadyTasks)
		return TaskToSend{}
	}
}

func (e *Expression) changeStatus(status ExprStatus) {
	e.mut.Lock()
	defer e.mut.Unlock()
	if e.Status == status {
		return
	}
	if e.Status != Completed && e.Status != Cancelled {
		e.Status = status
	} else {
		log.Printf("попытка изменения статуса выражения %d, когда его статус %v", e.ID, e.Status)
	}
}

func (e *Expression) Marshal() (result []byte, err error) {
	result, err = json.Marshal(&e)
	return
}

func (e *Expression) MarshalID() (result []byte, err error) {
	result, err = json.Marshal(&struct {
		ID int `json:"id"`
	}{ID: e.ID})
	return
}

func (e *Expression) WriteResultIntoTask(taskID int, result int64, timeAtReceiveTask time.Time) (err error) {
	task, timeAtSendingTask, ok := e.tasksHandler.popSentTask(taskID)
	if !ok {
		return TaskIDNotExist{taskID}
	}
	if factTime := timeAtReceiveTask.Sub(timeAtSendingTask); factTime > task.OperationTime {
		e.changeStatus(Cancelled)
		return TimeoutExecution{task.OperationTime, factTime, task.Operation,
			task.PairID}
	}
	err = task.WriteResult(result)
	if err != nil {
		log.Panic(err)
	}
	e.tasksHandler.CountUpdatedTask()
	//if tasksHandler.UpdatedAll
	if e.tasksHandler.Len() == 1 {
		e.changeStatus(Completed)
		e.writeResult(task.result)
	}
	return
}

func (e *Expression) writeResult(result int64) {
	e.mut.Lock()
	defer e.mut.Unlock()
	e.Result = result
}

func (e *Expression) GetTasksHandler() *Tasks {
	return e.tasksHandler
}

type ExpressionsJsonTitle struct {
	Expressions []*Expression `json:"expressions"`
}

func (e *ExpressionsJsonTitle) Marshal() (result []byte, err error) {
	result, err = json.Marshal(&e)
	return
}

type ExpressionJsonTitle struct {
	Expression *Expression `json:"expression"`
}

func (e *ExpressionJsonTitle) Marshal() (result []byte, err error) {
	result, err = json.Marshal(&e)
	return
}

type TaskStatus int

const (
	ReadyToCalc TaskStatus = iota
	Sent
	WaitingOtherTasks
	Calculated
)

type Task struct {
	PairID        int           `json:"id"`
	Arg1          interface{}   `json:"arg1"`
	Arg2          interface{}   `json:"arg2"`
	Operation     string        `json:"operation"`
	OperationTime time.Duration `json:"operationTime"`
	result        int64
	Status        TaskStatus
	mut           sync.Mutex
}

func (t *Task) Marshal() (result []byte, err error) {
	result, err = json.Marshal(&struct { // нужно отфильтровать публичные атрибуты (t.Status), поскольку
		// Marshal их распарсит, даже несмотря на отсутствие дополнительного поля формата `json:""`.
		PairID        int           `json:"id"`
		Arg1          interface{}   `json:"arg1"`
		Arg2          interface{}   `json:"arg2"`
		Operation     string        `json:"operation"`
		OperationTime time.Duration `json:"operationTime"`
	}{t.PairID, t.Arg1, t.Arg2, t.Operation, t.OperationTime})
	if err != nil {
		log.Panic(err)
	}
	return
}

func (t *Task) WriteResult(result int64) error {
	t.mut.Lock()
	defer t.mut.Unlock()
	if t.Status == Sent {
		t.result = result
		t.Status = Calculated
	} else if t.Status == Calculated {
		return errors.New("BUG: разработчиком ожидается, что результат одной и той же задачи не может быть записан" +
			" больше одного раза")
	}
	return nil
}

func (t *Task) ChangeStatus(newStatus TaskStatus) {
	t.mut.Lock()
	defer t.mut.Unlock()
	if t.Status == newStatus {
		return
	}
	if t.Status != Calculated && t.Status != newStatus {
		t.Status = newStatus
	}
}

func (t *Task) IsReadyToCalc() bool {
	return t.Status == ReadyToCalc
}

type AgentResult struct {
	ID     int   `json:"ID"`
	Result int64 `json:"result"`
}

func (a *AgentResult) Marshal() (result []byte, err error) {
	result, err = json.Marshal(&a)
	return
}
