package background

import (
	"fmt"
	"log/slog"
	"sync"
)

func Run(wg *sync.WaitGroup, logger *slog.Logger, fn func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic recovering", "err", fmt.Errorf("%v", err))
			}
		}()
		fn()
	}()
}
