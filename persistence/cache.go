package persistence

type ReadRequest struct {
	Id string
	ResponseChannel chan *User
}

type WriteRequest struct {
	User *User
	Remove bool
}

func UserCache(readChannel chan *ReadRequest, writeChannel chan *WriteRequest) {
	users := make(map[string]*User)
	var read *ReadRequest
	var write *WriteRequest
	for {
		select {
		case write = <-writeChannel:
			if write.Remove {
				delete(users, write.User.GetId())
			} else {
				users[write.User.GetId()] = write.User
			}
		case read = <-readChannel:
			u, _ := users[read.Id]
			read.ResponseChannel <- u
		}
	}
}
