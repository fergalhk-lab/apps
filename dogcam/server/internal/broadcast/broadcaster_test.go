package broadcast_test

import (
	"testing"
	"time"

	"github.com/fergalhk-lab/apps/dogcam/gen/dogcampb"
	"github.com/fergalhk-lab/apps/dogcam/server/internal/broadcast"
	"github.com/stretchr/testify/assert"
)

func TestBroadcaster_NoStartWithoutCamera(t *testing.T) {
	b := broadcast.New(2000)
	ch := b.Subscribe("client1")
	defer b.Unsubscribe("client1")

	select {
	case frame := <-ch:
		t.Fatalf("unexpected frame: %v", frame)
	case <-time.After(50 * time.Millisecond):
		// correct: no camera registered, no frames
	}
}

func TestBroadcaster_StartSentOnFirstSubscribe(t *testing.T) {
	b := broadcast.New(2000)
	controlCh := make(chan *dogcampb.ControlMessage, 4)
	b.RegisterCamera(controlCh)

	b.Subscribe("client1")
	defer b.Unsubscribe("client1")

	select {
	case msg := <-controlCh:
		assert.Equal(t, dogcampb.ControlMessage_START, msg.Command)
		assert.Equal(t, int32(2000), msg.FrameIntervalMs)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected START signal")
	}
}

func TestBroadcaster_NoStartOnSecondSubscribe(t *testing.T) {
	b := broadcast.New(2000)
	controlCh := make(chan *dogcampb.ControlMessage, 4)
	b.RegisterCamera(controlCh)

	b.Subscribe("client1")
	// drain START
	<-controlCh

	b.Subscribe("client2")
	defer b.Unsubscribe("client1")
	defer b.Unsubscribe("client2")

	select {
	case msg := <-controlCh:
		t.Fatalf("unexpected control message: %v", msg)
	case <-time.After(50 * time.Millisecond):
		// correct: START only on first subscriber
	}
}

func TestBroadcaster_StopSentWhenLastUnsubscribes(t *testing.T) {
	b := broadcast.New(2000)
	controlCh := make(chan *dogcampb.ControlMessage, 4)
	b.RegisterCamera(controlCh)

	b.Subscribe("client1")
	<-controlCh // drain START

	b.Unsubscribe("client1")

	select {
	case msg := <-controlCh:
		assert.Equal(t, dogcampb.ControlMessage_STOP, msg.Command)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected STOP signal")
	}
}

func TestBroadcaster_FramesFannedOut(t *testing.T) {
	b := broadcast.New(2000)
	ch1 := b.Subscribe("client1")
	ch2 := b.Subscribe("client2")
	defer b.Unsubscribe("client1")
	defer b.Unsubscribe("client2")

	frame := []byte{0xFF, 0xD8, 0xFF} // fake JPEG header
	b.Publish(frame)

	for _, ch := range []<-chan []byte{ch1, ch2} {
		select {
		case got := <-ch:
			assert.Equal(t, frame, got)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("expected frame on subscriber channel")
		}
	}
}

func TestBroadcaster_StartSentOnCameraConnectWhenSubscribersWaiting(t *testing.T) {
	b := broadcast.New(2000)
	// Subscribe before camera connects
	b.Subscribe("client1")
	defer b.Unsubscribe("client1")

	// Now camera connects
	controlCh := make(chan *dogcampb.ControlMessage, 4)
	b.RegisterCamera(controlCh)

	select {
	case msg := <-controlCh:
		assert.Equal(t, dogcampb.ControlMessage_START, msg.Command)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected START when camera connects with existing subscribers")
	}
}
