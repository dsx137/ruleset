package util

import (
	"errors"
	"fmt"
	"runtime/debug"

	"golang.org/x/sync/errgroup"
)

func Parallel(limit int, feeder func(func(func() error)) (err error)) error {
	if limit < 1 {
		return fmt.Errorf("invalid concurrency limit: %d", limit)
	}

	var g errgroup.Group
	g.SetLimit(limit)

	feed := func(task func() error) {
		g.Go(func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("task panicked: %v\n%s", r, debug.Stack())
				}
			}()
			if task == nil {
				return errors.New("nil task")
			}
			return task()
		})
	}

	var feederErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				feederErr = fmt.Errorf("feeder panicked: %v\n%s", r, debug.Stack())
			}
		}()
		feederErr = feeder(feed)
	}()

	waitErr := g.Wait()

	return errors.Join(feederErr, waitErr)
}
