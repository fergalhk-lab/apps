package broadcast

import (
	"sync"

	"github.com/fergalhk-lab/apps/dogcam/gen/dogcampb"
)

type Broadcaster struct {
	mu              sync.Mutex
	clients         map[string]chan []byte
	controlCh       chan<- *dogcampb.ControlMessage
	frameIntervalMs int32
}

func New(frameIntervalMs int32) *Broadcaster {
	return &Broadcaster{
		clients:         make(map[string]chan []byte),
		frameIntervalMs: frameIntervalMs,
	}
}

// RegisterCamera is called by the gRPC handler when the camera connects.
// If subscribers are already waiting, START is sent immediately.
func (b *Broadcaster) RegisterCamera(ch chan<- *dogcampb.ControlMessage) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.controlCh = ch
	if len(b.clients) > 0 {
		b.sendControl(&dogcampb.ControlMessage{
			Command:         dogcampb.ControlMessage_START,
			FrameIntervalMs: b.frameIntervalMs,
		})
	}
}

// UnregisterCamera is called by the gRPC handler when the camera disconnects.
func (b *Broadcaster) UnregisterCamera() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.controlCh = nil
}

// Subscribe registers an SSE client and returns a channel of JPEG frames.
// Sends START to the camera if this is the first subscriber.
func (b *Broadcaster) Subscribe(id string) <-chan []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	if existing, ok := b.clients[id]; ok {
		close(existing)
	}
	ch := make(chan []byte, 4)
	b.clients[id] = ch
	if len(b.clients) == 1 && b.controlCh != nil {
		b.sendControl(&dogcampb.ControlMessage{
			Command:         dogcampb.ControlMessage_START,
			FrameIntervalMs: b.frameIntervalMs,
		})
	}
	return ch
}

// Unsubscribe removes an SSE client. Sends STOP to the camera when the last
// subscriber leaves.
func (b *Broadcaster) Unsubscribe(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if ch, ok := b.clients[id]; ok {
		close(ch)
		delete(b.clients, id)
	}
	if len(b.clients) == 0 && b.controlCh != nil {
		b.sendControl(&dogcampb.ControlMessage{Command: dogcampb.ControlMessage_STOP})
	}
}

// Publish fans a JPEG frame out to all connected SSE clients.
// Slow clients drop the frame rather than blocking.
func (b *Broadcaster) Publish(frame []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, ch := range b.clients {
		select {
		case ch <- frame:
		default:
		}
	}
}

// sendControl sends a message to the camera's control channel without blocking.
// Must be called with b.mu held.
func (b *Broadcaster) sendControl(msg *dogcampb.ControlMessage) {
	select {
	case b.controlCh <- msg:
	default:
	}
}
