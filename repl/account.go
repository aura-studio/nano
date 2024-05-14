package repl

import (
	"context"
	"encoding/json"

	"github.com/aura-studio/nano/log"
	"github.com/redis/go-redis/v9"
)

// Account is cli's account
type Account struct {
	Username       string                       `json:"Username"`
	Aliases        map[string][]string          `json:"Aliases"`
	CmdSets        map[string]map[string]string `json:"CmdSets"`
	CurrentSet     string                       `json:"CurrentSet"`
	CurrentSetType int                          `json:"CurrentSetType"`
}

// Load loads account
func (a *Account) Load() (bool, error) {
	var isNew bool
	client, err := getRedisClient()
	if err != nil {
		return isNew, err
	}
	data, err := client.HGet(context.Background(), cliAccountKey, a.Username).Result()
	if err != redis.Nil && err != nil {
		return isNew, err
	}

	if err == redis.Nil {
		isNew = true
		JSONData, err := json.Marshal(currentAccount)
		if err != nil {
			return isNew, err
		}
		_, err = client.HSet(context.Background(), cliAccountKey, a.Username, string(JSONData)).Result()
		if err != nil {
			return isNew, err
		}
		log.Printf("new cli account: %s created", a.Username)
		log.Printf("%s logined successfully\n", a.Username)
		return isNew, nil
	}
	isNew = false
	var tmpAccount Account
	err = json.Unmarshal([]byte(data), &tmpAccount)
	if err != nil {
		return isNew, err
	}

	currentAccount = &tmpAccount
	log.Printf("%s logined successfully\n", a.Username)
	return isNew, nil
}

// Save saves account
func (a *Account) Save() error {
	JSONData, err := json.Marshal(a)
	if err != nil {
		return err
	}
	client, err := getRedisClient()
	if err != nil {
		return err
	}
	_, err = client.HSet(context.Background(), cliAccountKey, a.Username, string(JSONData)).Result()
	if err != nil {
		return err
	}
	return nil
}
