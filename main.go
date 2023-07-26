package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ah-its-andy/goconf"
	physicalfile "github.com/ah-its-andy/goconf/physicalFile"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gopkg.in/fsnotify.v1"
)

func main() {
	configFilePath := flag.String("c", "/etc/smsresender/config.yml", "path to config")
	flag.Parse()
	// initialize on application startup
	goconf.Init(func(b goconf.Builder) {
		b.AddSource(physicalfile.Yaml(*configFilePath)).AddSource(goconf.EnvironmentVariable(""))
	})

	WatchIncomesMessage(context.Background())

	select {}
}

type IncomesMessage struct {
	// Sender   string
	// Date     string
	// Time     string
	// Serial   string
	// Sequence string
	// Type     string //ext
	Dir      string
	FileName string
	Content  string
}

// 0 IN20230725
// 1 174425
// 2 00
// 3 CMHK
// 4 00.txt
func ParseIncomesMessage(s string) (*IncomesMessage, error) {
	dir := filepath.Dir(s)
	fileName := filepath.Base(s)
	ext := filepath.Ext(fileName)
	if !strings.EqualFold(ext, ".txt") {
		log.Printf("[DEBUG] File %s is not a text message, skip. \r\n", fileName)
		return &IncomesMessage{
			Dir:      dir,
			Content:  fmt.Sprintf("File '%s' is not a text message, please check file manually.", fileName),
			FileName: filepath.Base(s),
		}, nil
	}

	content, err := ReadFile(s)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", s, err)
	}

	return &IncomesMessage{
		Dir:      dir,
		Content:  string(content),
		FileName: filepath.Base(s),
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

func SendToTgbot(chatId int64, msgText string) error {
	client, err := tgbotapi.NewBotAPI(goconf.GetStringOrDefault("telebot.smsbot.token", ""))
	if err != nil {
		return err
	}
	msg := tgbotapi.NewMessage(chatId, msgText)
	if _, err := client.Send(msg); err != nil {
		return fmt.Errorf("SendToTgbot: %v", err)
	}
	return nil
}

func ResendIncomesMessage(im *IncomesMessage) error {
	builder := strings.Builder{}
	neatName := im.FileName[:len(im.FileName)-len(filepath.Ext(im.FileName))]
	parts := strings.Split(neatName, "_")
	//IN20230725_174431_00_+852193193_00
	columnNames := []string{
		"Date",
		"Time",
		"-",
		"Sender",
		"Index",
	}
	for i, part := range parts {
		if i+1 >= len(columnNames) || columnNames[i] == "-" {
			continue
		}
		builder.WriteString(columnNames[i])
		builder.WriteString(": ")
		builder.WriteString(part)
		builder.WriteString("\r\n")
	}
	builder.WriteString("Content:\r\n")
	builder.Write([]byte(im.Content))

	if err := SendToTgbot(1034079183, builder.String()); err != nil {
		return err
	}
	if err := SendPushPlus("收到新消息", builder.String()); err != nil {
		return err
	}

	return nil
}

func MoveFile(source, target string) error {
	// 检查目标目录是否存在
	targetDir := filepath.Dir(target)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return err
	}

	// 检查目标文件是否存在
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		// 目标文件存在,删除目标文件
		err := os.Remove(target)
		if err != nil {
			return err
		}
	}

	// 复制文件到目标路径
	err := CopyFile(source, target)
	if err != nil {
		return err
	}

	// 删除源文件
	return os.Remove(source)
}

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func WatchIncomesMessage(ctx context.Context) {
	incomesDir := goconf.GetStringOrDefault("incomesdir", "/etc/smsresenderd/incomes")
	log.Printf("Starting to watch inbox directory: %s\r\n", incomesDir)
	for {
		files, err := ioutil.ReadDir(incomesDir)
		if err != nil {
			panic(err)
		}

		if len(files) > 0 {
			for _, file := range files {
				if file.IsDir() {
					continue
				}
				event := filepath.Join(incomesDir, file.Name())
				log.Printf("[DEBUG] Found file %v", event)
				incomesMsg, err := ParseIncomesMessage(event)
				if err != nil {
					log.Printf("Error parsing incomes message from %s", event)
					return
				}
				if incomesMsg != nil {
					if err := ResendIncomesMessage(incomesMsg); err != nil {
						log.Printf("Error resending income message from %s, err: %v", event, err)
						return
					}
					sentDir := goconf.GetStringOrDefault("sentdir", "/etc/smsresenderd/sent")
					if err := MoveFile(filepath.Join(incomesMsg.Dir, incomesMsg.FileName), filepath.Join(sentDir, incomesMsg.FileName)); err != nil {
						log.Printf("Error resending income message from %s, err: %v", event, err)
					}
				}
			}
		}
		time.Sleep(time.Second * 5)
	}
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

func SendPushPlus(title, msgText string) error {
	data := map[string]string{
		"token":    goconf.GetStringOrDefault("pushplus.token", ""),
		"title":    title,
		"content":  msgText,
		"template": "txt",
		"channel":  "wechat",
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		return err
	}
	resp, err := http.Post("https://www.pushplus.plus/send", "application/json", &buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode > 299 {
		return fmt.Errorf("send Push Plus Response Error: [%d] %s", resp.StatusCode, string(respContent))
	}
	return nil
}
