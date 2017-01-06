package main

import "fmt"
import "os"
import "time"

import "github.com/mheese/go-systemd/sdjournal"

func process(recv chan sdjournal.JournalEntry) {
	fmt.Printf("In process\n")
	outputfile, err := os.OpenFile("logfile.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
//		fmt.Printf("Could not open file: %v\n", err)
		panic(err)
	}
	defer outputfile.Close()

	var line string
	for entry := range recv {
		line = entry["MESSAGE"].(string) + "\n"
		outputfile.WriteString(line)
	}
}

func main() {

	var done chan int
	var recv chan sdjournal.JournalEntry
	done = make(chan int, 1)
	recv = make(chan sdjournal.JournalEntry)

	jr, err := sdjournal.NewJournalReader(sdjournal.JournalReaderConfig{
		Since: time.Duration(1),
		//          NumFromTail: 0,
	})
	if err != nil {
		fmt.Printf("Could not create JournalReader: %v\n", err)
		return
	}
	go process(recv)

	fmt.Printf("Starting followjournal\n")
	jr.FollowJournal(done, recv)
	
}
