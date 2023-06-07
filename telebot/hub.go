package telebot

import (
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MsgHandlers []*MsgHandler
type MsgPredicate func(string) bool
type MsgFinisher func(botName string, msg *tgbotapi.Message, reply ReplyFunc)
type ReplyFunc func(msgText string)

func (handers MsgHandlers) Get(msgText string) []*MsgHandler {
	ret := make([]*MsgHandler, 0)
	for _, handler := range handers {
		if handler != nil && handler.predicateFn(msgText) {
			ret = append(ret, handler)
		}
	}
	return ret
}

type MsgHandler struct {
	predicateFn MsgPredicate
	fn          MsgFinisher
}

func NewMsgHandler(predicate MsgPredicate, fn MsgFinisher) *MsgHandler {
	return &MsgHandler{
		predicateFn: predicate,
		fn:          fn,
	}
}

func StartsWith(text string) func(string) bool {
	return func(msgText string) bool {
		return strings.HasPrefix(msgText, text)
	}
}

func Wildcard() func(string) bool {
	return func(s string) bool {
		return true
	}
}

type Hub struct {
	botName   string
	users     map[int64]*User
	botClient *tgbotapi.BotAPI
	opts      *Options
	handlers  MsgHandlers
}

func NewHub(botName string, opts *Options) (*Hub, error) {
	bot, err := tgbotapi.NewBotAPI(opts.Token)
	if err != nil {
		return nil, err
	}
	hub := &Hub{
		botName:   botName,
		users:     make(map[int64]*User),
		botClient: bot,
		opts:      opts,
		handlers:  make(MsgHandlers, 0),
	}

	if len(opts.UserIds) > 0 {
		for _, id := range opts.UserIds {
			hub.AddUser(NewUser(hub, id))
		}
	}

	return hub, nil
}

func (hub *Hub) AddUser(user *User) {
	hub.users[user.chatId] = user
	if strings.EqualFold(hub.opts.Mode, "server") {
		go user.Listen()
	}
	log.Printf("[DEBUG] user %d added to hub \r", user.chatId)
}

func (hub *Hub) RemoveUser(user *User) {
	delete(hub.users, user.chatId)
}

func (hub *Hub) Broadcast(msgText string) {
	for _, user := range hub.users {
		if user.chatId == 0 {
			log.Printf("[telebot] user %d is not ready, skip broadcast \r")
			continue
		}
		log.Printf("[telebot] broadcast to user %d: %s \r", user.chatId, msgText)
		msg := tgbotapi.NewMessage(user.chatId, msgText)
		user.Reply(&msg)
	}
}

func (hub *Hub) Listen() {
	for {
		updateConfig := tgbotapi.NewUpdate(0)
		updateConfig.Timeout = 30
		updates := hub.botClient.GetUpdatesChan(updateConfig)
		for update := range updates {
			if update.Message == nil {
				continue
			}

			chatId := update.Message.Chat.ID
			log.Printf("[telebot] received message from user %d: %s \r\n", chatId, update.Message.Text)

			if chatId <= 0 {
				continue
			}

			if !strings.HasPrefix(update.Message.Text, "/") {
				continue
			}

			allow := false
			var user *User
			for c, u := range hub.users {
				if chatId == c {
					allow = true
					user = u
				}
			}
			if !allow {
				continue
			}
			handlers := hub.handlers.Get(update.Message.Text)
			if len(handlers) > 0 {
				hub.safeExec(hub.botName, handlers, update.Message, user)
			}

			// if _, ok := hub.users[chatId]; !ok {
			// 	user := NewUser(hub, chatId)
			// 	hub.AddUser(user)
			// }
		}
	}
}

func (hub *Hub) safeExec(botName string, handlers []*MsgHandler, msg *tgbotapi.Message, user *User) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("[PANIC][telebot] broadcast failed: %v", err)
		}
	}()

	for _, handler := range handlers {
		handler.fn(botName, msg, func(msgText string) {
			newMsg := tgbotapi.NewMessage(user.chatId, msgText)
			user.Reply(&newMsg)
		})
	}
}

// 实现一个ReentrantLock，实现指定的资源KEY在一段时间内只能执行一次
type ReentrantLock struct {
	mutex   *sync.Mutex
	lockMap map[string]int
	running int32
}

func NewReentrantLock() *ReentrantLock {
	locker := &ReentrantLock{
		mutex:   &sync.Mutex{},
		lockMap: make(map[string]int),
	}
	return locker
}

func (lock *ReentrantLock) Lock(key string, ttl time.Duration) bool {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()

	if _, ok := lock.lockMap[key]; ok {
		return false
	}

	lock.lockMap[key] = int(time.Now().UTC().Add(ttl).Unix())
	return true
}

func (lock *ReentrantLock) RunGC() {
	if ok := atomic.CompareAndSwapInt32(&lock.running, 0, 1); !ok {
		return
	}
	go lock.rungc()
}

func (lock *ReentrantLock) rungc() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("[PANIC][telebot] reentrant lock gc failed: %v", err)
		}
	}()

	timer := time.NewTicker(time.Minute)
	for {
		<-timer.C
		lock.collect()
	}
}

func (lock *ReentrantLock) collect() {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()

	now := time.Now().UTC().Unix()
	delKs := make([]string, 0)
	for k, v := range lock.lockMap {
		if v <= int(now) {
			delKs = append(delKs, k)
		}
	}
	for _, k := range delKs {
		delete(lock.lockMap, k)
	}
	if len(delKs) > 0 {
		log.Printf("[telebot] reentrant lock gc: %d keys deleted \r", len(delKs))
	}
}
