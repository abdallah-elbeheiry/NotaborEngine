package notasound

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

type Sound struct {
	data []byte
}

type AudioFormat int

const (
	MP3 AudioFormat = iota
	WAV
	OGG
)

func newOtoContext(sampleRate int) (*oto.Context, <-chan struct{}, error) {
	ctx, ready, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate:   sampleRate,
		ChannelCount: 2,
		Format:       oto.FormatFloat32LE,
	})
	return ctx, ready, err
}

func load(path string, format AudioFormat) (*Sound, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var decoder io.Reader
	const sampleRate = 44100

	// Switch decoders based on format
	switch format {
	case MP3:
		decoder, err = mp3.DecodeWithSampleRate(sampleRate, f)
	case WAV:
		decoder, err = wav.DecodeWithSampleRate(sampleRate, f)
	case OGG:
		decoder, err = vorbis.DecodeWithSampleRate(sampleRate, f)
	default:
		return nil, fmt.Errorf("unsupported format")
	}

	if err != nil {
		return nil, err
	}

	// The rest of your conversion logic is correct for all three
	data, err := io.ReadAll(decoder)
	if err != nil {
		return nil, err
	}

	// Convert int16 (2 bytes) to float32 (4 bytes)
	// Note: len(data)*2 is correct because we go from 2 bytes/sample to 4 bytes/sample
	floatBytes := make([]byte, len(data)*2)
	for i := 0; i < len(data); i += 2 {
		raw := int16(data[i]) | int16(data[i+1])<<8
		f32 := float32(raw) / 32768.0
		u := math.Float32bits(f32)

		idx := i * 2
		floatBytes[idx+0] = byte(u)
		floatBytes[idx+1] = byte(u >> 8)
		floatBytes[idx+2] = byte(u >> 16)
		floatBytes[idx+3] = byte(u >> 24)
	}

	return &Sound{data: floatBytes}, nil
}

// play handles the lifecycle: creates player, plays, and closes when done.
func play(ctx *oto.Context, s *Sound) {
	go func() {
		// Create a reader for this specific playback instance
		r := bytes.NewReader(s.data)
		player := ctx.NewPlayer(r)

		player.Play()

		// Wait until the sound is finished
		for player.IsPlaying() {
			// Tiny sleep to prevent CPU spiking while checking status
			select {
			case <-time.After(time.Millisecond * 10):
			}
		}
	}()
}
