package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ah-its-andy/goconf"
	physicalfile "github.com/ah-its-andy/goconf/physicalFile"
	"github.com/ah-its-andy/smsresender/tastek"
	"github.com/ah-its-andy/smsresender/telebot"
	"github.com/ah-its-andy/smsresender/typeconv"
)

func main() {
	configFilePath := flag.String("c", "/etc/smsresender/config.yml", "path to config")
	flag.Parse()
	// initialize on application startup
	goconf.Init(func(b goconf.Builder) {
		b.AddSource(physicalfile.Yaml(*configFilePath)).AddSource(goconf.EnvironmentVariable(""))
	})
	InitTelebot("")

	devices, ok := goconf.GetSection("devices").GetRaw()
	if !ok {
		panic("device not found in config file")
	}
	deviceMap, ok := devices.(map[interface{}]interface{})
	if !ok {
		panic("device not found in config file")
	}
	for k, _ := range deviceMap {
		deviceName := fmt.Sprintf("%s", k)
		addr := goconf.GetStringOrDefault("devices."+deviceName+".addr", "")
		if len(addr) == 0 {
			panic("devices." + deviceName + ".addr is empty")
		}
		username := goconf.GetStringOrDefault("devices."+deviceName+".username", "")
		if len(addr) == 0 {
			panic("devices." + deviceName + ".username is empty")
		}
		password := goconf.GetStringOrDefault("devices."+deviceName+".password", "")
		if len(addr) == 0 {
			panic("devices." + deviceName + ".password is empty")
		}
		tok := goconf.GetStringOrDefault("devices."+deviceName+".tok", "")
		tokInterval := time.Second * 5
		if len(tok) > 0 {
			if t, err := time.ParseDuration(tok); err == nil {
				tokInterval = t
			}
		}
		channel := tastek.NewSmsChannel(deviceName, addr, username, password, tokInterval)
		go channel.Start()
	}

	select {}
}

func InitTelebot(env string) {
	disabled := os.Getenv("MONOREPO_DISABLE_TELEBOT")
	if strings.EqualFold(disabled, "1") {
		log.Printf("InitTelebot: MONOREPO_DISABLE_TELEBOT is set, telebot is disabled")
		return
	}
	telebotSec, ok := goconf.GetSection("telebot").GetRaw()
	if !ok {
		log.Printf("telebot config not found, telebot will not work")
		return
	}
	telebotMap, ok := telebotSec.(map[any]any)
	if !ok {
		fmt.Println("telebot config is not a map")
		os.Exit(-1)
	}
	telebotNames := make([]string, 0)
	for k, _ := range telebotMap {
		telebotNames = append(telebotNames, typeconv.MustString(k))
	}
	for _, name := range telebotNames {
		enabled := goconf.CastOrDefault("telebot."+name+".enabled", true, goconf.BooleanConversion).(bool)
		if !enabled {
			continue
		}
		token := goconf.GetStringOrDefault("telebot."+name+".token", "")
		if len(token) == 0 {
			continue
		}

		mode := goconf.GetStringOrDefault("telebot."+name+".mode", "client")
		users := goconf.GetStringOrDefault("telebot."+name+".users", "")
		listenForUser := goconf.CastOrDefault("telebot."+name+".listen", false, goconf.BooleanConversion).(bool)
		chatIds := make([]int64, 0)
		if len(users) > 0 {
			for _, id := range strings.Split(users, ",") {
				chatId, err := strconv.ParseInt(id, 10, 64)
				if err != nil {
					fmt.Println("telebot user id is not a number: " + id)
					os.Exit(-1)
				}
				chatIds = append(chatIds, chatId)
			}
		}

		telebot.AddBot(name, func(o *telebot.Options) {
			o.Token = token
			o.UserIds = chatIds
			o.ListenForUser = false
			o.UseRedisQueue = false
			o.Mode = mode
		})

		log.Printf("[telebot] telebot %s is started, listen for user: %v",
			name,
			listenForUser)
	}
}
