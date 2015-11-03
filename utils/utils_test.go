package utils

import (
	"testing"
	"time"
)

func TestSimpleNotify(t *testing.T) {
	sn := NewSimpleNotify()

	// Run notify on no channels, make sure it doesn't panic or etc.
	sn.Notify()

	a := make(chan struct{}, 1)
	sn.Register(a)
	sn.Notify()
	select {
	case <-a:
	case <-time.After(200 * time.Millisecond):
		t.Error("Expected Notify() to call a single channel.")
	}

	b := make(chan struct{}, 1)
	sn.Register(b)
	sn.Notify()

	aCount, bCount := 0, 0
	timeout := time.After(200 * time.Millisecond)
	for i := 0; i < 2; i++ {
		select {
		case <-a:
			aCount++
		case <-b:
			bCount++
		case <-timeout:
			t.Error("Timed out waiting for Notify() to call all channels")
		}
	}

	if aCount != 1 {
		t.Errorf("Expected Notify to call each channel once. Expected aCount of 1, got %d", aCount)
	}

	if bCount != 1 {
		t.Errorf("Expected Notify to call each channel once. Expected bCount of 1, got %d", bCount)
	}

	// Using a non-buffered channel here. Please see the explanation below
	// on the line that checks for:
	//
	//		if notBufferedCount != 0 {
	notBuffered := make(chan struct{})

	sn.Register(notBuffered)
	sn.Notify()

	aCount, bCount, notBufferedCount := 0, 0, 0
	timeout = time.After(200 * time.Millisecond)
	for i := 0; i < 2; i++ {
		select {
		case <-a:
			aCount++
		case <-b:
			bCount++
		case <-notBuffered:
			notBufferedCount++
		case <-timeout:
			t.Error("Timed out waiting for Notify() to call all channels")
		}
	}

	if aCount != 1 {
		t.Errorf("Expected Notify to call each buffered channel once. Expected aCount of 1, got %d", aCount)
	}

	if bCount != 1 {
		t.Errorf("Expected Notify to call each buffered channel once. Expected bCount of 1, got %d", bCount)
	}

	// Not buffered, means the Notify will be unable to send to the channel
	// because nothing is *currently* receiving from the channel when
	// Notify() occurs. Notify() should not block and wait for it.
	//
	// If Notify() does block, the receiver (in this test) will never get
	// a chance to listen either.. so really this shouldn't happen.
	if notBufferedCount != 0 {
		t.Errorf("Expected Notify() to ignored the notBuffered channel because, at the time of sending, it was not receiving. Wanted 0, Got %d",
			notBufferedCount)
	}

	sn.Unregister(b)
	sn.Notify()

	_, ok := <-b
	if ok {
		t.Errorf("Expected Notify() not to call unregistered channels. Wanted false, got %t", ok)
	}
}
