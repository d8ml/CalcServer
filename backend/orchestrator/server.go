package main

func StartServer() (err error) {
	s := GetDefaultServer(getHandler())
	err = s.ListenAndServe()
	return
}
