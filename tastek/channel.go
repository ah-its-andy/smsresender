package tastek

import (
	"fmt"
	"net/http/cookiejar"
	"time"

	"github.com/ah-its-andy/smsresender/dao"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type SmsChannel struct {
	channelName string
	host        string
	username    string
	password    string
	tock        time.Duration
}

func NewSmsChannel(channelName, host, username, password string, tock time.Duration) *SmsChannel {
	return &SmsChannel{
		channelName: channelName,
		host:        host,
		username:    username,
		password:    password,
		tock:        tock,
	}
}

func (s *SmsChannel) Start() {
	go s.poll()
	log.Printf("[DEBUG] channel %s started, addr: %s, username: %s, password: %s, tok: %s", s.channelName, s.host, s.username, s.password, s.tock.String())
}

func (s *SmsChannel) pullSms() error {
	channelDbConn.Retrive()
	defer channelDbConn.Release()
	jar, _ := cookiejar.New(nil)
	session := NewSession(jar)
	if sms, err := s.getSms(session); err != nil {
		return fmt.Errorf("[ERROR] Failed to get SMS: %v", err)
	} else {
		err := channelDbConn.Conn.Transaction(func(tx *gorm.DB) error {
			for _, message := range sms.Messages {
				model := &dao.SmsModel{
					Device:    s.channelName,
					MessageId: message.ID,
					Sender:    message.Number,
					Content:   message.Content,
					RecTime:   message.Date,
					State:     10,
				}
				if err := dao.CreateSms(tx, model); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("[ERROR] Failed to create SMS: %v", err)
		}
	}
	return nil
}

func (s *SmsChannel) poll() {
	for {
		if err := s.pullSms(); err != nil {
			log.Printf("poll error: %v", err)
		}
		time.Sleep(time.Second * 5)
	}
}

func (s *SmsChannel) getSms(session *Session) (*SmsTotalResult, error) {
	if err := SignIn(session, s.host, s.username, s.password); err != nil {
		return nil, fmt.Errorf("sign in failed: %w", err)
	}
	if sms, err := SmsTotal(session, s.host); err != nil {
		return nil, fmt.Errorf("failed to get sms: %w", err)
	} else {
		return sms, nil
	}
}
