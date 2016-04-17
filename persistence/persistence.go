package persistence

import (
	"encoding/json"
	"github.com/jsutton9/todoistist/config"
	"io/ioutil"
	"os"
	"time"
)

type UpdateRecord struct {
	Ids []int           `json:"ids"`
	Timestamp string    `json:"time"`
	Time time.Time      `json:"-"`
	AddTimestamp string `json:"add_time"`
	AddTime time.Time   `json:"-"`
}

type Persister struct {
	Path string                           `json:"-"`
	Config config.Config                  `json:"config"`
	UpdateHistory map[string]UpdateRecord `json:"updateHistory"`
}

func Load(user string) (Persister, error) {
	dir := os.Getenv("HOME")+"/.todoistist/records/"
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

	for name, record := range p.UpdateHistory {
		if record.Timestamp != "" {
			record.Time, err = time.Parse("2006-01-02T15:04:05-0700", record.Timestamp)
			p.UpdateHistory[name] = record
			if err != nil {
				return p, err
			}
		}
		if record.AddTimestamp != "" {
			record.AddTime, err = time.Parse("2006-01-02T15:04:05-0700", record.AddTimestamp)
			p.UpdateHistory[name] = record
			if err != nil {
				return p, err
			}
		}
	}

	return p, nil
}

func (p Persister) Save() error {
	for key, record := range p.UpdateHistory {
		record.Timestamp = record.Time.Format("2006-01-02T15:04:05-0700")
		record.AddTimestamp = record.AddTime.Format("2006-01-02T15:04:05-0700")
		p.UpdateHistory[key] = record
	}

	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(p.Path, data, 0755)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(os.Getenv("HOME")+"/.todoistist/records", 0755)
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
