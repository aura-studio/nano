package filepersist

import (
	"encoding/json"
	"os"
)

type FilePersist struct {
	path string
}

func NewFilePersist(path string) *FilePersist {
	return &FilePersist{
		path: path,
	}
}

func (fp *FilePersist) Set(v interface{}) error {
	if b, err := json.Marshal(v); err != nil {
		return err
	} else {
		return os.WriteFile(fp.path, b, 0644)
	}
}

func (fp *FilePersist) Get(v interface{}) error {
	if !existsFile(fp.path) {
		if err := os.WriteFile(fp.path, []byte("null"), 0644); err != nil {
			return err
		}
	}

	if b, err := os.ReadFile(fp.path); err != nil {
		return err
	} else {
		return json.Unmarshal(b, v)
	}
}

func existsFile(f string) bool {
	if _, err := os.Stat(f); err != nil {
		return os.IsExist(err)
	}
	return true
}
