package telebot

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type User struct {
	hub    *Hub
	chatId int64

	msgChan   chan *tgbotapi.MessageConfig
	locker    *ReentrantLock
	listening bool
}

func NewUser(hub *Hub, chatId int64) *User {
	return &User{
		hub:    hub,
		chatId: chatId,
		locker: NewReentrantLock(),
	}
}

func (user *User) Reply(msg *tgbotapi.MessageConfig) {
	if user.msgChan == nil || !user.listening {
		log.Printf("[ERROR][telebot] user %d is not listening, skip reply, msgText: %s \r", msg.Text)
		return
	}
	user.msgChan <- msg
}

func (user *User) Listen() {
	if user.locker.running == 0 {
		user.locker.RunGC()
	}
	user.msgChan = make(chan *tgbotapi.MessageConfig)
	user.listening = true
	go func() {
		for {
			msg := <-user.msgChan
			encoder := sha256.New()
			encoder.Write([]byte(msg.Text))
			hash := encoder.Sum(nil)
			hashHex := hex.EncodeToString(hash)

			if ok := user.locker.Lock(hashHex, time.Second*5); ok {
				if _, err := user.hub.botClient.Send(msg); err != nil {
					log.Printf("[ERROR][telebot] failed to send message to user %d: %v \r\n", user.chatId, err)
				}
			} else {
				log.Printf("[DEBUG][telebot] same message is broadcasting, skip broadcast \r")
			}
		}
	}()
}
