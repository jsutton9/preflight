package persistence

import (
	"encoding/json"
	"github.com/jsutton9/preflight/api/errors"
	"github.com/jsutton9/preflight/security"
	"github.com/jsutton9/preflight/user"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"io/ioutil"
	"os"
)

type Persister struct {
	Session *mgo.Session
	databaseName string
	UserCollection *mgo.Collection
	NodeCollection *mgo.Collection
	UserCacheReadChannel chan *ReadRequest
	UserCacheWriteChannel chan *WriteRequest
}

type ServerSettings struct {
	Port int                       `json:"port"`
	CertFile string                `json:"certFile"`
	KeyFile string                 `json:"keyFile"`
	ErrLog string                  `json:"errLog"`
	DatabaseServer string          `json:"databaseServer"`
	DatabaseUsersCollection string `json:"databaseUsersCollection"`
	TrelloAppKey string            `json:"trelloAppKey"`
	SecretFile string              `json:"secretFile"`
}

type Node struct {
	Secret string
}

type LoggerCloser struct {
	*log.Logger
	file *os.File
	isStderr bool
}

func New(url, database string) (*Persister, *errors.PreflightError) {
	session, err := mgo.Dial(url)
	if err != nil {
		return nil, &errors.PreflightError{
			Status: 500,
			InternalMessage: "persistence.New: " +
				"error dialing \""+url+"\": \n\t" + err.Error(),
			ExternalMessage: "There was an error connecting to the database.",
		}
	}

	userCollection := session.DB(database).C("users")
	nodeCollection := session.DB(database).C("nodes")

	readChannel := make(chan *ReadRequest)
	writeChannel := make(chan *WriteRequest)
	go UserCache(readChannel, writeChannel)

	p := Persister{
		Session: session,
		databaseName: database,
		UserCollection: userCollection,
		NodeCollection: nodeCollection,
		UserCacheReadChannel: readChannel,
		UserCacheWriteChannel: writeChannel,
	}

	return &p, nil
}

func (p Persister) Close() {
	p.Session.Close()
}

func (p Persister) Copy() *Persister {
	session := p.Session.Copy()
	userCollection := session.DB(p.databaseName).C("users")
	nodeCollection := session.DB(p.databaseName).C("nodes")
	return &Persister{
		Session: session,
		databaseName: p.databaseName,
		UserCollection: userCollection,
		NodeCollection: nodeCollection,
		UserCacheReadChannel: p.UserCacheReadChannel,
		UserCacheWriteChannel: p.UserCacheWriteChannel,
	}
}

func (p Persister) RegisterNode(secretFile string) *errors.PreflightError {
	secret, pErr := security.GenerateNodeSecret()
	if pErr != nil {
		return pErr.Prepend("persistence.Persister.RegisterNode: error generating secret: ")
	}

	err := ioutil.WriteFile(secretFile, []byte(secret), 0600)
	if os.IsNotExist(err) {
		_, pErr := createFile(secretFile)
		if pErr != nil {
			pErr.Prepend("persistence.Persister.RegisterNode: error making file: ")
			return pErr
		}
		err = ioutil.WriteFile(secretFile, []byte(secret), 0600)
		if err != nil {
			return &errors.PreflightError{
				Status: 500,
				InternalMessage: "persistence.Persister.RegisterNode: " +
					"error writing secret file: \n\t" + err.Error(),
				ExternalMessage: "There was an error registering the node.",
			}
		}
	} else if err != nil {
		return &errors.PreflightError{
			Status: 500,
			InternalMessage: "persistence.Persister.RegisterNode: " +
				"error writing secret file: \n\t" + err.Error(),
			ExternalMessage: "There was an error registering the node.",
		}
	}

	node := Node{Secret:secret}
	err = p.NodeCollection.Insert(&node)
	if err != nil {
		return &errors.PreflightError{
			Status: 500,
			InternalMessage: "persistence.Persister.RegisterNode: " +
				"error adding node to db: \n\t" + err.Error(),
			ExternalMessage: "There was an error adding the node to the database.",
		}
	}

	return nil
}

