package commerce_service

import (
	"context"
	"time"
)

type NotificationDispatcher interface{ DispatchNotifications(context.Context) error }

func RunOutbox(ctx context.Context, dispatcher NotificationDispatcher, onError func(error)) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		if err := dispatcher.DispatchNotifications(ctx); err != nil && ctx.Err() == nil {
			onError(err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
