package session

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/solgar/upendo/database"
	"github.com/solgar/upendo/security"
	"github.com/solgar/upendo/settings"
)

const (
	cmdGetSession    = 1
	cmdCreateSession = 2
	cmdRemoveExpired = 3
	cmdExpireSession = 4
	cmdExpireUserID  = 5
)

var (
	instance         *Manager
	roleIntValueMap  = map[string]int{"anon": 1, "user": 2, "admin": 3, "root": 4}
	sessionsFilePath = settings.StartDir + "sessions.json"
	initDone         = false
)

type response struct {
	session *Session
	err     error
}

type command struct {
	respChan  chan<- response
	code      int
	sessionID string
	r         *http.Request
	user      string
	userID    int
}

// Session object contains information about user such as its name or role. It's
// used by SessionManager
type Session struct {
	ID         string
	UserID     int
	UserName   string
	UserRole   string
	Agent      string
	RemoteAddr string
	LastAccess int64
}

// Manager is the object to interact with sessions data. It allows to create, retrieve and delete sessions.
type Manager struct {
	activeSessionsMap map[string]*Session
	commandsChan      chan *command
	db                *sql.DB
}

///////////////////////////////////////////////////////////////// functions
func Initialize() {
	if initDone {
		panic("Initialization already done.")
	}
	initDone = true

	security.Initialize()

	instance = &Manager{}
	instance.initialize()
	go instance.commandProcessor()
	go instance.periodicExpiredSessionsClean()
	fmt.Println("Session manager started.")

	if settings.RestoreSessions {
		fileData, err := ioutil.ReadFile(sessionsFilePath)
		if err != nil {
			fmt.Println("Cannot restore sessions:", err)
			return
		}
		err = json.Unmarshal(fileData, &(instance.activeSessionsMap))
		if err != nil {
			fmt.Println("Cannot unmarshal sessions:", err)
		}
	}
}

// GetManager returns pointer to Manager singleton.
func GetManager() *Manager {
	return instance
}

// Deinit shuts down session management and if proper option was set it stores session data into a json file.
func Deinit() {
	b, err := json.Marshal(instance.activeSessionsMap)
	if err != nil {
		fmt.Println("Cannot marshal sessions:", err)
	}
	_ = os.Remove(sessionsFilePath)
	err = ioutil.WriteFile(sessionsFilePath, b, 0644)
	if err != nil {
		fmt.Println("Cannot archive sessions:", err)
	}
}

// RoleOrHigher TODO
func RoleOrHigher(userRole, role string) bool {
	ur := roleIntValueMap[userRole]
	r := roleIntValueMap[role]
	return ur >= r
}

// RoleOrLower TODO
func RoleOrLower(userRole, role string) bool {
	ur := roleIntValueMap[userRole]
	r := roleIntValueMap[role]
	return ur <= r
}

///////////////////////////////////////////////////////////////// SessionManager methods
func (s *Manager) initialize() {
	s.activeSessionsMap = make(map[string]*Session)
	s.commandsChan = make(chan *command)
	var err error
	s.db, err = database.ConnectToDb("user", "pass", "dbname")
	if err != nil {
		panic(err)
	}
}

func generateSessionID(user string, r *http.Request) string {
	addr := strings.Split(r.RemoteAddr, ":")
	remoteIP := ""

	if len(addr) > 0 {
		remoteIP = addr[0]
	}

	h := sha256.New()
	io.WriteString(h, user)
	io.WriteString(h, r.UserAgent())
	io.WriteString(h, remoteIP)
	io.WriteString(h, security.GenerateRandomSalt())
	sID := h.Sum(nil)

	return hex.EncodeToString(sID)
}

func (s *Manager) periodicExpiredSessionsClean() {
	for {
		time.Sleep(time.Minute * 10)
		rchan := make(chan response)
		s.commandsChan <- &command{respChan: rchan, code: cmdRemoveExpired}
		<-rchan
	}
}

