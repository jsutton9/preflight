package persistence

import (
	"github.com/jsutton9/preflight/user"
)

type ReadRequest struct {
	Id string
	Email string
	TokenSecret string
	ResponseChannel chan *user.User
}

type WriteRequest struct {
	User *user.User
	Remove bool
	OnlyIfCached bool
}

// TODO: update publish channel as param
func UserCache(readChannel chan *ReadRequest, writeChannel chan *WriteRequest) {
	byId := make(map[string]*user.User)
	byEmail := make(map[string]*user.User)
	byToken := make(map[string]*user.User)
	var read *ReadRequest
	var write *WriteRequest
	for {
		select {
		case write = <-writeChannel:
			u := write.User
			cached := byId[u.GetId()]
			if cached != nil {
				delete(byId, cached.GetId())
				delete(byEmail, cached.Email)
				for _, token := range cached.Security.Tokens {
					delete(byToken, token.Secret)
				}
				// TODO: if remove, publish user removal
				// TODO: if remove, publish checklist removal
			}
			if ! write.Remove && (! write.OnlyIfCached || cached != nil) {
				byId[u.GetId()] = u
				byEmail[u.Email] = u
				for _, token := range u.Security.Tokens {
					byToken[token.Secret] = u
				}
				// TODO: publish new user
				// TODO: publish updated checklists
			}
		case read = <-readChannel:
			var u *user.User
			if read.Id != "" {
				u, _ = byId[read.Id]
			} else if read.Email != "" {
				u, _ = byEmail[read.Email]
			} else if read.TokenSecret != "" {
				u, _ = byToken[read.TokenSecret]
			}
			read.ResponseChannel <- u
		}
	}
}
