package serversession

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	sessions "github.com/goincremental/negroni-sessions"
	"github.com/urfave/negroni"
)

const (
	currentUserKey  = "oauth2_current_user" // 세션에 저장되는 CurrentUser의 키
	sessionDuration = time.Hour             // 로그인 세션 유지 시간
	// 애플리케이션에서 사용할 세션의 키 정보
	sessionKey    = "simple_chat_session"
	sessionSecret = "simple_chat_session_secret"

	nextPageKey     = "next_page" // 세션에 저장되는 next page의 키
	authSecurityKey = "auth_security_key"
)

// GetCurrentUserKey 는 현재 유저 키를 반환하는 함수입니다.
func GetCurrentUserKey() string {
	return currentUserKey
}

// GetSessionDuration 는 로그인 세션 유지 시간을 반환하는 함수입니다.
func GetSessionDuration() string {
	return currentUserKey
}

// SessionUser 는 세션에 저장할 유저 정보를 담고 있습니다.
type SessionUser struct {
	UID       string    `json:"uid"`
	Name      string    `json:"name"`
	Email     string    `json:"user"`
	AvatarURL string    `json:"avatar_url"`
	Expired   time.Time `json:"expired"`
}

// Valid 메서드는 현재 시간 기준으로 만료 시간을 확인합니다
func (u *SessionUser) Valid() bool {
	// 현재 시간 기준으로 만료 시간 확인
	return u.Expired.Sub(time.Now()) > 0
}

// Refresh 메서드는 만료 시간을 연장합니다.
func (u *SessionUser) Refresh() {
	// 만료 시간 시간 연장
	u.Expired = time.Now().Add(sessionDuration)
}

// GetCurrentUser 메서드는 세션에서 현재 유저 정보를 가져옵니다.
func GetCurrentUser(r *http.Request) *SessionUser {
	// 세션에서 CurrentUser 정보를 가져옴
	s := sessions.GetSession(r)

	if s.Get(currentUserKey) == nil {
		return nil
	}

	data := s.Get(currentUserKey).([]byte)
	var u SessionUser
	json.Unmarshal(data, &u)
	return &u
}

// SetCurrentUser 함수는 세션에 현재 유저를 세팅합니다.
func SetCurrentUser(r *http.Request, u *SessionUser) {
	if u != nil {
		// CurrentUser 만료 시간 갱신
		u.Refresh()
	}

	// 세션에 CurrentUser 정보를 json으로 저장
	s := sessions.GetSession(r)
	val, _ := json.Marshal(u)
	s.Set(currentUserKey, val)
}

// LoginRequired sms ...
func LoginRequired(ignore ...string) negroni.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		// ignore url이면 다음 핸들러 실행
		for _, s := range ignore {
			if strings.HasPrefix(r.URL.Path, s) {
				next(w, r)
				return
			}
		}
		// CurrentUser 정보를 가져옴
		u := GetCurrentUser(r)

		// CurrentUser 정보가 유효하면 만료 시간을 갱신하고 다음 핸들러 실행
		if u != nil && u.Valid() {
			SetCurrentUser(r, u)
			next(w, r)
			return
		}

		// CurrentUser 정보가 유효하지 않으면 CurrentUser를 nil로 세팅
		SetCurrentUser(r, nil)

		// 로그인 후 이동할 url을 세션에 저장(r)
		sessions.GetSession(r).Set(nextPageKey, r.URL.RequestURI())

		// 로그인 페이지로 리다이렉트
		http.Redirect(w, r, "/login", http.StatusFound)
	}
}
