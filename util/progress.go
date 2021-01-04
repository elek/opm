package util

import (
	"github.com/rs/zerolog/log"
	"time"
)

type Progress struct {
	start   time.Time
	lastLog time.Time
	counter int
}

func CreateProgress() *Progress {
	now := time.Now()
	return &Progress{
		start:   now,
		lastLog: now,
		counter: 0,
	}
}
func (p *Progress) Increment() {
	p.counter++
	if time.Since(p.lastLog).Seconds() > 5 {
		p.lastLog = time.Now()
		log.Info().Msgf("Processed %d iteration with the average speed: %f/sec", p.counter, float64(p.counter)/time.Since(p.start).Seconds())
	}
}

func (p *Progress) End() {
	seconds := time.Since(p.start).Seconds()
	log.Info().Msgf("Processed %d iteration under %f seconds,  with the average speed: %f/sec", p.counter, seconds, float64(p.counter)/seconds)
}
