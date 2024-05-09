package main

import (
	"context"
	"github.com/GlazeLab/PureGamer/src/modules/entry"
	"github.com/GlazeLab/PureGamer/src/modules/exit"
	"github.com/GlazeLab/PureGamer/src/modules/optimizer"
	"github.com/GlazeLab/PureGamer/src/modules/pinging"
	"github.com/GlazeLab/PureGamer/src/modules/relaying"
	"github.com/GlazeLab/PureGamer/src/modules/superadmin"
	"github.com/GlazeLab/PureGamer/src/node"
	logging "github.com/ipfs/go-log/v2"
)

func main() {
	ctx := context.TODO()
	n, err := node.Listen(ctx)
	if err != nil {
		panic(err)
	}

	logging.SetAllLoggers(logging.LevelInfo)
	admin, err := superadmin.NewSuperAdmin(n)
	if err != nil {
		panic(err)
	}
	exits, err := exit.NewExit(n)
	if err != nil {
		panic(err)
	}
	err = relaying.Register(n, exits)
	if err != nil {
		panic(err)
	}
	err = pinging.Register(n)
	if err != nil {
		panic(err)
	}
	optimized, err := optimizer.NewOptimizer(n)
	if err != nil {
		panic(err)
	}

	admin.Handle(ctx)
	optimized.Handle(ctx)
	go optimized.RunSpeedTest(ctx)

	err = entry.Listen(n, exits, optimized)
	if err != nil {
		panic(err)
	}
	select {}
}
