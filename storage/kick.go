package storage

import (
	"fmt"
	"time"
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

	_, err := o.rdb.Set(key, 1, time.Minute*5).Result()
	if err != nil {
		return err
	}

	return nil
}

func (o *InstanceObj) DelKicked(chatID int64, userID int) error {
	key := fmt.Sprintf("irma_kick_{%d_%d}", chatID, userID)

	_, err := o.rdb.Del(key).Result()
	if err != nil {
		return err
	}

	return nil
}
