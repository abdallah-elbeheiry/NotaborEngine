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
	MasterVolume float32 // Linked to Engine settings
	Mute         bool
}

type activeTrack struct {
	player      *oto.Player
	localVolume float32
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

// SetSoundsFolder configures the base folder used by Play when resolving sound file names.
func (m *SoundManager) SetSoundsFolder(path string) {
	m.soundsFolder = path
	m.folderGiven = true
}

// Play starts a sound if it is not already active.
func (m *SoundManager) Play(sound string, format AudioFormat, volume float32, loop bool) error {
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

		soundVolume := volume * m.MasterVolume
		if m.Mute {
			soundVolume = 0
		}
		p := play(m.ctx, s, soundVolume)
		track := &activeTrack{
			player:      p,
			localVolume: volume,
		}
		m.activeSounds.Store(sound, track)

		if loop {
			for {
				val, stillActive := m.activeSounds.Load(sound)
				if !stillActive || val != track {
					closePlayer(track.player)
					return
				}

				if !p.IsPlaying() {
					_, _ = p.Seek(0, 0)
					p.Play()
				}
				time.Sleep(10 * time.Millisecond)
			}
		}

		for p.IsPlaying() {
			time.Sleep(10 * time.Millisecond)
		}
		m.removeTrack(sound, track)
	}()
	return nil
}

// Stop halts a playing sound and releases its active playback state.
func (m *SoundManager) Stop(sound string) {
	if val, ok := m.activeSounds.Load(sound); ok {
		if track, ok := val.(*activeTrack); ok {
			track.player.Pause()
			_, _ = track.player.Seek(0, 0)
			closePlayer(track.player)
		}
		m.activeSounds.Delete(sound)
	}
}

// UpdateLiveVolume reapplies master and mute state to all currently playing sounds.
func (m *SoundManager) UpdateLiveVolume() {
	m.activeSounds.Range(func(key, value any) bool {
		if track, ok := value.(*activeTrack); ok {
			vol := float64(track.localVolume * m.MasterVolume)
			if m.Mute {
				vol = 0
			}
			track.player.SetVolume(vol)
		}
		return true
	})
}

func (m *SoundManager) removeTrack(sound string, track *activeTrack) {
	if current, ok := m.activeSounds.Load(sound); ok && current == track {
		m.activeSounds.Delete(sound)
	}
	closePlayer(track.player)
}

func closePlayer(player *oto.Player) {
	if player == nil {
		return
	}
	type closer interface {
		Close() error
	}
	if c, ok := any(player).(closer); ok {
		_ = c.Close()
	}
}
