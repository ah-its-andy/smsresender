package tastek_test

import (
	"net/http/cookiejar"
	"testing"

	"github.com/ah-its-andy/smsresender/tastek"
)

func TestGetSms(t *testing.T) {
	jar, _ := cookiejar.New(nil)
	session := tastek.NewSession(jar)
	if err := tastek.SignIn(session, "192.168.3.1", "admin", "12345678"); err != nil {
		t.Error(err)
	}
	if sms, err := tastek.SmsTotal(session, "192.168.3.1"); err != nil {
		t.Error(err)
	} else {
		t.Log(sms)
	}
}
