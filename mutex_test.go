package gobloom

import (
	"testing"
	"time"
)

func TestNewMutex(t *testing.T) {
	t.Parallel()

	exclusiveMutex, err := NewMutex(LockTypeExclusive)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if _, ok := exclusiveMutex.(*ExclusiveMutex); !ok {
		t.Errorf("Expected ExclusiveMutex, got %T", exclusiveMutex)
	}

	readWriteMutex, err := NewMutex(LockTypeReadWrite)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if _, ok := readWriteMutex.(*ReadWriteMutex); !ok {
		t.Errorf("Expected ReadWriteMutex, got %T", readWriteMutex)
	}

	_, err = NewMutex(0)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestMutexLocks(t *testing.T) {
	type testCase struct {
		name         string
		mutex        Mutex
		lock         bool
		getLockFuncs func(mutex Mutex) (func(), func(), func())
	}

	tests := []testCase{
		{
			name:  "ExclusiveMutex_Read_Read",
			mutex: &ExclusiveMutex{},
			getLockFuncs: func(mutex Mutex) (func(), func(), func()) {
				return mutex.RLock, mutex.RLock, mutex.RUnlock
			},
			lock: true,
		},
		{
			name:  "ExclusiveMutex_Write_Read",
			mutex: &ExclusiveMutex{},
			getLockFuncs: func(mutex Mutex) (func(), func(), func()) {
				return mutex.WLock, mutex.RLock, mutex.RUnlock
			},
			lock: true,
		},
		{
			name:  "ExclusiveMutex_Read_Write",
			mutex: &ExclusiveMutex{},
			getLockFuncs: func(mutex Mutex) (func(), func(), func()) {
				return mutex.WLock, mutex.RLock, mutex.RUnlock
			},
			lock: true,
		},
		{
			name:  "ReadWriteMutex_Write_Write",
			mutex: &ExclusiveMutex{},
			getLockFuncs: func(mutex Mutex) (func(), func(), func()) {
				return mutex.WLock, mutex.WLock, mutex.WUnlock
			},
			lock: true,
		},
		{
			name:  "ReadWriteMutex_Read_Read",
			mutex: &ReadWriteMutex{},
			getLockFuncs: func(mutex Mutex) (func(), func(), func()) {
				return mutex.RLock, mutex.RLock, mutex.RUnlock
			},
			lock: false,
		},
		{
			name:  "ReadWriteMutex_Write_Read",
			mutex: &ReadWriteMutex{},
			getLockFuncs: func(mutex Mutex) (func(), func(), func()) {
				return mutex.WLock, mutex.RLock, mutex.RUnlock
			},
			lock: true,
		},
		{
			name:  "ReadWriteMutex_Read_Write",
			mutex: &ReadWriteMutex{},
			getLockFuncs: func(mutex Mutex) (func(), func(), func()) {
				return mutex.WLock, mutex.RLock, mutex.RUnlock
			},
			lock: true,
		},
		{
			name:  "ExclusiveMutex_Write_Write",
			mutex: &ReadWriteMutex{},
			getLockFuncs: func(mutex Mutex) (func(), func(), func()) {
				return mutex.WLock, mutex.WLock, mutex.WUnlock
			},
			lock: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel()
			done := make(chan bool)
			timeout := time.After(1 * time.Second)
			lock1, lock2, unlock2 := tc.getLockFuncs(tc.mutex)

			lock1()

			go func() {
				// Attempt to acquire another read lock
				lock2()
				defer unlock2()
				done <- true
			}()

			select {
			case <-done:
				if tc.lock {
					t.Fatal("Test passed, but should have locked")
				}
				break
			case <-timeout:
				if !tc.lock {
					t.Fatal("Test timed out")
				}
				break
			}
		})
	}
}
