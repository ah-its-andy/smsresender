package tastek

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ah-its-andy/goconf"
	"github.com/ah-its-andy/smsresender/dao"
	"github.com/ah-its-andy/smsresender/db"
	"gorm.io/gorm"
)

var dbconn *DbConn
var dbconn2 *DbConn

func InitDbConn() {
	gdb, err := db.OpenConnection(db.DefaultOptions())
	if err != nil {
		log.Panic(err)
	}
	dbconn = &DbConn{
		Conn: gdb,
	}
	gdb2, err := db.OpenConnection(db.DefaultOptions())
	if err != nil {
		log.Panic(err)
	}
	dbconn2 = &DbConn{
		Conn: gdb2,
	}
}

type DbConn struct {
	Conn *gorm.DB

	inUse int32

	m sync.Mutex
}

func (conn *DbConn) Release() {
	atomic.StoreInt32(&conn.inUse, 0)
}

func (conn *DbConn) Retrive() bool {
	conn.m.Lock()
	defer conn.m.Unlock()

	for {
		if conn.inUse == 0 {
			break
		}
	}
	atomic.StoreInt32(&conn.inUse, 1)
	return true
}

func CronlyPullMessage() {
	for {
		if err := PullMessages(); err != nil {
			log.Printf("[ERROR] Failed to pull message: %v", err)
		}
		time.Sleep(time.Second * 5)
	}
}

func PullMessages() error {
	dbconn2.Retrive()
	defer dbconn2.Release()

	smslist, err := dao.GetSmsList(dbconn.Conn, 10, 20)
	if err != nil {
		return err
	}

	for _, s := range smslist {
		if err = SendMessage(s.ID); err != nil {
			log.Printf("[ERROR] SendMessage failed: %v", err)
		}
	}

	return nil
}

func SendMessage(id uint) error {
	dbconn.Retrive()
	defer dbconn.Release()

	err := dbconn.Conn.Transaction(func(tx *gorm.DB) error {
		model, err := dao.GetSmsById(tx, id)
		if err != nil {
			return err
		}
		if model == nil {
			return nil
		}
		if model.State == 20 {
			return nil
		}

		model.State = 20
		if err := dao.UpdateSms(tx, model); err != nil {
			return err
		}

		subject := fmt.Sprintf("%s 有新消息", model.Device)
		content := strings.Builder{}
		content.WriteString("來自：")
		content.WriteString(model.Sender)
		content.WriteString("\r\n")
		content.WriteString("時間：")
		content.WriteString(model.RecTime)
		content.WriteString("\r\n")
		content.WriteString("--------")
		content.WriteString("\r\n")
		content.WriteString("正文：\r\n")
		content.WriteString(model.Content)
		if err := sendPushPlus(subject, content.String()); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = dbconn.Conn.Transaction(func(tx *gorm.DB) error {
		model, err := dao.GetSmsById(tx, id)
		if err != nil {
			return err
		}
		if model == nil {
			return nil
		}

		if model.State == 30 {
			return nil
		}

		model.State = 30
		if err := dao.UpdateSms(tx, model); err != nil {
			return err
		}

		content := strings.Builder{}
		content.WriteString("設備：")
		content.WriteString(model.Device)
		content.WriteString("\r\n")
		content.WriteString("來自：")
		content.WriteString(model.Sender)
		content.WriteString("\r\n")
		content.WriteString("時間：")
		content.WriteString(model.RecTime)
		content.WriteString("\r\n")
		content.WriteString("--------")
		content.WriteString("\r\n")
		content.WriteString("正文：\r\n")
		content.WriteString(model.Content)
		if err := sendTelebotMessage(content.String()); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func sendTelebotMessage(message string) error {
	host := "https://api.telegram.org/bot"
	botToken := goconf.GetStringOrDefault("telebot.smsbot.token", "")
	if len(botToken) == 0 {
		return fmt.Errorf("bot token is required")
	}
	chatIds := goconf.GetStringOrDefault("telebot.smsbot.users", "")
	if len(chatIds) == 0 {
		return fmt.Errorf("chat id is required")
	}
	chatIdSlice := strings.Split(chatIds, ",")

	for _, chatId := range chatIdSlice {
		payload := url.Values{}
		payload.Set("chat_id", chatId)
		payload.Set("text", message)
		reqUrl := fmt.Sprintf("%s%s/sendMessage?%s", host, botToken, payload.Encode())
		client := GetProxiedHttpClient()
		resp, err := client.Get(reqUrl)
		if err != nil {
			return err
		}
		if resp.StatusCode > 299 {
			return fmt.Errorf("response status code is %s", resp.Status)
		}
	}
	return nil
}

func GetProxiedHttpClient() *http.Client {
	proxyUrl := goconf.GetStringOrDefault("proxy.url", "")
	if len(proxyUrl) == 0 {
		return http.DefaultClient
	}
	purl, err := url.Parse(proxyUrl)
	if err != nil {
		return http.DefaultClient
	}
	transport := http.Transport{
		Proxy: http.ProxyURL(purl),
	}
	return &http.Client{
		Transport: &transport,
	}
}

func sendPushPlus(subject, msg string) error {
	//http://www.pushplus.plus/send?token=7dc703f6b5694ecaa36efaa293dc4a6b&title=XXX&content=XXX&template=html
	token := goconf.GetStringOrDefault("pushplus.token", "")
	if len(token) == 0 {
		return fmt.Errorf("pushplus.token cannot be empty")
	}
	reqUrl := "http://www.pushplus.plus/send"
	payload := map[string]any{
		"token":    token,
		"title":    subject,
		"content":  msg,
		"template": "txt",
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return err
	}
	req, err := http.NewRequest("POST", reqUrl, &buf)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("http response: %v", resp.Status)
	}

	return nil
}
