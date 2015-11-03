package utils

import (
	"net"
	"strconv"
	"sync"
)

func NewSimpleNotify() *SimpleNotify {
	return &SimpleNotify{
		nodata:    struct{}{},
		listeners: make(map[chan<- struct{}]struct{}),
	}
}

// SimpleNotify is a type which allows pubsub for events without data.
// No transmition of data is performed, only notification of a
// given event.
type SimpleNotify struct {
	// nodata is used as the value for every listener, since we don't
	// care about the value.
	nodata struct{}

	// A map of channels that want to be notified of successful http
	// registers. Note that we're purely using the map to quickly locate
	// subscribers during removal.
	listeners map[chan<- struct{}]struct{}

	lock sync.Mutex
}

// Notify sends an empty struct to all *currently receiving* registered
// channels, notifying them that an event has occured.
func (sn *SimpleNotify) Notify() {
	sn.lock.Lock()
	defer sn.lock.Unlock()

	for ch, _ := range sn.listeners {
		// Select ensures that if the channel isn't able to receive, it
		// doesn't block forever. Instead, if a channel isn't waiting or
		// buffered then it will simply ignore it.
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// Register adds a notify channel to the SimpleNotify struct.
//
// The caller should never close this channel before Unregistering it.
func (sn *SimpleNotify) Register(ch chan<- struct{}) {
	sn.lock.Lock()
	defer sn.lock.Unlock()

	sn.listeners[ch] = sn.nodata
}

// Unregister removes a channel from the SimpleNotify struct, and
// closes it.
func (sn *SimpleNotify) Unregister(ch chan<- struct{}) {
	sn.lock.Lock()
	defer sn.lock.Unlock()

	_, ok := sn.listeners[ch]
	if ok {
		delete(sn.listeners, ch)
		close(ch)
	}
}

// RandomPort() returns a random port to be used with net.Listen(). It's an
// helper function to register to kontrol before binding to a port. Note that
// this racy, there is a possibility that someoe binds to the port during the
// time you get the port and someone else finds it. Therefore use in caution.
func RandomPort() (int, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer l.Close()

	_, port, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(port)
}
