package storage

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

func (o *InstanceObj) IsKicked(chatID int64, userID int) (bool, error) {
	k := fmt.Sprintf("irma_kick_{%d_%d}", chatID, userID)

	res, err := o.rdb.Exists(k).Result()
	if err != nil {
		return false, err
	}

	if res == 0 {
		return false, nil
	}

	return true, nil
}

func (o *InstanceObj) SetKicked(chatID int64, userID int) error {
	key := fmt.Sprintf("irma_kick_{%d_%d}", chatID, userID)

	_, err := o.rdb.Set(key, 1, time.Minute*10).Result()
	if err != nil {
		return err
	}

	return nil
}

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
