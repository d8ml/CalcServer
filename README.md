Сервис подсчёта арифметических выражений. Поддерживает операторы +, -, /, *, а также скобочки для приоритезации отдельных 
частей выражения.

Разделён на оркестратор и агент. Оркестратор отвечает за приём новых выражений,
а агент — за их вычисление.

# Развёртывание
`git clone https://github.com/Debianov/calc-ya-go-24.git`

В `backend/orchestrator/config.go` в строке `return &http.Server{Addr: "127.0.0.1:8000", Handler: handler}` может быть 
изменён адрес `Addr` на любой желаемый. Соответствующие изменения также нужно сделать в `backend/agent/config.go`, иначе 
оркестратор и агент не будут корректно работать.

Для работы программы желательна последняя версия Go 1.24 ([как обновить Go](https://go.dev/doc/install), 
если в репозиториях пакетных менеджеров ещё нет новой версии). **Работа проекта протестирована на 
версии 1.24.**

# Переменные среды
Необходимые переменные среды для работы оркестратора:
```
TIME_ADDITION_MS
TIME_SUBTRACTION_MS
TIME_MULTIPLICATIONS_MS
TIME_DIVISIONS_MS
```
Формат значений переменных: `<число><ns/us/ms/s/m>`

Необходимые переменные среды для работы агента:
```
COMPUTING_POWER
```
Формат значений: число.


Пример файла переменных в Linux:
```shell
# filename: calc.env
#!/bin/sh
export TIME_ADDITION_MS=2s
export TIME_SUBTRACTION_MS=2s
export TIME_MULTIPLICATIONS_MS=2s
export TIME_DIVISIONS_MS=2s
export COMPUTING_POWER=10
```

Экспортирование переменных в Linux:
`source calc.env`

# Запуск

Запуск оркестратора:
```shell
cd ./backend/orchestrator
go run github.com/Debianov/calc-ya-go-24/backend/orchestrator
```
Запуск агента:
```shell
cd ./backend/agent
go run github.com/Debianov/calc-ya-go-24/backend/agent
```
Для успешного запуска агента необходимо, чтобы оркестратор был запущен.

# Использование

## Внешние endpoint-ы

Запрос на регистрацию нового выражения:
```shell
curl --location 'localhost:8000/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2+2*4" 
}'
```

Запрос на получение списка выражений:
```shell
curl --location 'localhost:8000/api/v1/expressions'
```

Запрос на получение конкретного выражения по id:
```shell
curl --location 'localhost:8000/api/v1/expressions/id'
```

## Внутренние endpoint-ы
Используются агентом.

Запрос на получение задачи (GET):
```shell
curl --location 'localhost:8000/internal/task'
```

Запрос на отправку задачи (POST):
```shell
curl --location 'localhost:8000/internal/task' \
--header 'Content-Type: application/json' \
--data '{
  "id": 0,
  "result": 10
}'
```

Ответы возвращаются также в формате json. В случае, если код ответа не 200 и не 201,
будет возвращена пустая строка.

# Тестирование
Для работы также необходимы экспортированные переменные окружения.
```shell
cd ./backend/orchestrator
go test
```
