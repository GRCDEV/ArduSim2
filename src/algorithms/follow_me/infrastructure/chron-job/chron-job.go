package chronjob

import "time"

type ChronJob struct {
	fn     func()
	ticker *time.Ticker
	period time.Duration
	stop   chan any
}

func NewChronJob(fn func(), period time.Duration) *ChronJob {
	return &ChronJob{
		fn:     fn,
		period: period,
		stop:   make(chan any, 1),
	}
}

func (c *ChronJob) loop() {
	for {
		select {
		case <-c.stop:
			return
		case <-c.ticker.C:
			c.fn()
		}
	}
}

func (c *ChronJob) Start() {
	if c.ticker == nil {
		c.ticker = time.NewTicker(c.period)
		go c.loop()
	}
	c.ticker.Reset(c.period)
}

func (c *ChronJob) Pause() {
	c.ticker.Stop()
}

func (c *ChronJob) Stop() {
	c.ticker.Stop()
	c.ticker = nil
	c.stop <- nil
}
