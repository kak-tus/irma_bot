package storage

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type ActionType string

const (
	ActionTypeDelete ActionType = "del"
	ActionTypeKick   ActionType = "kick"
)

type Action struct {
	ChatID    int64
	MessageID int
	Type      ActionType
	UserID    int
}

func (o *InstanceObj) AddToActionPool(ctx context.Context, act Action, delay time.Duration) error {
	val, err := o.enc.MarshalToString(act)
	if err != nil {
		return err
	}

	sc := time.Now().Add(delay).In(time.UTC).Unix()

	z := redis.Z{
		Member: val,
		Score:  float64(sc),
	}

	_, err = o.rdb.ZAdd(ctx, "irma_kick_pool", &z).Result()
	if err != nil {
		return err
	}

	return nil
}

func (o *InstanceObj) GetFromActionPool(ctx context.Context) ([]Action, error) {
	sc := time.Now().In(time.UTC).Unix()

	pool := o.rdb.TxPipeline()

	z := redis.ZRangeBy{
		Max: strconv.Itoa(int(sc)),
		Min: "0",
	}

	rangeCmd := pool.ZRangeByScore(ctx, "irma_kick_pool", &z)
	_ = pool.ZRemRangeByScore(ctx, "irma_kick_pool", z.Min, z.Max)

	_, err := pool.Exec(ctx)
	if err != nil {
		return nil, err
	}

	acts := make([]Action, 0)

	for _, v := range rangeCmd.Val() {
		var a Action

		err := o.enc.UnmarshalFromString(v, &a)
		if err != nil {
			return nil, err
		}

		acts = append(acts, a)
	}

	return acts, nil
}
