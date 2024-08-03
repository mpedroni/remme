package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"time"

	yaml "gopkg.in/yaml.v3"
)

type Reminder struct {
	Hint              string `yaml:"hint"`
	Message           string `yaml:"message"`
	IntervalInMinutes int    `yaml:"interval-in-minutes"`
}

func (r *Reminder) Validate() error {
	if r.Hint == "" {
		return errors.New("name cannot be empty")
	}

	if r.Message == "" {
		return errors.New("message cannot be empty")
	}

	if r.IntervalInMinutes == 0 {
		return errors.New("interval cannot be zero")
	}

	return nil
}

func main() {
	logger := log.New(os.Stdout, "[remme] ", log.LstdFlags)

	files, err := os.ReadDir("./reminders")
	if err != nil {
		logger.Printf("error reading reminders directory: %v\n", err)
	}

	logger.Printf("found %d reminders configs\n", len(files))

	reminders := make([]Reminder, 0)

	for _, entry := range files {
		if entry.IsDir() {
			continue
		}

		f, err := os.ReadFile("./reminders/" + entry.Name())
		if err != nil {
			logger.Printf("error reading file '%s': %v\n", entry.Name(), err)
		}

		var r Reminder

		if err = yaml.Unmarshal(f, &r); err != nil {
			logger.Printf("error unmarshalling file '%s': %v\n", entry.Name(), err)

		}

		reminders = append(reminders, r)
	}

	for _, r := range reminders {
		if err := r.Validate(); err != nil {
			logger.Printf("reminder with name '%s' is invalid: %v\n", r.Hint, err)
		}

		go func(r Reminder) {
			ticker := time.NewTicker(time.Duration(r.IntervalInMinutes) * time.Minute)
			for range ticker.C {
				logger.Printf("sending reminder: %s\n", r.Hint)
				cmd := exec.Command("notify-send", "-u", "NORMAL", r.Hint, r.Message)
				if err := cmd.Run(); err != nil {
					logger.Printf("unable to send notification: %v\n", err)
				}
			}
		}(r)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	fmt.Println("shutting down")
}
