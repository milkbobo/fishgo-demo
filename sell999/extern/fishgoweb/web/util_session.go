package web

import (
	"github.com/astaxie/beego/session"
	_ "github.com/astaxie/beego/session/redis"
	_ "github.com/milkbobo/fishgoweb/web/util_session"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type SessionStore interface {
	Set(key, value interface{}) error
	Get(key interface{}) interface{}
	Delete(key interface{}) error
	SessionID() string
	SessionRelease()
	Flush() error
}

type Session interface {
	WithContext(ctx Context) Session
	SessionStart() (session SessionStore, err error)
}

type SessionConfig struct {
	Driver          string `toml:"driver"`
	CookieName      string `toml:"cookieName"`
	EnableSetCookie bool   `toml:"enableSetCookie"`
	GcLifeTime      int    `toml:"gclifetime"`
	Secure          bool   `toml:"secure"`
	CookieLifeTime  int    `toml:"cookieLifeTime"`
	ProviderConfig  string `toml:"providerConfig"`
	Domain          string `toml:"domain"`
	SessionIdLength int    `toml:"sessionIdLength"`
}

type sessionImplement struct {
	*session.Manager
	config SessionConfig
	ctx    Context
}

type sessionStoreImplement struct {
	session.Store
	responseWriter http.ResponseWriter
}

func NewSession(config SessionConfig) (Session, error) {
	if config.Driver == "" {
		return nil, nil
	}
	if config.CookieName == "" {
		config.CookieName = "beego_session"
	}
	if config.CookieLifeTime == 0 {
		config.CookieLifeTime = 3600
	}
	if config.GcLifeTime == 0 {
		config.GcLifeTime = 3600
	}

	cf := new(session.ManagerConfig)
	cf.Domain = config.Domain
	cf.SessionIDLength = int64(config.SessionIdLength)
	cf.CookieLifeTime = config.CookieLifeTime
	cf.ProviderConfig = config.ProviderConfig
	cf.Secure = config.Secure
	cf.Gclifetime = int64(config.GcLifeTime)
	cf.EnableSetCookie = config.EnableSetCookie
	cf.CookieName = config.CookieName
	cf.Gclifetime = int64(config.GcLifeTime)

	sessionManager, err := session.NewManager(config.Driver, cf)
	if err != nil {
		return nil, err
	}
	go sessionManager.GC()

	return &sessionImplement{
		Manager: sessionManager,
		config:  config,
	}, nil
}

func NewSessionFromConfig(configName string) (Session, error) {

	sessionlink := SessionConfig{}
	sessionlink.Driver = globalBasic.Config.Get().Session.Driver
	sessionlink.CookieName = globalBasic.Config.Get().Session.CookieName
	sessionlink.EnableSetCookie = globalBasic.Config.Get().Session.EnableSetCookie
	sessionlink.GcLifeTime = globalBasic.Config.Get().Session.GcLifeTime
	sessionlink.Secure = globalBasic.Config.Get().Session.Secure
	sessionlink.CookieLifeTime = globalBasic.Config.Get().Session.CookieLifeTime
	sessionlink.ProviderConfig = globalBasic.Config.Get().Session.ProviderConfig
	sessionlink.Domain = globalBasic.Config.Get().Session.Domain
	sessionlink.SessionIdLength = globalBasic.Config.Get().Session.SessionIdLength

	return NewSession(sessionlink)
}

func newSessionStoreImplement(store session.Store, responseWriter http.ResponseWriter) SessionStore {
	result := sessionStoreImplement{
		Store:          store,
		responseWriter: responseWriter,
	}
	return &result
}

func (manager *sessionImplement) WithContext(ctx Context) Session {
	result := *manager
	result.ctx = ctx
	return &result
}

func (manager *sessionImplement) SessionStart() (session SessionStore, err error) {
	w := manager.ctx.GetRawResponseWriter().(http.ResponseWriter)
	r := manager.ctx.GetRawRequest().(*http.Request)

	result, errOrgin := manager.Manager.SessionStart(w, r)
	if errOrgin != nil {
		return newSessionStoreImplement(result, w), errOrgin
	}
	//获取当前的cookie值
	cookie, err := r.Cookie(manager.config.CookieName)
	if err != nil || cookie.Value == "" {
		return newSessionStoreImplement(result, w), errOrgin
	}
	sid, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		return newSessionStoreImplement(result, w), errOrgin
	}

	//补充延续session时间的逻辑
	cookieValue := w.Header().Get("Set-Cookie")
	cookieName := manager.config.CookieName
	if strings.Index(cookieValue, cookieName) != -1 {
		return newSessionStoreImplement(result, w), err
	}
	cookie = &http.Cookie{
		Name:     manager.config.CookieName,
		Value:    url.QueryEscape(sid),
		Path:     "/",
		HttpOnly: true,
		Secure:   manager.config.Secure,
		Domain:   manager.config.Domain,
	}
	if manager.config.CookieLifeTime > 0 {
		cookie.MaxAge = manager.config.CookieLifeTime
		cookie.Expires = time.Now().Add(time.Duration(manager.config.CookieLifeTime) * time.Second)
	}
	if manager.config.EnableSetCookie {
		http.SetCookie(w, cookie)
	}
	return newSessionStoreImplement(result, w), errOrgin
}

func (this *sessionStoreImplement) SessionRelease() {
	this.Store.SessionRelease(this.responseWriter)
}
