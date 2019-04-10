package storage

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

func (o *InstanceObj) GetNewbieMessages(chatID int64, userID int) (int, error) {
	k := fmt.Sprintf("irma_newbie_{%d_%d}", chatID, userID)

	res, err := o.rdb.Get(k).Result()
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

func (o *InstanceObj) AddNewbieMessages(chatID int64, userID int) error {
	k := fmt.Sprintf("irma_newbie_{%d_%d}", chatID, userID)

	pipe := o.rdb.TxPipeline()

	pipe.Incr(k)
	pipe.Expire(k, time.Hour*24*30)

	_, err := pipe.Exec()
	if err != nil {
		return err
	}

	return nil
}
