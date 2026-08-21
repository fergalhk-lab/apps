package capture

import (
	"fmt"
	"os/exec"
	"strings"
)

// Capturer captures a single JPEG frame.
type Capturer interface {
	Capture() ([]byte, error)
}

// FFmpegCapturer captures from a v4l2 device (USB webcam).
type FFmpegCapturer struct {
	device string
}

func NewFFmpeg(device string) *FFmpegCapturer {
	return &FFmpegCapturer{device: device}
}

func (c *FFmpegCapturer) Capture() ([]byte, error) {
	return exec.Command(
		"ffmpeg", "-y",
		"-f", "v4l2", "-i", c.device,
		"-frames:v", "1",
		"-f", "image2pipe", "-vcodec", "mjpeg",
		"pipe:1",
	).Output()
}

// LibcameraCapturer captures using libcamera-still (Pi camera module).
type LibcameraCapturer struct{}

func NewLibcamera() *LibcameraCapturer {
	return &LibcameraCapturer{}
}

func (c *LibcameraCapturer) Capture() ([]byte, error) {
	return exec.Command(
		"libcamera-still", "--output", "-", "--codec", "jpeg", "--immediate", "--nopreview",
	).Output()
}

// New returns the appropriate Capturer for the given device string.
// Use "libcamera" for Pi camera modules, or a /dev/videoN path for USB webcams.
func New(device string) (Capturer, error) {
	switch {
	case device == "libcamera":
		return NewLibcamera(), nil
	case strings.HasPrefix(device, "/dev/video"):
		return NewFFmpeg(device), nil
	default:
		return nil, fmt.Errorf("unknown camera device %q: use 'libcamera' or '/dev/videoN'", device)
	}
}
