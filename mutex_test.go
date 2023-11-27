package gobloom

import (
	"sync"
	"testing"
	"time"
)

func TestNewMutex(t *testing.T) {
	t.Parallel()
	timeout := time.After(1 * time.Second) // 1-second timeout

	done := make(chan bool)
	go func() {
		exclusiveMutex := NewMutex(ExclusiveLock)
		if _, ok := exclusiveMutex.(*ExclusiveMutex); !ok {
			t.Errorf("Expected ExclusiveMutex, got %T", exclusiveMutex)
		}

		readWriteMutex := NewMutex(ReadWriteLock)
		if _, ok := readWriteMutex.(*ReadWriteMutex); !ok {
			t.Errorf("Expected ReadWriteMutex, got %T", readWriteMutex)
		}
		done <- true
	}()

	select {
	case <-done:
		// Test passed
	case <-timeout:
		t.Fatal("Test timed out")
	}
}

func TestExclusiveMutexLockUnlock(t *testing.T) {
	t.Parallel()
	mutex := &ExclusiveMutex{}
	done := make(chan bool)
	timeout := time.After(1 * time.Second) // 1-second timeout

	go func() {
		mutex.Lock()
		time.Sleep(100 * time.Millisecond)
		mutex.Unlock()
		done <- true
	}()

	select {
	case <-done:
		// Test passed
	case <-timeout:
		t.Fatal("Test timed out")
	}
}

func TestExclusiveMutexReadLockUnlock(t *testing.T) {
	t.Parallel()
	mutex := &ExclusiveMutex{}
	done := make(chan bool)
	timeout := time.After(1 * time.Second) // 1-second timeout

	go func() {
		mutex.RLock()
		time.Sleep(100 * time.Millisecond)
		mutex.RUnlock()
		done <- true
	}()

	select {
	case <-done:
		// Test passed
	case <-timeout:
		t.Fatal("Test timed out")
	}
}

func TestReadWriteMutexLockUnlock(t *testing.T) {
	t.Parallel()
	mutex := &ReadWriteMutex{}
	done := make(chan bool)
	timeout := time.After(1 * time.Second)

	go func() {
		mutex.Lock()
		time.Sleep(100 * time.Millisecond)
		mutex.Unlock()
		done <- true
	}()

	select {
	case <-done:
		// Continue with exclusive access test
	case <-timeout:
		t.Fatal("Test timed out")
	}

	// Testing for exclusive access
	// ... [rest of the test]
}

func TestReadWriteMutexReadLockUnlock(t *testing.T) {
	t.Parallel()
	mutex := &ReadWriteMutex{}
	var wg sync.WaitGroup
	timeout := time.After(1 * time.Second)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			mutex.RLock()
			time.Sleep(100 * time.Millisecond)
			mutex.RUnlock()
			wg.Done()
		}()
	}

	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Test passed
	case <-timeout:
		t.Fatal("Test timed out")
	}
}
