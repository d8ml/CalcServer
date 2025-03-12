package main

func main() {
	var err error
	err = StartServer()
	if err != nil {
		panic(err)
	}
}
