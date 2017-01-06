package main

import "fmt"
import "os"
import "os/signal"
import "syscall"
import "time"

import "github.com/mheese/go-systemd/sdjournal"

func process(recv chan sdjournal.JournalEntry, rotate chan os.Signal) {
	fmt.Printf("In process\n")
	outputfile, err := os.OpenFile("logfile.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
//		fmt.Printf("Could not open file: %v\n", err)
		panic(err)
	}
	defer outputfile.Close()

	var line string
	for {
		select {
		case entry := <- recv:
			line = entry["MESSAGE"].(string) + "\n"
			outputfile.WriteString(line)
		case <-rotate:
			outputfile.Close()
			outputfile, err = os.OpenFile("logfile.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				//		fmt.Printf("Could not open file: %v\n", err)
				panic(err)
			}
		}
	}
}

func main() {
	done := make(chan int, 1)
	recv := make(chan sdjournal.JournalEntry)
	rotate := make(chan os.Signal)

	signal.Notify(rotate, syscall.SIGHUP)

	jr, err := sdjournal.NewJournalReader(sdjournal.JournalReaderConfig{
		Since: time.Duration(1),
		//          NumFromTail: 0,
	})
	if err != nil {
		fmt.Printf("Could not create JournalReader: %v\n", err)
		return
	}
	go process(recv, rotate)

	fmt.Printf("Starting followjournal\n")
	jr.FollowJournal(done, recv)
	
}
