package publisher

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/rocketb/asperitas/pkg/logger"
)

type Collector interface {
	Collect() (map[string]any, error)
}

type Publisher func(map[string]any)

type Publish struct {
	log       *logger.Logger
	collector Collector
	publisher []Publisher
	wg        sync.WaitGroup
	timer     *time.Timer
	shutdown  chan struct{}
}

func New(log *logger.Logger, collector Collector, interval time.Duration, publisher ...Publisher) (*Publish, error) {
	p := Publish{
		log:       log,
		collector: collector,
		publisher: publisher,
		timer:     time.NewTimer(interval),
		shutdown:  make(chan struct{}),
	}

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		for {
			p.timer.Reset(interval)
			select {
			case <-p.timer.C:
				p.update()
			case <-p.shutdown:
				return
			}
		}
	}()

	return &p, nil
}

func (p *Publish) Stop() {
	close(p.shutdown)
	p.wg.Wait()
}

func (p *Publish) update() {
	data, err := p.collector.Collect()
	if err != nil {
		p.log.Error(context.Background(), "publish", "status", "collect data", "msg", err)
		return
	}

	for _, pub := range p.publisher {
		pub(data)
	}
}

type StdOut struct {
	log *logger.Logger
}

func NewStdout(log *logger.Logger) *StdOut {
	return &StdOut{log}
}

func (s *StdOut) Publish(data map[string]any) {
	ctx := context.Background()

	rawJSON, err := json.Marshal(data)
	if err != nil {
		s.log.Error(ctx, "stdout", "status", "marshal data", "msg", err)
		return
	}

	var d map[string]any
	if err := json.Unmarshal(rawJSON, &d); err != nil {
		s.log.Error(ctx, "stdout", "status", "unmarshal data", "msg", err)
		return
	}

	memStats, ok := (d["memstats"]).(map[string]any)
	if ok {
		d["heap"] = memStats["Alloc"]
	}

	// Remove unnecessary keys.
	delete(d, "memstats")
	delete(d, "cmdline")

	out, err := json.MarshalIndent(d, "", "    ")
	if err != nil {
		return
	}
	s.log.Info(ctx, "stdout", "data", string(out))
}
