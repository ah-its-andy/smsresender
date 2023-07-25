package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ah-its-andy/goconf"
	physicalfile "github.com/ah-its-andy/goconf/physicalFile"
	"gopkg.in/fsnotify.v1"

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

	Startup()

	select {}
}

type IncomesMessage struct {
	// Sender   string
	// Date     string
	// Time     string
	// Serial   string
	// Sequence string
	// Type     string //ext
	FileName string
}

// 0 IN20230725
// 1 174425
// 2 00
// 3 CMHK
// 4 00.txt
func ParseIncomesMessage(s string) (*IncomesMessage, error) {
	parts := strings.Split(s, "_")
	if len(parts) < 5 || !strings.HasPrefix(parts[0], "IN") || !strings.Contains(parts[4], ".") {
		return nil, fmt.Errorf("invalid message format")
	}

	return &IncomesMessage{
		// Date:     parts[0][2:],
		// Time:     parts[1],
		// Serial:   parts[2],
		// Sender:   parts[3],
		// Sequence: parts[4],
		// Type:     strings.Split(parts[5], ".")[1],
		FileName: s,
	}, nil
}

func ReadFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

func ResendIncomesMessage(im *IncomesMessage) error {
	content, err := ReadFile(im.FileName)
	if err != nil {
		return err
	}
	builder := strings.Builder{}
	builder.WriteString("FileName: ")
	builder.WriteString(im.FileName)
	builder.WriteString("\r\n")
	builder.WriteString("Content:\r\n")
	builder.WriteString("====== BEGIN ======\r\n")
	builder.Write(content)
	builder.WriteString("\r\n======= END =======")

	sentDir := goconf.GetStringOrDefault("sentdir", "/etc/smsresenderd/sent")
	destFile, err := os.Create(filepath.Join(sentDir, filepath.Base(im.FileName)))
	if err != nil {
		return err
	}
	defer destFile.Close()
	if _, err := destFile.Write(content); err != nil {
		return err
	}

	telebot.Broadcast("smsbot", builder.String())

	if err := os.Remove(im.FileName); err != nil {
		return err
	}

	return nil
}

func Startup() {
	c := WatchIncomesMessage(context.Background())
	for {
		msg := <-c
		if msg == nil {
			continue
		}
		ResendIncomesMessage(msg)
	}
}

func WatchIncomesMessage(ctx context.Context) <-chan *IncomesMessage {
	incomesDir := goconf.GetStringOrDefault("incomesdir", "/etc/smsresenderd/incomes")

	messages := make(chan *IncomesMessage)
	defer close(messages)
	events := WatchDir(ctx, incomesDir)
	for {
		event, ok := <-events
		if !ok {
			break
		}
		fileName := filepath.Base(event)
		incomesMsg, err := ParseIncomesMessage(fileName)
		if err != nil {
			log.Printf("Error parsing incomes message from %s", event)
			continue
		}
		messages <- incomesMsg
	}
	return messages
}

func WatchDir(ctx context.Context, dir string) <-chan string {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)
	events := make(chan string)

	go func() {
		defer close(done)
		defer close(events)

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					events <- event.Name
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			case <-ctx.Done():
				return
			}
		}
	}()

	err = watcher.Add(dir)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		<-ctx.Done()
		watcher.Close()
	}()

	return events
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
