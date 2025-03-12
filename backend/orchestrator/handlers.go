package main

import (
	"encoding/json"
	"errors"
	"github.com/Debianov/calc-ya-go-24/backend"
	"github.com/Debianov/calc-ya-go-24/pkg"
	"io"
	"log"
	"net/http"
	"slices"
	"strconv"
	"time"
)

var exprsList = backend.ExpressionListEmptyFabric()

func calcHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err error
	)
	if r.Method != http.MethodPost {
		return
	}
	if r.Header.Get("Content-Type") != "application/json" {
		return
	}
	var (
		buf           []byte
		requestStruct backend.RequestJson
		reader        io.ReadCloser
	)
	reader = r.Body
	buf, err = io.ReadAll(reader)
	if err != nil {
		log.Panic(err)
	}
	err = json.Unmarshal(buf, &requestStruct)
	if err != nil {
		log.Panic(err)
	}
	postfix, ok := pkg.GeneratePostfix(requestStruct.Expression)
	if !ok {
		w.WriteHeader(422)
		return
	}
	expr, _ := exprsList.ExprFabricAdd(postfix)
	marshaledExpr, err := expr.MarshalID()
	if err != nil {
		log.Panic(err)
	}
	w.WriteHeader(201)
	_, err = w.Write(marshaledExpr)
	if err != nil {
		log.Panic(err)
	}
}

func expressionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		return
	}
	var err error
	exprs := exprsList.GetAllExprs()
	slices.SortFunc(exprs, func(expression *backend.Expression, expression2 *backend.Expression) int {
		if expression.ID >= expression2.ID {
			return 0
		} else {
			return -1
		}
	})
	var exprsJsonHandler = backend.ExpressionsJsonTitle{Expressions: exprs}
	exprsHandlerInBytes, err := exprsJsonHandler.Marshal()
	if err != nil {
		log.Panic(err)
	}
	_, err = w.Write(exprsHandlerInBytes)
	if err != nil {
		log.Panic(err)
	}
}

func expressionIdHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		return
	}
	var err error
	id := r.PathValue("ID")
	idInINt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		log.Panic(err)
	}
	expr, exist := exprsList.Get(int(idInINt))
	if !exist {
		w.WriteHeader(404)
		return
	}
	var exprJsonHandler = backend.ExpressionJsonTitle{expr}
	exprHandlerInBytes, err := json.Marshal(&exprJsonHandler)
	if err != nil {
		log.Panic(err)
	}
	_, err = w.Write(exprHandlerInBytes)
	if err != nil {
		log.Panic(err)
	}
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		taskGetHandler(w, r)
	} else if r.Method == http.MethodPost {
		taskPostHandler(w, r)
	}
}

func taskGetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		return
	}
	var err error
	expr := exprsList.GetReadyExpr()
	if expr == nil {
		w.WriteHeader(404)
		return
	}
	responseInJson := expr.FabricReadyExprSendTask()
	if responseInJson.Task == nil {
		w.WriteHeader(404)
		return
	}
	taskJsonHandlerInBytes, err := responseInJson.Marshal()
	if err != nil {
		log.Panic(err)
	}
	_, err = w.Write(taskJsonHandlerInBytes)
	if err != nil {
		log.Panic(err)
	}
	responseInJson.Task.ChangeStatus(backend.Sent)
}

func taskPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}
	if r.Header.Get("Content-Type") != "application/json" {
		return
	}
	var (
		err    error
		reqBuf = make([]byte, r.ContentLength)
	)
	_, err = r.Body.Read(reqBuf)
	if err != nil && err != io.EOF {
		log.Panic(err)
	}
	var (
		reqInJson backend.AgentResult
	)
	err = json.Unmarshal(reqBuf, &reqInJson)
	if err != nil {
		log.Panic(err)
		//w.WriteHeader(422) // TODO проверка структуры
	}
	exprId, _ := pkg.Unpair(reqInJson.ID)
	expr, ok := exprsList.Get(exprId)
	if !ok {
		w.WriteHeader(404)
		return
	}
	err = expr.WriteResultIntoTask(reqInJson.ID, reqInJson.Result, time.Now())
	if err != nil {
		if errors.Is(err, backend.TaskIDNotExist{}) {
			w.WriteHeader(404)
			return
		} else {
			log.Panic(err)
		}
	}

}

func panicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("response %s, status code: 500", w)
				writeInternalServerError(w)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func writeInternalServerError(w http.ResponseWriter) {
	w.WriteHeader(500)
	return
}

func getHandler() (handler http.Handler) {
	var mux = http.NewServeMux()
	mux.HandleFunc("/api/v1/calculate", calcHandler)
	mux.HandleFunc("/api/v1/expressions", expressionsHandler)
	mux.HandleFunc("/api/v1/expressions/{ID}", expressionIdHandler)
	mux.HandleFunc("/internal/task", taskHandler)
	handler = panicMiddleware(mux)
	return
}
