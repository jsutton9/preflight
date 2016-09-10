package persistence

type ReadRequest struct {
	Id string
	Email string
	TokenSecret string
	ResponseChannel chan *User
}

type WriteRequest struct {
	User *User
	Remove bool
	OnlyIfCached bool
}

func UserCache(readChannel chan *ReadRequest, writeChannel chan *WriteRequest) {
	byId := make(map[string]*User)
	byEmail := make(map[string]*User)
	byToken := make(map[string]*User)
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
		case read = <-readChannel:
			var u *User
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
