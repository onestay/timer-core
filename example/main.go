package main

import "github.com/onestay/timer-core"

import "fmt"

import "time"

func main() {
	t := timer.New()
	t.StartTimer()
	go func() {
		time.Sleep(2 * time.Second)
		t.PauseTimer()
		time.Sleep(3 * time.Second)
		t.ResumeTimer()
	}()
	for {
		v := <-t.Updates
		fmt.Println(v.Seconds())
	}
}
