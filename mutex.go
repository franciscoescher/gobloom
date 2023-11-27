package gobloom

import (
	"errors"
	"sync"
)

var ErrInvalidLockType = errors.New("invalid lock type")

type Mutex interface {
	WLock()
	WUnlock()
	RLock()
	RUnlock()
}

func NewMutex(l LockType) (Mutex, error) {
	switch l {
	case LockTypeNone:
		return nil, nil
	case LockTypeExclusive:
		return &ExclusiveMutex{}, nil
	case LockTypeReadWrite:
		return &ReadWriteMutex{}, nil
	}
	return nil, ErrInvalidLockType
}

type ExclusiveMutex struct {
	m sync.Mutex
}

func (l *ExclusiveMutex) WLock() {
	l.m.Lock()
}

func (l *ExclusiveMutex) WUnlock() {
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

func (l *ReadWriteMutex) WLock() {
	l.m.Lock()
}

func (l *ReadWriteMutex) WUnlock() {
	l.m.Unlock()
}

func (l *ReadWriteMutex) RLock() {
	l.m.RLock()
}

func (l *ReadWriteMutex) RUnlock() {
	l.m.RUnlock()
}