func (p Persister) GetNodeSecret(filename string) (string, *errors.PreflightError) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", &errors.PreflightError{
			Status: 500,
			InternalMessage: "persistence.GetNodeSecret: error reading secret file: " +
				"\n\t" + err.Error(),
			ExternalMessage: "There was a server authentication error.",
		}
	}

	return string(data[:]), nil
}

func (p Persister) ValidateNodeSecret(secret string) (bool, *errors.PreflightError) {
	n, err := p.NodeCollection.Find(bson.M{"secret": secret}).Count()
	if err != nil {
		return false, &errors.PreflightError{
			Status: 500,
			InternalMessage: "persistence.Persister.ValidateNodeSecret: error querying db: " +
				"\n\t" + err.Error(),
			ExternalMessage: "There was an error querying the database.",
		}
	}

	return n>0, nil
}

func (p Persister) AddUser(email, password string) (*user.User, *errors.PreflightError) {
	existing_user, _ := p.GetUserByEmail(email)
	if existing_user != nil {
		return nil, &errors.PreflightError{
			Status: 409,
			InternalMessage: "persistence.Persister.AddUser: " +
				"\n\tuser with email " + email + " already exists",
			ExternalMessage: "There is already a user with email " + email,
		}
	}

	newUser, pErr := user.New(email, password)
	if pErr != nil {
		return nil, pErr.Prepend("persistence.Persister.AddUser: ")
	}
	err := p.UserCollection.Insert(newUser)
	if err != nil {
		return nil, &errors.PreflightError{
			Status: 500,
			InternalMessage: "persistence.Persister.AddUser: " +
				"error inserting user:\n\t" + err.Error(),
			ExternalMessage: "There was an error adding the user to the database.",
		}
	}

	return newUser, nil
}

func (p Persister) UpdateUser(newUser *user.User) *errors.PreflightError {
	p.UserCacheWriteChannel <- &WriteRequest{User: newUser, OnlyIfCached: true}

	err := p.UserCollection.Update(bson.M{"_id": newUser.Id}, newUser)
	if err != nil {
		return &errors.PreflightError{
			Status: 500,
			InternalMessage: "persistence.Persister.UpdateUser: " +
				"error updating user:\n\t" + err.Error(),
			ExternalMessage: "There was an error updating the user in the database.",
		}
	}

	return nil
}

func (p Persister) DeleteUser(targetUser *user.User) *errors.PreflightError {
	p.UserCacheWriteChannel <- &WriteRequest{User: targetUser, Remove: true}

	err := p.UserCollection.Remove(bson.M{"_id": targetUser.Id})
	if err != nil {
		return &errors.PreflightError{
			Status: 500,
			InternalMessage: "persistence.Persister.DeleteUser: " +
				"error removing user:\n\t" + err.Error(),
			ExternalMessage: "There was an error removing the user from the database.",
		}
	}

	return nil
}

func (p Persister) GetUser(id string) (*user.User, *errors.PreflightError) {
	responseChannel := make(chan *user.User)
	p.UserCacheReadChannel <- &ReadRequest{Id: id, ResponseChannel: responseChannel}
	matchingUser := <-responseChannel
	if matchingUser != nil {
		return matchingUser, nil
	}

	matchingUser = &user.User{}
	err := p.UserCollection.FindId(bson.ObjectIdHex(id)).One(matchingUser)
	if err != nil {
		return nil, &errors.PreflightError{
			Status: 404,
			InternalMessage: "persistence.Persister.GetUser: " +
				"error finding user id=" + id + ": \n\t" + err.Error(),
			ExternalMessage: "User with id " + id + " not found",
		}
	}

	return matchingUser, nil
}

func (p Persister) GetUserByEmail(email string) (*user.User, *errors.PreflightError) {
	responseChannel := make(chan *user.User)
	p.UserCacheReadChannel <- &ReadRequest{Email: email, ResponseChannel: responseChannel}
	matchingUser := <-responseChannel
	if matchingUser != nil {
		return matchingUser, nil
	}

	matchingUser = &user.User{}
	err := p.UserCollection.Find(bson.M{"email": email}).One(matchingUser)
	if err != nil {
		return nil, &errors.PreflightError{
			Status: 404,
			InternalMessage: "persistence.Persister.GetUserByEmail: " +
				"error finding user email=" + email + ": \n\t" + err.Error(),
			ExternalMessage: "User with email " + email + " not found",
		}
	}

	return matchingUser, nil
}

