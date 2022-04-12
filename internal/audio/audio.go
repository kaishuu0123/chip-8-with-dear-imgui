package audio

import (
	"math"

	"github.com/gordonklaus/portaudio"
)

type Audio struct {
	stream         *portaudio.Stream
	SampleRate     float64
	OutputChannels int
	Channel        chan float32

	StreamBufferLengthPerFrame int

	PhaseL float64
	PhaseR float64
}

func NewAudio() (*Audio, error) {
	err := portaudio.Initialize()
	if err != nil {
		return nil, err
	}

	a := Audio{}
	a.Channel = make(chan float32, 44100)
	return &a, nil
}

func (a *Audio) Close() {
	a.stream.Stop()
	a.stream.Close()
	portaudio.Terminate()
}

func (a *Audio) Start() error {
	host, err := portaudio.DefaultHostApi()
	if err != nil {
		return err
	}
	parameters := portaudio.HighLatencyParameters(nil, host.DefaultOutputDevice)
	stream, err := portaudio.OpenStream(parameters, a.Callback)
	if err != nil {
		return err
	}
	if err := stream.Start(); err != nil {
		return err
	}

	a.stream = stream
	a.SampleRate = parameters.SampleRate
	a.OutputChannels = parameters.Output.Channels

	a.StreamBufferLengthPerFrame = int((a.SampleRate * (1 / 60.0))) + 1

	a.PhaseL = 0
	return nil
}

func (a *Audio) Stop() error {
	return a.stream.Close()
}

func (a *Audio) Callback(out [][]float32) {
	var output float32 = 0
	for i := range out[0] {
		select {
		case sample := <-a.Channel:
			output = sample
		default:
			output = 0
		}
		out[0][i] = output
		out[1][i] = output
	}
}

func (a *Audio) OutSineWave() {
	tone := 440.0 // A4
	step := tone / a.SampleRate

	buffer := make([]float32, a.StreamBufferLengthPerFrame)
	var volume float32 = 0.3

	// Create Sine wave + Volume
	for i, _ := range buffer {
		sine := math.Sin(2.0 * math.Pi * a.PhaseL)
		_, a.PhaseL = math.Modf(a.PhaseL + step)
		buffer[i] = float32(sine) * volume
	}

	for i, _ := range buffer {
		a.Channel <- buffer[i]
	}
}
