package authgoogle

import (
	"Liature-Server/serversession"
	"log"
	"net/http"

	sessions "github.com/goincremental/negroni-sessions"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
)

const (
	nextPageKey     = "next_page" // 세션에 저장되는 next page의 키
	authSecurityKey = "auth_security_key"
)

func init() {
	// gomniauth 정보 세팅
	gomniauth.SetSecurityKey(authSecurityKey)
	gomniauth.WithProviders(
		google.New("476773135653-1m6k8n8v6b7lu718nk5s3jp2bc79l72k.apps.googleusercontent.com", "1h7278fFvMdmCIDsEy8CtLWX", "http://127.0.0.1:3000/auth/callback/google"),
	)
}

// AuthGoogle 함수는 구글 소셜 로그인 작업을 수행합니다.
func AuthGoogle(w http.ResponseWriter, r *http.Request, action string, provider string) {
	s := sessions.GetSession(r)

	switch action {
	case "login":
		// gomniauth.Provider의 login 페이지로 이동
		p, err := gomniauth.Provider(provider)
		if err != nil {
			log.Fatalln(err)
		}
		loginURL, err := p.GetBeginAuthURL(nil, nil)
		if err != nil {
			log.Fatalln(err)
		}
		http.Redirect(w, r, loginURL, http.StatusFound)
	case "callback":
		// gomniauth 콜백 처리
		p, err := gomniauth.Provider(provider)
		if err != nil {
			log.Fatalln(err)
		}
		creds, err := p.CompleteAuth(objx.MustFromURLQuery(r.URL.RawQuery))
		if err != nil {
			log.Fatalln(err)
		}

		// 콜백 결과로부터 사용자 정보 확인
		user, err := p.GetUser(creds)
		if err != nil {
			log.Fatalln(err)
		}

		if err != nil {
			log.Fatalln(err)
		}

		u := &serversession.SessionUser{
			UID:       user.Data().Get("id").MustStr(),
			Name:      user.Name(),
			Email:     user.Email(),
			AvatarURL: user.AvatarURL(),
		}

		serversession.SetCurrentUser(r, u)
		http.Redirect(w, r, s.Get(nextPageKey).(string), http.StatusFound)
	default:
		http.Error(w, "Auth action '"+action+"' is not supported", http.StatusNotFound)
	}
}
