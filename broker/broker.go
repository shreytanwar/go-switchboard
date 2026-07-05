import (
	"sync"
	"time"
)

type Message struct {
	Topic	 	string	//topic from/to?
	Payload 	[]byte
	At 			time.Time
}

type Subscriber struct {
	Ch chan Message // Each subscriber gets its own buffered channel (chan Message, buffer size ~64) so a slow consumer doesn't block the broker.
	Topic string
	DroppedCount atomic.Int64
}

type Broker struct {
	mu sync.RWMutex // read write mutex
	topics map[string][]*Subscriber // topic -> list of subscribers
}

func New() *Broker {
	return &Broker{
		topics: make(map[string][]*Subscriber),
	}
}
func(b *Broker) Subscribe(topic string) *Subscriber {
	sub := &Subscriber{
		Ch:	make(chan Message, 64), // buffered channel
		Topic: topic,
	}
	b.mu.Lock()
	b.topics[topic] = append(b.topics[topic], sub)
	b.mu.Unlock()

	return sub
}
func(b *Broker)	Unsubscribe(sub *Subscriber) {
	b.muLock()
	defer b.mu.Unlock()

	subs := b.topics(sub.Topic)

	for i, s := range subs {
		if s == sub {
			b.topics[sub.Topic] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
	b.topics[sub.Topic] = 
}
func(b *Broker) Publish(topic string, payload []byte) {

	msg := Message {
		Topic:   topic,
		Payload: payload,
		At:      time.Now(),
	}

	b.mu.RLock()
	subs := make([]*Subscriber, len(b.topics[topic]))
	copy(subs, b.topics[topic])
	b.mu.RUnlock()
	

	for _, sub := range subs {
		// non-blocking send — slow consumers drop messages rather than stalling everyone
		select {
		case sub.Ch <- msg:
			// message sent to subscriber
		default:
		}
	}
	msg := Message{
		Topic: topic,
		Payload: payload,
		At: time.Now(),
	}

	for _, sub := range subs {
		select {
		case sub.Ch <- msg:
			// message sent to subscriber
		default:
			// subscriber's channel is full, drop the message
			dropped := sub.DroppedCount.Add(1) // atomic counter
			log.Printf("warn: subscriber on topic %q dropped message (total dropped: %d)", sub.Topic, dropped)
		}
	}
}