func (s *Manager) createSession(user string, r *http.Request) *Session {
	sessionObject := &Session{LastAccess: time.Now().Unix()}
	sessionID := generateSessionID(user, r)

	sessionObject.ID = sessionID
	sessionObject.UserName = user
	sessionObject.RemoteAddr = strings.Split(r.RemoteAddr, ":")[0]
	sessionObject.Agent = r.UserAgent()

	err := s.db.QueryRow("SELECT id FROM User WHERE login=?", sessionObject.UserName).Scan(&sessionObject.UserID)
	if err != nil {
		panic(err)
	}

	s.db.QueryRow("SELECT Role.name FROM Role INNER JOIN UserRole ON UserRole.role_id=Role.id WHERE UserRole.user_id=?", sessionObject.UserID).Scan(&sessionObject.UserRole)
	if err != nil {
		panic(err)
	}

	return sessionObject
}

func (s *Manager) commandProcessor() {
	for cmd := range s.commandsChan {
		switch cmd.code {

		case cmdGetSession:
			sessionObject, ok := s.activeSessionsMap[cmd.sessionID]
			ok = ok && sessionObject.RemoteAddr == strings.Split(cmd.r.RemoteAddr, ":")[0]
			if ok {
				cmd.respChan <- response{sessionObject, nil}
			} else {
				cmd.respChan <- response{nil, errors.New("No session for id: " + cmd.sessionID)}
			}
			break

		case cmdCreateSession:
			sessionObject := s.createSession(cmd.user, cmd.r)
			s.activeSessionsMap[sessionObject.ID] = sessionObject
			cmd.respChan <- response{sessionObject, nil}
			break

		case cmdRemoveExpired:
			var toRemove []string
			now := time.Now().Unix()

			for k, v := range s.activeSessionsMap {
				if now-v.LastAccess > int64(time.Hour) {
					toRemove = append(toRemove, k)
				}
			}

			for _, v := range toRemove {
				delete(s.activeSessionsMap, v)
			}
			cmd.respChan <- response{}
			break

		case cmdExpireSession:
			delete(s.activeSessionsMap, cmd.sessionID)
			break

		case cmdExpireUserID:
			var k string
			var v *Session
			found := false
			for k, v = range s.activeSessionsMap {
				if v.UserID == cmd.userID {
					found = true
					break
				}
			}
			if found {
				delete(s.activeSessionsMap, k)
			}
			break
		}
	}
	fmt.Println("commandProcessor finished!")
}

// ExpireSessionByUserID TODO
func (s *Manager) ExpireSessionByUserID(userID int) {
	s.commandsChan <- &command{make(chan response), cmdExpireUserID, strconv.Itoa(userID), nil, "", userID}
	fmt.Println("session for user with id:", userID, "- expired")
}

// ExpireSession TODO
func (s *Manager) ExpireSession(params map[string]interface{}) {
	w := params["__writer"].(http.ResponseWriter)
	currentSession := params["session"].(*Session)
	if currentSession != nil {
		s.commandsChan <- &command{make(chan response), cmdExpireSession, currentSession.ID, nil, "", 0}
		fmt.Println("session:", currentSession.ID, " expired")
	}
	expiredCookie := &http.Cookie{Name: "data", Value: "", Expires: time.Unix(0, 0)}
	http.SetCookie(w, expiredCookie)
}

// GetSession TODO
func (s *Manager) GetSession(request *http.Request) *Session {
	c, err := request.Cookie("data")

	if err != nil || c == nil || c.Value == "" {
		return nil
	}

	sessionID := c.Value
	respChan := make(chan response)
	s.commandsChan <- &command{respChan, cmdGetSession, sessionID, request, "", 0}
	resp := <-respChan
	if resp.session != nil {
		fmt.Println("session:", sessionID, " exists")
	} else {
		fmt.Println("session:", sessionID, " does not exists")
	}
	return resp.session
}

// CreateSession TODO
func (s *Manager) CreateSession(params map[string]interface{}) *Session {
	w := params["__writer"].(http.ResponseWriter)
	r := params["request"].(*http.Request)

	respChan := make(chan response)
	s.commandsChan <- &command{respChan: respChan, code: cmdCreateSession, r: r, user: params["login"].(string)}
	resp := <-respChan
	fmt.Println("session:", resp.session.ID, " created")

	sessionCookie := &http.Cookie{Name: "data", Value: resp.session.ID}
	http.SetCookie(w, sessionCookie)

	return resp.session
}
