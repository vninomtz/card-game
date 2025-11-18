package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type Event struct {
	Event string
	Data  string
}

type Listener struct {
	url     string
	eventch chan *Message
}

func NewEventListener(url string, evench chan *Message) *Listener {
	return &Listener{
		url:     url,
		eventch: evench,
	}
}

func (l *Listener) Listen() error {
	resp, err := http.Get(l.url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	reader := bufio.NewReader(resp.Body)
	var ev Event
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return errors.New(fmt.Sprintf("Error to read from Body: %s", err))
		}
		line = strings.TrimSpace(line)
		if line == "" {
			l.Parse(ev)
			ev = Event{}
			continue
		}
		if strings.HasPrefix(line, "event:") {
			ev.Event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		}
		if strings.HasPrefix(line, "data:") {
			ev.Data = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		}
	}
}
func (l *Listener) Parse(ev Event) {
	var msg Message

	if err := json.Unmarshal([]byte(ev.Data), &msg); err != nil {
		log.Printf("Error to parse Message from event %v\n", ev)
		return
	}

	l.eventch <- &msg
}
