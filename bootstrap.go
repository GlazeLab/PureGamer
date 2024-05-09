package main

import (
	"context"
	"github.com/GlazeLab/PureGamer/src/modules/superadmin"
	"github.com/GlazeLab/PureGamer/src/node"
	logging "github.com/ipfs/go-log/v2"
)

func main() {
	ctx := context.TODO()
	n, err := node.Listen(ctx)
	logging.SetAllLoggers(logging.LevelInfo)
	if err != nil {
		panic(err)
	}
	su, err := superadmin.NewSuperAdmin(n)
	if err != nil {
		panic(err)
	}
	errs := su.Handle(ctx)
	for {
		select {
		case err := <-errs:
			panic(err)
		case <-ctx.Done():
			return
		}
	}
}
