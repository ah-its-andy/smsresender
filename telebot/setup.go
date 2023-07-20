package telebot

import (
	"log"
	"strings"
)

type Options struct {
	BotName       string
	Token         string
	UserIds       []int64
	ListenForUser bool
	UseRedisQueue bool
	Mode          string
}

var hubs map[string]*Hub

func AddBot(name string, fn func(*Options)) {
	if hubs == nil {
		hubs = make(map[string]*Hub)
	}
	opts := Options{
		BotName: name,
		UserIds: make([]int64, 0),
	}
	if fn != nil {
		fn(&opts)
	}
	hub, err := NewHub(name, &opts)
	if err != nil {
		log.Printf("[ERROR][telebot] add bot failed: %v", err)
		return
	}
	hubs[name] = hub
}

func Run() {
	if hubs == nil {
		return
	}
	for _, hub := range hubs {
		if hub.opts.ListenForUser {
			go hub.Listen()
		}
	}
}

func Broadcast(botName, msgText string) {
	if bot, ok := hubs[botName]; ok {
		if bot.opts.UseRedisQueue {
			broadcast(botName, msgText)
		}
	}
}

func broadcast(botName, msgText string) {
	if bot, ok := hubs[botName]; ok {
		go func() {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("[ERROR][telebot] broadcast failed: %v", err)
				}
			}()

			bot.Broadcast(msgText)
		}()
	}
}

func Consume(botName string, predicate MsgPredicate, finisher MsgFinisher) {
	msgHandler := NewMsgHandler(predicate, finisher)
	if strings.EqualFold(botName, "*") {
		for _, bot := range hubs {
			bot.handlers = append(bot.handlers, msgHandler)
		}
	} else if bot, ok := hubs[botName]; ok {
		bot.handlers = append(bot.handlers, NewMsgHandler(predicate, finisher))
	}
}

const AdminBotName = "adminbot"
