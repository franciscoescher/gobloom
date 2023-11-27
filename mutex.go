package gobloom

import "sync"

type Mutex interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
}

func NewMutex(l LockType) Mutex {
	switch l {
	case ExclusiveLock:
		return &ExclusiveMutex{}
	case ReadWriteLock:
		return &ReadWriteMutex{}
	}
	return nil
}

type ExclusiveMutex struct {
	m sync.Mutex
}

func (l *ExclusiveMutex) Lock() {
	l.m.Lock()
}

func (l *ExclusiveMutex) Unlock() {
	l.m.Unlock()
}

func (l *ExclusiveMutex) RLock() {
	l.m.Lock()
}

func (l *ExclusiveMutex) RUnlock() {
	l.m.Unlock()
}

type ReadWriteMutex struct {
	m sync.RWMutex
}

func (l *ReadWriteMutex) Lock() {
	l.m.Lock()
}

func (l *ReadWriteMutex) Unlock() {
	l.m.Unlock()
}

func (l *ReadWriteMutex) RLock() {
	l.m.RLock()
}

func (l *ReadWriteMutex) RUnlock() {
	l.m.RUnlock()
}
