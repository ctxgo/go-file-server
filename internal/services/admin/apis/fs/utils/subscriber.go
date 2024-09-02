package utils

import (
	"context"
	"sync"
)

type Publishers[T any] struct {
	publishers map[string]*Publisher[T]
	sync.RWMutex
}

func NewPublishers[T any]() *Publishers[T] {
	return &Publishers[T]{
		publishers: make(map[string]*Publisher[T]),
	}
}

func (ps *Publishers[T]) get(k string) (*Publisher[T], bool) {
	v, ok := ps.publishers[k]
	return v, ok
}

func (ps *Publishers[T]) Get(k string) (*Publisher[T], bool) {
	ps.RLock()
	defer ps.RUnlock()
	return ps.get(k)
}

func (ps *Publishers[T]) Create(ctx context.Context, k string) (*Publisher[T], bool) {
	ps.Lock()
	defer ps.Unlock()
	v, ok := ps.get(k)
	if ok {
		return v, false
	}
	v = NewPublisher[T](ctx)
	ps.publishers[k] = v
	return v, true
}

func (ps *Publishers[T]) Del(k string) {
	ps.Lock()
	defer ps.Unlock()
	delete(ps.publishers, k)
}

// Subscriber 订阅者接口
type Subscriber[T any] interface {
	Receive(T)
}

type Message struct {
	K string
	V string
}

func NewMessage(k, v string) Message {
	return Message{
		K: k,
		V: v,
	}
}

type ChanSubscriber chan Message

func (cs ChanSubscriber) Receive(msg Message) {
	cs <- msg
}

// Publisher 消息分发器
type Publisher[T any] struct {
	mu          sync.RWMutex
	subscribers map[Subscriber[T]]struct{}
	done        <-chan struct{}
	messageChan chan T
	lastMessage T
	wg          sync.WaitGroup
	cancel      context.CancelFunc
}

func NewPublisher[T any](ctx context.Context) *Publisher[T] {
	ctx, cancel := context.WithCancel(ctx)

	publisher := &Publisher[T]{
		subscribers: make(map[Subscriber[T]]struct{}),
		messageChan: make(chan T, 100),
		done:        ctx.Done(),
		cancel:      cancel,
	}
	go publisher.run()
	return publisher
}

func (p *Publisher[T]) run() {
	for {
		select {
		case msg, ok := <-p.messageChan:
			if !ok {
				continue
			}
			p.mu.Lock()
			for sub := range p.subscribers {
				sub.Receive(msg)
			}
			p.mu.Unlock()
			p.wg.Done()
		case <-p.Done():
			return
		}
	}
}

func (p *Publisher[T]) AddSubscriber(sub Subscriber[T]) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.subscribers[sub] = struct{}{}
}

func (p *Publisher[T]) RemoveSubscriber(sub Subscriber[T]) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.subscribers, sub)
}

func (p *Publisher[T]) Done() <-chan struct{} {
	return p.done
}

func (p *Publisher[T]) Close() {
	p.wg.Wait()
	close(p.messageChan)
	p.cancel()
}

func (p *Publisher[T]) GetLastMessage() T {
	return p.lastMessage
}

func (p *Publisher[T]) Publish(msg T) {
	p.wg.Add(1)
	p.lastMessage = msg
	p.messageChan <- msg
}
