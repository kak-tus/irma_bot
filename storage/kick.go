package storage

import (
	"context"
	"fmt"
	"time"
)

func (o *InstanceObj) IsKicked(ctx context.Context, chatID int64, userID int) (bool, error) {
	k := fmt.Sprintf("irma_kick_{%d_%d}", chatID, userID)

	res, err := o.rdb.Exists(ctx, k).Result()
	if err != nil {
		return false, err
	}

	if res == 0 {
		return false, nil
	}

	return true, nil
}

func (o *InstanceObj) SetKicked(ctx context.Context, chatID int64, userID int, ttl time.Duration) error {
	key := fmt.Sprintf("irma_kick_{%d_%d}", chatID, userID)

	_, err := o.rdb.Set(ctx, key, 1, ttl*2).Result()
	if err != nil {
		return err
	}

	return nil
}

func (o *InstanceObj) DelKicked(ctx context.Context, chatID int64, userID int64) error {
	key := fmt.Sprintf("irma_kick_{%d_%d}", chatID, userID)

	_, err := o.rdb.Del(ctx, key).Result()
	if err != nil {
		return err
	}

	return nil
}
