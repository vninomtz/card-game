package main

import "log"

func main() {
	srv := NewServer("", 8000)

	err := srv.Run()

	if err != nil {
		log.Println(err)
	}
}
