package nostr

import (
	"context"
	"sync"
)

type Subscription struct {
	id    string
	conn  *Connection
	mutex sync.Mutex

	Relay             *Relay
	Filters           Filters
	Events            chan Event
	EndOfStoredEvents chan struct{}

	stopped  bool
	emitEose sync.Once
}

type EventMessage struct {
	Event Event
	Relay string
}

func (sub *Subscription) Unsub() {
	sub.mutex.Lock()
	defer sub.mutex.Unlock()

	sub.conn.WriteJSON([]interface{}{"CLOSE", sub.id})
	if sub.stopped == false && sub.Events != nil {
		close(sub.Events)
	}
	sub.stopped = true
}

func (sub *Subscription) Sub(ctx context.Context, filters Filters) {
	sub.Filters = filters
	sub.Fire(ctx)
}

func (sub *Subscription) Fire(ctx context.Context) {
	message := []interface{}{"REQ", sub.id}
	for _, filter := range sub.Filters {
		message = append(message, filter)
	}

	sub.conn.WriteJSON(message)

	// the subscription ends once the context is canceled
	go func() {
		<-ctx.Done()
		sub.Unsub()
	}()
}
