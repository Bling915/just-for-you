package moment

import (
	"context"
	appshared "github.com/nanfang/just-for-you/server/internal/application/shared"
	domain "github.com/nanfang/just-for-you/server/internal/domain/moment"
	"time"
)

func signReadURLs(ctx context.Context, moments []domain.Moment, photos PhotoStorage, clock appshared.Clock) (map[string]string, error) {
	result := map[string]string{}
	expires := clock.Now().Add(10 * time.Minute)
	for _, m := range moments {
		for _, p := range m.Photos() {
			if _, ok := result[p.ObjectKey()]; ok {
				continue
			}
			url, err := photos.SignRead(ctx, p.ObjectKey(), expires)
			if err != nil {
				return nil, err
			}
			result[p.ObjectKey()] = url
		}
	}
	return result, nil
}
