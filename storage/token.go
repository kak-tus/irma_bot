package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	tokenKey = "irma_token_{%s}"
	tokenTTL = time.Minute * 10
)

type TokenData struct {
	ChatID int64
	TTL    time.Duration
}

func (o *InstanceObj) GetTokenData(ctx context.Context, token string) (TokenData, error) {
	key := fmt.Sprintf(tokenKey, token)

	res, err := o.rdb.Get(ctx, key).Int64()
	if err != nil {
		if err == redis.Nil {
			return TokenData{}, nil
		}

		return TokenData{}, err
	}

	data := TokenData{
		ChatID: res,
	}

	resDur, err := o.rdb.TTL(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return TokenData{}, nil
		}

		return TokenData{}, err
	}

	data.TTL = resDur

	return data, nil
}

func (o *InstanceObj) NewToken(ctx context.Context, chatID int64) (string, error) {
	token := o.gen.Generate()

	key := fmt.Sprintf(tokenKey, token)

	err := o.rdb.Set(ctx, key, chatID, tokenTTL).Err()
	if err != nil {
		return "", err
	}

	return token, nil
}
