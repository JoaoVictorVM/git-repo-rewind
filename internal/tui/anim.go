package tui

import (
	"math"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/harmonica"
)

const (
	animFPS       = 60
	animFrequency = 6.0
	animDamping   = 1.0
	settleEpsilon = 0.5
)

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Second/animFPS, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type counterAnim struct {
	spring harmonica.Spring
	pos    float64
	vel    float64
}

func newCounterAnim() counterAnim {
	return counterAnim{spring: harmonica.NewSpring(harmonica.FPS(animFPS), animFrequency, animDamping)}
}

func (a *counterAnim) update(target float64) {
	a.pos, a.vel = a.spring.Update(a.pos, a.vel, target)
}

func (a *counterAnim) settled(target float64) bool {
	return math.Abs(target-a.pos) < settleEpsilon && math.Abs(a.vel) < settleEpsilon
}

func (a *counterAnim) snap(target float64) {
	a.pos, a.vel = target, 0
}

func (a counterAnim) value() int {
	v := int(math.Round(a.pos))
	if v < 0 {
		return 0
	}
	return v
}
