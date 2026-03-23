package notasound

import (
	"errors"
	"sync"
	"time"

	"github.com/ebitengine/oto/v3"
)

type SoundManager struct {
	ctx          *oto.Context
	ready        <-chan struct{}
	cache        sync.Map // Map[string]*Sound
	activeSounds sync.Map // Map[string]*oto.Player (for stopping/volume)

	soundsFolder string
	folderGiven  bool
	MasterVolume float32 // Linked to Engine Settings
	Mute         bool
}

// NewSoundManager creates a new SoundManager
func NewSoundManager(masterVolume float32, mute bool) (*SoundManager, error) {
	ctx, ready, err := newOtoContext(44100)
	if err != nil {
		return nil, err
	}

	return &SoundManager{
		ctx:          ctx,
		ready:        ready,
		MasterVolume: masterVolume,
		Mute:         mute,
	}, nil
}

func (m *SoundManager) SetSoundsFolder(path string) {
	m.soundsFolder = path
	m.folderGiven = true
}

func (m *SoundManager) Play(sound string, format AudioFormat, volume float32, loop bool) error {
	if m.Mute {
		return nil
	}
	if !m.folderGiven {
		return errors.New("SoundManager: sounds folder not set")
	}
	fullPath := m.soundsFolder + "/" + sound

	if _, alreadyPlaying := m.activeSounds.Load(sound); alreadyPlaying {
		m.Stop(sound)
	}

	go func() {
		<-m.ready

		var s *Sound
		val, ok := m.cache.Load(fullPath)
		if ok {
			s = val.(*Sound)
		} else {
			var err error
			s, err = load(fullPath, format)
			if err != nil {
				return
			}
			m.cache.Store(fullPath, s)
		}

		// Calculate combined volume
		p := play(m.ctx, s, m.MasterVolume*volume)
		m.activeSounds.Store(sound, p)

		if loop {
			for {
				if _, stillActive := m.activeSounds.Load(sound); !stillActive {
					return
				}

				if !p.IsPlaying() {
					_, _ = p.Seek(0, 0)
					p.Play()
				}
				time.Sleep(10 * time.Millisecond)
			}
		}
		m.activeSounds.Store(sound, p)
	}()
	return nil
}

func (m *SoundManager) Stop(sound string) {
	if val, ok := m.activeSounds.Load(sound); ok {
		p := val.(*oto.Player)
		p.Pause()
		m.activeSounds.Delete(sound)
		_, _ = p.Seek(0, 0)
	}
}
