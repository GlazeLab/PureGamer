package optimizer

import (
	"context"
	"github.com/GlazeLab/PureGamer/src/model"
	"github.com/vmihailenco/msgpack/v5"
)

func (o *Optimizer) Handle(ctx context.Context) <-chan error {
	errCh := make(chan error)
	go func() {
		for {
			msg, err := o.sub.Next(ctx)
			if err != nil {
				log.Error(err)
				errCh <- err
				continue
			}
			var latencies model.Latencies
			err = msgpack.Unmarshal(msg.GetData(), &latencies)
			if err != nil {
				log.Error(err)
				errCh <- err
				continue
			}
			fromNode := msg.GetFrom().String()

			existEdges := o.gr.IterateEdges(fromNode)
			for _, to := range existEdges {
				if latency, ok := latencies[to]; ok {
					o.gr.AddEdge(fromNode, to, latency)
					delete(latencies, to)
				} else {
					o.gr.RemoveEdge(fromNode, to)
					edgesToBeRemoved := o.gr.IterateEdges(to)
					for _, edge := range edgesToBeRemoved {
						o.gr.RemoveEdge(to, edge)
					}
				}
			}
			for to, latency := range latencies {
				o.gr.AddEdge(fromNode, to, latency)
			}

		}
	}()
	return errCh
}

func (o *Optimizer) OptimizedRoutes(entry string, exit string) []string {
	log.Infof("Entry: %s, Exit: %s", entry, exit)
	log.Infof("%v", o.gr.IterateEdges(entry))
	routes, dist := o.gr.ShortestPath(entry, exit)
	log.Infof("Optimized: %f", dist)
	return routes
}
