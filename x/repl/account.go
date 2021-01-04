package repl

import (
	"github.com/go-redis/redis"

	"github.com/lonng/nano/serialize/json"
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
	data, err := client.HGet(cliAccountKey, a.Username).Result()
	if err != redis.Nil && err != nil {
		return isNew, err
	}

	if err == redis.Nil {
		isNew = true
		JSONData, err := json.NewSerializer().Marshal(currentAccount)
		if err != nil {
			return isNew, err
		}
		_, err = client.HSet(cliAccountKey, a.Username, string(JSONData)).Result()
		if err != nil {
			return isNew, err
		}
		logger.Printf("new cli account: %s created", a.Username)
		logger.Printf("%s logined successfully\n", a.Username)
		return isNew, nil
	}
	isNew = false
	var tmpAccount Account
	err = json.NewSerializer().Unmarshal([]byte(data), &tmpAccount)
	if err != nil {
		return isNew, err
	}

	currentAccount = &tmpAccount
	logger.Printf("%s logined successfully\n", a.Username)
	return isNew, nil
}

// Save saves account
func (a *Account) Save() error {
	JSONData, err := json.NewSerializer().Marshal(a)
	if err != nil {
		return err
	}
	client, err := getRedisClient()
	if err != nil {
		return err
	}
	_, err = client.HSet(cliAccountKey, a.Username, string(JSONData)).Result()
	if err != nil {
		return err
	}
	return nil
}
