package services

import (
	"fmt"
	"sync"
	"time"
)

type Event struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}

type EventBroker struct {
	mu          sync.RWMutex
	subscribers map[chan Event]struct{}
}

func NewEventBroker() *EventBroker {
	return &EventBroker{subscribers: make(map[chan Event]struct{})}
}

func (broker *EventBroker) Subscribe() chan Event {
	channel := make(chan Event, 16)
	broker.mu.Lock()
	broker.subscribers[channel] = struct{}{}
	broker.mu.Unlock()
	return channel
}

func (broker *EventBroker) Unsubscribe(channel chan Event) {
	broker.mu.Lock()
	delete(broker.subscribers, channel)
	close(channel)
	broker.mu.Unlock()
}

func (broker *EventBroker) Publish(eventType string, payload interface{}) {
	broker.mu.RLock()
	defer broker.mu.RUnlock()
	event := Event{Type: eventType, Timestamp: time.Now(), Payload: payload}
	for channel := range broker.subscribers {
		select {
		case channel <- event:
		default:
			fmt.Printf("drop event %s\n", eventType)
		}
	}
}
