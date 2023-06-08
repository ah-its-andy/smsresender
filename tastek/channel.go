package tastek

import (
	"fmt"
	"net/http/cookiejar"
	"time"

	"github.com/ah-its-andy/smsresender/dao"
	"github.com/ah-its-andy/smsresender/db"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type SmsChannel struct {
	channelName string
	host        string
	username    string
	password    string
	tock        time.Duration
	// c         chan *SmsTotalResult
	closeChan chan bool
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

// func (s *SmsChannel) Consume() <-chan *SmsTotalResult {
// 	return s.c
// }

func (s *SmsChannel) Close() {
	s.closeChan <- true
}

func (s *SmsChannel) poll() {
	s.closeChan = make(chan bool)
	timer := time.NewTimer(s.tock)
	for {
		breaks := false
		select {
		case <-timer.C:
			gdb, err := db.OpenConnection(db.DefaultOptions())
			if err != nil {
				log.Printf("[ERROR] db.OpenConnection failed: %v", err)
			}
			jar, _ := cookiejar.New(nil)
			session := NewSession(jar)
			if sms, err := s.getSms(session); err != nil {
				log.Printf("[ERROR] Failed to get SMS: %v", err)
			} else {
				err := gdb.Transaction(func(tx *gorm.DB) error {
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
					log.Printf("[ERROR] Failed to create SMS: %v", err)
					return
				}

				// msgContent := strings.Builder{}
				// if len(sms.Messages) == 1 {
				// 	msgContent.WriteString(fmt.Sprintf("发件人: %s", sms.Messages[0].Number))
				// 	msgContent.WriteString("\r\n")
				// 	if len(sms.Messages[0].Content) == 0 {
				// 		msgContent.WriteString("正文: 没有内容")
				// 		msgContent.WriteString("\r\n")
				// 	} else {
				// 		msgContent.WriteString(fmt.Sprintf("正文: %s", sms.Messages[0].Content))
				// 		msgContent.WriteString("\r\n")
				// 	}
				// } else {
				// 	for _, msg := range sms.Messages {
				// 		msgContent.WriteString(fmt.Sprintf("发件人: %s", msg.Number))
				// 		msgContent.WriteString("\r\n")
				// 		if len(msg.Content) == 0 {
				// 			msgContent.WriteString("正文: 没有内容")
				// 			msgContent.WriteString("\r\n")
				// 		} else {
				// 			msgContent.WriteString(fmt.Sprintf("正文: %s", msg.Content))
				// 			msgContent.WriteString("\r\n")
				// 		}
				// 		msgContent.WriteString("========")
				// 		msgContent.WriteString("\r\n")
				// 	}
				// }

				// if err := SendMessage(msgSubject, msgContent.String()); err != nil {
				// 	log.Printf("[ERROR] SendMessage failed: %v", err)
				// } else {
				// 	// for _, msg := range sms.Messages {
				// 	// 	DelSms(session, s.host, msg.ID)
				// 	// }
				// }
			}
		case <-s.closeChan:
			log.Printf("[DEBUG] Tastek channel closing")
			breaks = true
		}
		if breaks {
			break
		}
	}
	log.Printf("[DEBUG] Tastek channel closed")
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
