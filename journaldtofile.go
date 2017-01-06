package main

import "flag"
import "fmt"
import "os"
import "os/signal"
import "syscall"
import "time"

import "github.com/mheese/go-systemd/sdjournal"

func process(filename string, recv chan sdjournal.JournalEntry, rotate chan os.Signal) {
	outputfile, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer outputfile.Close()

	var line string
	for {
		select {
		case entry := <-recv:
			line = entry["MESSAGE"].(string) + "\n"
			outputfile.WriteString(line)
		case <-rotate:
			outputfile.Close()
			outputfile, err = os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				panic(err)
			}
		}
	}
}

func main() {
	var filename string
	var unit string
	flag.StringVar(&filename, "logfile", "/var/log/fromjournal.log", "File name of logfile to write to")
	flag.StringVar(&unit, "unit", "", "Only process messages from this systemd unit")
	flag.Parse()

	done := make(chan int, 1)
	recv := make(chan sdjournal.JournalEntry)
	rotate := make(chan os.Signal)

	signal.Notify(rotate, syscall.SIGHUP)

	config := sdjournal.JournalReaderConfig{
		Since: time.Duration(1),
	}
	var unitlog string
	if unit != "" {
		config.Matches = []sdjournal.Match{{Field: "_SYSTEMD_UNIT", Value: unit}}
		unitlog = " for systemd unit " + unit
	} else {
		unitlog = ""
	}
	jr, err := sdjournal.NewJournalReader(config)
	if err != nil {
		fmt.Printf("Could not create JournalReader: %v\n", err)
		return
	}
	go process(filename, recv, rotate)

	fmt.Printf("Saving journald data%v to %v\n", unitlog, filename)
	jr.FollowJournal(done, recv)

}
