package persistence

import (
	"time"
)

type Persister {
	ConfigPath string,
	UpdateHistory map[string][time.Time],
}

func Load() Persister {
}

func (p Persister) Save() {
}
