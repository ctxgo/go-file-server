package utils

import (
	"sync"
	"sync/atomic"
)

type Message struct {
	K string
	V string
}

func NewMessage(k, v string) Message {
	return Message{K: k, V: v}
}

type Subscriber[T any] struct {
	messages chan T
	done     chan struct{}
	mu       sync.RWMutex
	closed   bool
	once     sync.Once
}

func (s *Subscriber[T]) Messages() <-chan T {
	return s.messages
}

func (s *Subscriber[T]) Close() {
	s.once.Do(func() {
		close(s.done)
		s.mu.Lock()
		defer s.mu.Unlock()
		s.closed = true
		close(s.messages)
	})
}

func (s *Subscriber[T]) receive(msg T) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return
	}
	select {
	case s.messages <- msg:
	case <-s.done:
		return
	}
}

// Publisher implements the IPublisher interface, handling message distribution.
type Publisher[T any] struct {
	mu          sync.RWMutex
	subscribers map[*Subscriber[T]]struct{}
	messageChan chan T
	closed      bool
	done        chan struct{}
	lastMessage atomic.Value
}

func NewPublisher[T any]() *Publisher[T] {

	publisher := &Publisher[T]{
		subscribers: make(map[*Subscriber[T]]struct{}),
		messageChan: make(chan T, 100),
		done:        make(chan struct{}),
	}
	go publisher.run()
	return publisher
}

func (p *Publisher[T]) run() {
	for msg := range p.messageChan {
		p.mu.RLock()
		for sub := range p.subscribers {
			go sub.receive(msg)
		}
		p.mu.RUnlock()
	}
}

// CreateSubscriber 创建并返回一个新的订阅者
func (p *Publisher[T]) CreateSubscriber() *Subscriber[T] {
	sub := &Subscriber[T]{
		messages: make(chan T),
		done:     make(chan struct{}),
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.subscribers[sub] = struct{}{}
	go p.monitorSubscriber(sub)
	return sub
}

func (p *Publisher[T]) monitorSubscriber(sub *Subscriber[T]) {
	<-sub.done
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.subscribers, sub)
}

func (p *Publisher[T]) Close() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	p.mu.Unlock()
	close(p.messageChan)
	close(p.done)
}

func (p *Publisher[T]) Done() <-chan struct{} {
	return p.done
}

func (p *Publisher[T]) LastMessage() T {
	val := p.lastMessage.Load()
	if val != nil {
		return val.(T)
	}
	return *new(T)
}

func (p *Publisher[T]) Publish(msg T) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.closed {
		return
	}
	p.messageChan <- msg
	p.lastMessage.Store(msg)
}
