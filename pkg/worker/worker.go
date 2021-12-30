package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pkg/errors"
)

const (
	Threads = 30
)

type Config struct {
	Threads int
}

func (c Config) GetWithDefault() *Config {
	if c.Threads == 0 {
		c.Threads = Threads
	}

	return &c
}

func New(cfg *Config) *services {
	c := cfg.GetWithDefault()

	return &services{cfg: c, ch: make(chan struct{}, c.Threads)}
}

type services struct {
	cfg *Config

	ch chan struct{}
}

func (s *services) Run(ctx context.Context, fn func() error) {
	counter := make([]uint, s.cfg.Threads)

	for i := 0; i < s.cfg.Threads; i++ {
		go s.work(ctx, i, &counter[i], fn)
	}

	s.counter(ctx, time.Now(), counter)
}

func (s *services) counter(ctx context.Context, start time.Time, counter []uint) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
			ms := time.Now().Sub(start)

			var i uint
			for _, v := range counter {
				i += v
			}

			q := float64(i) / ms.Seconds()
			fmt.Println("comb/sec:", q, "duration:", ms.Seconds(), i)
		}
	}
}

func (s *services) work(ctx context.Context, i int, c *uint, fn func() error) {
	log.Printf("[%d] worker start", i)

	defer func() {
		log.Printf("[%d] worker exit", i)
		s.ch <- struct{}{}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := fn(); err != nil {
			log.Panicf("worker fn %+v", errors.WithStack(err))
		}

		*c++
	}
}

func (s *services) Wait() {
	for i := 0; i < s.cfg.Threads; i++ {
		<-s.ch
	}

	close(s.ch)
}
