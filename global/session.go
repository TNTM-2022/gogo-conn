package global

import (
	"encoding/json"
	"fmt"
	concurrentMap "github.com/orcaman/concurrent-map"
	"go-connector/config"
	"sync"
)

var sessions = concurrentMap.New() // sid >> sessionType string:sessionType
var uidSid = concurrentMap.New()   // uid >> sid  修改 后端推送 问题

func SessionsCount() int {
	return sessions.Count()
}

func GetSessionBySid(sid uint32) (*sessionType, bool) {
	if v, ok := sessions.Get(fmt.Sprintf("%v", sid)); ok {
		if vv, ok := v.(*sessionType); ok {
			return vv, ok
		}
	}
	return nil, false
}
func GetSessionByUid(uid uint32) (*sessionType, bool) {
	_sid, ok := uidSid.Get(fmt.Sprintf("%v", uid))
	if !ok {
		return nil, false
	}
	sid, ok := _sid.(uint32)
	if !ok {
		return nil, false
	}
	if v, ok := sessions.Get(fmt.Sprintf("%v", sid)); ok {
		if vv, ok := v.(*sessionType); ok {
			return vv, ok
		}
	}
	return nil, false
}

func GetSidByUid(uid uint32) (uint32, bool) {
	if v, ok := uidSid.Get(fmt.Sprintf("%v", uid)); ok {
		if vv, ok := v.(uint32); ok {
			fmt.Println("1", v)
			return vv, true
		}
	}
	fmt.Println("2")
	return 0, false
}

type SessionSettings struct {
	//settings map[string]interface{}
	settings map[string]json.RawMessage
	locker   *sync.RWMutex
}

func (s SessionSettings) MarshalJSON() (buf []byte, err error) {
	if s.locker == nil {
		buf = []byte("{}")
		return
	}
	s.locker.RLock()
	buf, err = json.Marshal(s.settings)
	s.locker.RUnlock()
	return
}

type SessionInterface interface {
	Bind(uint32) bool
	Unbind()
	Destroy()
	Set(string, json.RawMessage)
	Unset(string)
}
type sessionType struct {
	Sid           uint32          `json:"id"`
	FrontServerId string          `json:"frontendId"`
	Uid           uint32          `json:"uid,omitempty"`
	Settings      SessionSettings `json:"settings,omitempty"`
}

func CreateSession(sid uint32) SessionInterface {
	s := &sessionType{
		Sid:           sid,
		FrontServerId: *config.ServerID,
		Settings: SessionSettings{
			locker: &sync.RWMutex{},
		},
	}
	sessions.Set(fmt.Sprintf("%v", sid), s)
	return s
}

func (s *sessionType) Bind(uid uint32) bool {
	s.Uid = uid
	if !sessions.Has(fmt.Sprintf("%v", s.Sid)) {
		fmt.Println("no found session")
		return false
	}
	uidSid.Set(fmt.Sprintf("%v", uid), s.Sid)
	fmt.Println(">>>>> ", fmt.Sprintf("%v", uid), fmt.Sprintf("%v", s.Sid))
	return true
}

func (s *sessionType) Unbind() {
	uidSid.Remove(fmt.Sprintf("%v", s.Uid))
	s.Uid = 0
}
func (s *sessionType) Destroy() {
	if s.Uid > 0 {
		uidSid.RemoveCb(fmt.Sprintf("%v", s.Uid), func(k string, v interface{}, exists bool) bool {
			if !exists {
				return false
			}
			if _uid, ok := v.(uint32); ok && _uid == s.Uid {
				return true
			}
			return false
		})
	}
	sessions.Remove(fmt.Sprintf("%v", s.Sid))
}

func (s *sessionType) Set(k string, v json.RawMessage) {
	s.Settings.locker.Lock()
	s.Settings.settings[k] = v
	s.Settings.locker.Unlock()
}
func (s *sessionType) Unset(k string) {
	s.Settings.locker.Lock()
	defer s.Settings.locker.Unlock()
	delete(s.Settings.settings, k)
}
