package redimopersist

import (
	"encoding/json"

	"github.com/aura-studio/redimo"
)

type RedimoPersist struct {
	*redimo.Client
	DDBKey string
}

func NewRedimoPersist(c *redimo.Client, ddbKey string) *RedimoPersist {
	return &RedimoPersist{
		Client: c,
		DDBKey: ddbKey,
	}
}

func (rp *RedimoPersist) Set(v interface{}) error {
	if b, err := json.Marshal(v); err != nil {
		return err
	} else {
		_, err := rp.SET(rp.DDBKey, string(b))
		if err != nil {
			return err
		}
	}
	return nil
}

func (rp *RedimoPersist) Get(v interface{}) error {
	if val, err := rp.GET(rp.DDBKey); err != nil {
		return err
	} else {
		return json.Unmarshal([]byte(val.String()), v)
	}
}
