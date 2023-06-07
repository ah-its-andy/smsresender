package tastek

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

type Session struct {
	jar *cookiejar.Jar
}

func NewSession(jar *cookiejar.Jar) *Session {
	return &Session{jar: jar}
}

func (s *Session) Do(req *http.Request) (*http.Response, error) {
	httpCli := &http.Client{
		Jar: s.jar,
	}
	return httpCli.Do(req)
}

func (s *Session) PostForm(url string, data url.Values) (*http.Response, error) {
	httpCli := &http.Client{
		Jar: s.jar,
	}
	return httpCli.PostForm(url, data)
}
