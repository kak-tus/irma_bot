package storage

import (
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

type Action struct {
	ChatID    int64
	MessageID int
	Type      string
	UserID    int
}

func (o *InstanceObj) AddToActionPool(act Action) error {
	val, err := o.enc.MarshalToString(act)
	if err != nil {
		return err
	}

	sc := time.Now().Add(time.Minute).In(time.UTC).Unix()

	z := redis.Z{
		Member: val,
		Score:  float64(sc),
	}

	_, err = o.rdb.ZAdd("irma_kick_pool", z).Result()
	if err != nil {
		return err
	}

	return nil
}

func (o *InstanceObj) GetFromActionPool() ([]Action, error) {
	sc := time.Now().In(time.UTC).Unix()

	pool := o.rdb.TxPipeline()

	z := redis.ZRangeBy{
		Max: strconv.Itoa(int(sc)),
		Min: "0",
	}

	rangeCmd := pool.ZRangeByScore("irma_kick_pool", z)
	_ = pool.ZRemRangeByScore("irma_kick_pool", z.Min, z.Max)

	_, err := pool.Exec()
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
