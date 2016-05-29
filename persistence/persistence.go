package persistence

import (
	"github.com/jsutton9/preflight/config"
	"github.com/jsutton9/preflight/security"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type User struct {
	Id bson.ObjectId               `json:"id" bson:"_id,omitempty"`
	Settings GeneralSettings       `json:"generalSettings"`
	Security security.SecurityInfo `json:"securityInfo"`
	Checklists []config.Checklist  `json:"checklists"`
}

type GeneralSettings struct {
	Timezone string `json:"timezone"`
}
