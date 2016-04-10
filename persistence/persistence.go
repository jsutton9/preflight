package persistence

import (
	"encoding/json"
	"github.com/jsutton9/todoistist/config"
	"io/ioutil"
	"os"
	"time"
)

type UpdateRecord struct {
	Uuids []string   `json:"uuids"`
	Timestamp string `json:"time"`
	Time time.Time   `json:"-"`
}

type Persister struct {
	Path string                           `json:"-"`
	Config config.Config                  `json:"config"`
	UpdateHistory map[string]UpdateRecord `json:"updateHistory"`
}

func Load(user string) (Persister, error) {
	dir := "/var/lib/todoistist/"
	p := Persister{Path:dir+user+".json"}

	data, err := ioutil.ReadFile(p.Path)
	if err != nil {
		if os.IsNotExist(err) {
			p.UpdateHistory = make(map[string]UpdateRecord)
			return p, nil
		} else {
			return p, err
		}
	}
	err = json.Unmarshal(data, &p)
	if err != nil {
		return p, err
	}

	for _, record := range p.UpdateHistory {
		if record.Timestamp != "" {
			record.Time, err = time.Parse("2006-01-02T15:04:05-0700", record.Timestamp)
			if err != nil {
				return p, err
			}
		}
	}

	return p, nil
}

func (p Persister) Save() error {
	for _, record := range p.UpdateHistory {
		record.Timestamp = record.Time.Format("2006-01-02T15:04:05-0700")
	}

	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(p.Path, data, 0755)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll("/var/lib/todoistist", 0755)
			if err != nil {
				return err
			}
			err = ioutil.WriteFile(p.Path, data, 0755)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}
