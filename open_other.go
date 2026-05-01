// +build !darwin

package main

func initOpenFileHandler() <-chan string {
	return make(chan string)
}