func (p Persister) GetUserByToken(secret string) (*user.User, *errors.PreflightError) {
	responseChannel := make(chan *user.User)
	p.UserCacheReadChannel <- &ReadRequest{TokenSecret: secret, ResponseChannel: responseChannel}
	matchingUser := <-responseChannel
	if matchingUser != nil {
		return matchingUser, nil
	}

	matchingUser = &user.User{}
	err := p.UserCollection.Find(bson.M{"security.tokens.secret": secret}).One(matchingUser)
	if err != nil {
		return nil, &errors.PreflightError{
			Status: 401,
			InternalMessage: "persistence.Persister.GetUserByToken: " +
				"error finding user by token: \n\t" + err.Error(),
			ExternalMessage: "No user with that token was found",
		}
	}

	return matchingUser, nil
}

func GetServerSettings(filename string) (*ServerSettings, *errors.PreflightError) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, &errors.PreflightError{
			Status: 500,
			InternalMessage: "persistence.GetServerSettings: error reading file: " +
				"\n\t" + err.Error(),
			ExternalMessage: "There was an error.",
		}
	}

	settings := &ServerSettings{
		Port: 443,
		ErrLog: "",
		DatabaseServer: "localhost",
		DatabaseUsersCollection: "users",
		SecretFile: "/etc/preflight/secret",
	}
	err = json.Unmarshal(contents, settings)
	if err != nil {
		return nil, &errors.PreflightError{
			Status: 500,
			InternalMessage: "persistence.GetServerSettings: error parsing json: " +
				"\n\t" + err.Error(),
			ExternalMessage: "There was an error.",
		}
	}

	return settings, nil
}

func (s ServerSettings) GetLogger() (*LoggerCloser, *errors.PreflightError) {
	logger := new(LoggerCloser)
	logger.file = os.Stderr
	logger.isStderr = true

	if s.ErrLog != "" {
		logger.isStderr = false
		var err error
		logger.file, err = os.OpenFile(s.ErrLog, 1, 0660)
		if os.IsNotExist(err) {
			var pErr *errors.PreflightError
			logger.file, pErr = createFile(s.ErrLog)
			if pErr != nil {
				pErr.Prepend("persistence.ServerSettings.GetLogger: " +
					"error creating file: ")
				return nil, pErr
			}
		} else if err != nil {
			return nil, &errors.PreflightError{
				Status: 500,
				InternalMessage: "persistence.ServerSettings.GetLogger: " +
					"error opening log file \"" + s.ErrLog +
					"\": \n\t" + err.Error(),
				ExternalMessage: "There was an error.",
			}
		}
	}

	logger.Logger = log.New(logger.file, "", log.Ldate | log.Ltime)
	return logger, nil
}

func (s ServerSettings) GetPersister() (*Persister, *errors.PreflightError) {
	persister, err := New(s.DatabaseServer, s.DatabaseUsersCollection)
	if err != nil {
		err.Prepend("persistence.ServerSettings.GetPersister: error making Persister: ")
	}
	return persister, err
}

func (l LoggerCloser) Close() error {
	if ! l.isStderr {
		return l.file.Close()
	}
	return nil
}

func createFile(path string) (*os.File, *errors.PreflightError) {
	nameStart := len(path)
	for ; nameStart>0; nameStart-- {
		if path[nameStart-1] == "/"[0] {
			break
		}
	}
	dir := path[:nameStart]
	if len(dir) > 0 {
		err := os.MkdirAll(dir, os.ModeDir | 0774)
		if err != nil {
			return nil, &errors.PreflightError{
				Status: 500,
				InternalMessage: "persistence.createFile: " +
					"error making directory: \n\t" + err.Error(),
				ExternalMessage: "There was an error.",
			}
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, &errors.PreflightError{
			Status: 500,
			InternalMessage: "persistence.createFile: error creating file \"" +
				path + "\": \n\t" + err.Error(),
			ExternalMessage: "There was an error.",
		}
	}

	return f, nil
}
