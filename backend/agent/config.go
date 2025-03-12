package main

func getDefaultAgent() *Agent {
	return &Agent{ServerURL: "http://localhost:8000", getEndpoint: "/internal/task",
		sendEndpoint: "/internal/task"}
}
