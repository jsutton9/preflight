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

func UserCache(readChannel chan *ReadRequest, writeChannel chan *WriteRequest,
		updateChannel chan *user.UserDelta) {
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
			}
			if ! write.Remove && (! write.OnlyIfCached || cached != nil) {
				byId[u.GetId()] = u
				byEmail[u.Email] = u
				for _, token := range u.Security.Tokens {
					byToken[token.Secret] = u
				}
			}
			go publishUpdate(cached, u, write.Remove, updateChannel)

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

func publishUpdate(oldUser *user.User, newUser *user.User, remove bool,
		updateChannel chan *user.UserDelta) {
	var delta *user.UserDelta
	if remove {
		delta = oldUser.GetDelta(nil)
	} else {
		delta = oldUser.GetDelta(newUser)
	}
	if delta != nil {
		updateChannel <- delta
	}
}
