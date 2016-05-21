package persistence

import (
	"encoding/json"
	"errors"
	"github.com/jsutton9/preflight/config"
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
	dir := os.Getenv("HOME")+"/.preflight/records/"
	p := Persister{Path:dir+user+".json"}

	data, err := ioutil.ReadFile(p.Path)
	if err != nil {
		if os.IsNotExist(err) {
			p.UpdateHistory = make(map[string]UpdateRecord)
			return p, nil
		} else {
			return p, errors.New("persistence.Load: error reading records: " +
				"\n\t" + err.Error())
		}
	}
	err = json.Unmarshal(data, &p)
	if err != nil {
		return p, errors.New("persistence.Load: error parsing records: " +
			"\n\t" + err.Error())
	}

	for name, record := range p.UpdateHistory {
		if record.Timestamp != "" {
			record.Time, err = time.Parse("2006-01-02T15:04:05-0700", record.Timestamp)
			p.UpdateHistory[name] = record
			if err != nil {
				return p, errors.New("persistence.Load: error parsing timestamp " +
					"\"" + record.Timestamp + "\": " + "\n\t" + err.Error())
			}
		}
		if record.AddTimestamp != "" {
			record.AddTime, err = time.Parse("2006-01-02T15:04:05-0700", record.AddTimestamp)
			p.UpdateHistory[name] = record
			if err != nil {
				return p, errors.New("persistence.Load: error parsing add timestamp " +
					"\"" + record.AddTimestamp + "\": " + "\n\t" + err.Error())
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
		return errors.New("persistence.Persister.Save: error marshalling data: " +
			"\n\t" + err.Error())
	}
	err = ioutil.WriteFile(p.Path, data, 0755)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(os.Getenv("HOME")+"/.preflight/records", 0755)
			if err != nil {
				return errors.New("persistence.Persister.Save: error making records dir: " +
					"\n\t" + err.Error())
			}
			err = ioutil.WriteFile(p.Path, data, 0755)
			if err != nil {
				return errors.New("persistence.Persister.Save: error writing records file: " +
					"\n\t" + err.Error())
			}
		} else {
			return errors.New("persistence.Persister.Save: error writing records file: " +
				"\n\t" + err.Error())
		}
	}

	return nil
}
