package worker

import (
	"context"
	"time"

	"backend-challenge/apps/common/pkg/ctxutils"
	"backend-challenge/apps/common/pkg/logger"
	"backend-challenge/apps/gateway-api/internal/usecase"
)

type UserCountWorker struct {
	useCase  usecase.UserUseCase
	logger   logger.Logger
	interval time.Duration
}

func NewUserCountWorker(useCase usecase.UserUseCase, log logger.Logger, interval time.Duration) *UserCountWorker {
	if interval <= 0 {
		interval = 10 * time.Second
	}
	return &UserCountWorker{
		useCase:  useCase,
		logger:   log,
		interval: interval,
	}
}

func (w *UserCountWorker) Start(ctx context.Context) {
	w.logger.Info("Starting UserCountWorker", logger.Duration("interval", w.interval))
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Stopping UserCountWorker...")
			return
		case <-ticker.C:
			countCtx := ctxutils.SetRequestID(context.Background(), "worker-user-count-"+time.Now().Format("150405"))
			countCtx = ctxutils.SetAuthUser(countCtx, "system", "system@internal")
			countCtx, cancel := context.WithTimeout(countCtx, 5*time.Second)
			count, err := w.useCase.CountUsers(countCtx)
			cancel()
			if err != nil {
				w.logger.Error("Failed to count users", logger.Err(err))
			} else {
				w.logger.Info("Total users count in DB", logger.Int64("total_users", count))
			}
		}
	}
}
