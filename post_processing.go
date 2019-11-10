package main

import (
	"context"
	"fmt"
	"time"
)

type PostProcessorConfig struct {
	Count    int
	Interval time.Duration
}

// Cancelator contains func to cancel N lattest ODD records
type Cancelator interface {
	Cancel(context.Context, int) error
}

func NewPostProcessor(ctx context.Context, count int, interval time.Duration, manager Cancelator) Runnable {
	return &postProcessor{
		ctx:          ctx,
		interval:     interval,
		cancelNumber: count,
		manager:      manager,
	}
}

type postProcessor struct {
	interval     time.Duration
	ctx          context.Context
	cancelNumber int
	manager      Cancelator
}

func (pp *postProcessor) Run() {

	defer func() {
		if r := recover(); r != nil {
			print(fmt.Sprintf("Recovered in post-processing %v", r))
			go pp.Run()
		}
	}()

	print("Starting post-processing system")

	ticker := time.NewTicker(pp.interval)
	defer ticker.Stop()
	for {
		select {
		case <-pp.ctx.Done():
			return
		case <-ticker.C:
			if err := pp.cancelRecords(); err != nil {
				print(fmt.Sprintf("failed to cancel %v latest records %v", pp.cancelNumber, err))
			}
		}
	}

}

func (pp *postProcessor) cancelRecords() error {
	println("Cancel latest records started")
	err := pp.manager.Cancel(pp.ctx, pp.cancelNumber)
	println("Cancel latest records done")
	return err
}
