package optimizer

import (
	"context"
	"github.com/vmihailenco/msgpack/v5"
	"time"
)

func (o *Optimizer) RunSpeedTest(ctx context.Context) {
	log.Info("Start speed test")
	ticker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-ticker.C:
			latencies := SpeedTest(ctx, o.n)
			log.Infof("Latencies: %v", latencies)
			latenciesBytes, err := msgpack.Marshal(latencies)
			if err != nil {
				log.Error(err)
				continue
			}
			err = o.top.Publish(ctx, latenciesBytes)
			if err != nil {
				log.Error(err)
				continue
			}
		case <-ctx.Done():
			log.Info("Stop speed test")
			return
		}
	}
}
