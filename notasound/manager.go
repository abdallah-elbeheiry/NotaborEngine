package notasound

import (
	"errors"
	"fmt"
	"sync"

	"github.com/ebitengine/oto/v3"
)

type SoundManager struct {
	ctx          *oto.Context
	ready        <-chan struct{}
	cache        sync.Map // Map[string]*Sound
	soundsFolder string
	folderGiven  bool
}

// NewSoundManager creates a new SoundManager
func NewSoundManager() (*SoundManager, error) {
	ctx, ready, err := newOtoContext(44100)
	if err != nil {
		return nil, err
	}

	return &SoundManager{
		ctx:   ctx,
		ready: ready,
	}, nil
}

func (m *SoundManager) SetSoundsFolder(path string) {
	m.soundsFolder = path
	m.folderGiven = true
}

func (m *SoundManager) Play(sound string, format AudioFormat) error {
	if !m.folderGiven {
		return errors.New("SoundManager: sounds folder not set, use SetSoundsFolder() first")
	}
	if m.soundsFolder != "" {
		sound = m.soundsFolder + "/" + sound
	}

	go func() {
		<-m.ready

		var s *Sound
		val, ok := m.cache.Load(sound)

		if ok {
			s = val.(*Sound)
		} else {
			var err error
			s, err = load(sound, format)
			if err != nil {
				fmt.Printf("SoundManager Error: could not load %s: %v\n", sound, err)
				return
			}
			m.cache.Store(sound, s)
		}

		play(m.ctx, s)
	}()
	return nil
}
