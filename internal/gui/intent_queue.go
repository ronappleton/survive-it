package gui

import "github.com/appengine-ltd/survive-it/internal/parser"

type CommandSink interface {
	EnqueueIntent(parser.Intent)
}

type intentQueue struct {
	ch chan parser.Intent
}

func newIntentQueue(size int) *intentQueue {
	if size < 1 {
		size = 16
	}
	return &intentQueue{ch: make(chan parser.Intent, size)}
}

func (q *intentQueue) EnqueueIntent(intent parser.Intent) {
	if q == nil {
		return
	}
	select {
	case q.ch <- intent:
	default:
		// Drop only when queue is saturated; parser input is non-critical.
	}
}

func (q *intentQueue) Dequeue() (parser.Intent, bool) {
	if q == nil {
		return parser.Intent{}, false
	}
	select {
	case intent := <-q.ch:
		return intent, true
	default:
		return parser.Intent{}, false
	}
}
