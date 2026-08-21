package capture_test

import (
	"testing"

	"github.com/fergalhk-lab/apps/dogcam/cam/internal/capture"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// FakeCapturer satisfies the Capturer interface for tests.
type FakeCapturer struct {
	Frames [][]byte
	idx    int
	Err    error
}

func (f *FakeCapturer) Capture() ([]byte, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	frame := f.Frames[f.idx%len(f.Frames)]
	f.idx++
	return frame, nil
}

func TestFakeCapturer_ImplementsInterface(t *testing.T) {
	var _ capture.Capturer = &FakeCapturer{}
}

func TestFFmpegCapturer_ImplementsInterface(t *testing.T) {
	var _ capture.Capturer = capture.NewFFmpeg("/dev/video0")
}

func TestLibcameraCapturer_ImplementsInterface(t *testing.T) {
	var _ capture.Capturer = capture.NewLibcamera()
}

func TestNewCapturer_FFmpegForV4L2Device(t *testing.T) {
	c, err := capture.New("/dev/video0")
	require.NoError(t, err)
	assert.IsType(t, &capture.FFmpegCapturer{}, c)
}

func TestNewCapturer_LibcameraForLibcameraDevice(t *testing.T) {
	c, err := capture.New("libcamera")
	require.NoError(t, err)
	assert.IsType(t, &capture.LibcameraCapturer{}, c)
}

func TestNewCapturer_ErrorForUnknownDevice(t *testing.T) {
	_, err := capture.New("unknown-device")
	require.Error(t, err)
}
