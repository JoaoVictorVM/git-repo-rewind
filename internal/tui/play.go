package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const playInterval = 300 * time.Millisecond

type playTickMsg int

func playTick(generation int) tea.Cmd {
	return tea.Tick(playInterval, func(time.Time) tea.Msg {
		return playTickMsg(generation)
	})
}

func (m Model) togglePlay() (Model, tea.Cmd) {
	m.playing = !m.playing
	if !m.playing {
		return m, nil
	}
	m.playGen++
	return m, playTick(m.playGen)
}

func (m Model) advancePlay(generation int) (tea.Model, tea.Cmd) {
	if !m.playing || generation != m.playGen {
		return m, nil
	}

	next := m.engine.Step(m.cursor, m.granularity, true)
	if next.Equal(m.cursor) {
		m.playing = false
		return m, nil
	}

	moved, animCmd := m.moveCursor(next)
	return moved, tea.Batch(animCmd, playTick(moved.playGen))
}
