package storage

import (
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

	sc := time.Now().In(time.UTC).Unix()

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
