package persistence

import (
	"ioutil"
	"time"
)

type UpdateRecord {
	Uuids []string   `json:"uuids"`
	Timestamp string `json:"time"`
	Time time.Time   `json:"-"`
}

type Persister {
	Path string                             `json:"-"`
	Config config.Config                    `json:"config"`
	UpdateHistory map[string][UpdateRecord] `json:"updateHistory"`
}

func Load() Persister {
}

func (p Persister) Save() {
}
