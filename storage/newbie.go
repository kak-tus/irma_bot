package storage

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

// 2 months
const storeNewbieFirstMessageTTL = time.Hour * 24 * 40 * 2

func (o *InstanceObj) GetNewbieMessages(ctx context.Context, chatID int64, userID int) (int, error) {
	k := fmt.Sprintf("irma_newbie_{%d_%d}", chatID, userID)

	res, err := o.rdb.Get(ctx, k).Result()
	if err != nil && err != redis.Nil {
		return 0, err
	} else if err != nil && err == redis.Nil {
		return 0, nil
	}

	i, err := strconv.Atoi(res)
	if err != nil {
		return 0, err
	}

	return i, nil
}

func (o *InstanceObj) AddNewbieMessages(ctx context.Context, chatID int64, userID int) error {
	k := fmt.Sprintf("irma_newbie_{%d_%d}", chatID, userID)

	pipe := o.rdb.TxPipeline()

	pipe.Incr(ctx, k)
	pipe.Expire(ctx, k, storeNewbieFirstMessageTTL)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}
