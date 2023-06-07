package smsresender

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/ah-its-andy/goconf"
	physicalfile "github.com/ah-its-andy/goconf/physicalFile"
	"github.com/mochi-co/mqtt/v2"
	"github.com/mochi-co/mqtt/v2/hooks/auth"
	"github.com/mochi-co/mqtt/v2/listeners"
	"github.com/mochi-co/mqtt/v2/packets"
)

func main() {
	configFilePath := flag.String("c", "/etc/smsresender/config.yml", "path to config")
	flag.Parse()
	// initialize on application startup
	goconf.Init(func(b goconf.Builder) {
		b.AddSource(physicalfile.Yaml(*configFilePath))
	})
	// Create the new MQTT Server.
	server := mqtt.New(nil)

	// Allow all connections.
	_ = server.AddHook(new(auth.AllowHook), nil)
	_ = server.AddHook(&SmsHook{}, nil)

	// Create a TCP listener on a standard port.
	tcp := listeners.NewTCP("t1", goconf.GetStringOrDefault("application.bind_addr", ":6070"), nil)
	err := server.AddListener(tcp)
	if err != nil {
		log.Fatal(err)
	}

	err = server.Serve()
	if err != nil {
		log.Fatal(err)
	}
}

type SmsHook struct {
	mqtt.HookBase
}

func (h *SmsHook) ID() string {
	return "sms"
}

// Provides indicates which hook methods this hook provides.
func (h *SmsHook) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqtt.OnPublish,
		mqtt.OnConnect,
		mqtt.OnDisconnect,
	}, []byte{b})
}

func (h *SmsHook) OnConnect(cl *mqtt.Client, pk packets.Packet) {
	sendMessage("Client Connect", fmt.Sprintf("Client %s on %s", cl.ID, "Connected"))
}

func (h *SmsHook) OnDisconnect(cl *mqtt.Client, err error, expire bool) {
	sendMessage("Client Disconnect", fmt.Sprintf("Client %s on %s", cl.ID, "Disconnected"))
}

func (h *SmsHook) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	msg := string(pk.Payload)
	sendMessage("NEW MESSAGE", msg)
	return packets.Packet{
		TopicName: "direct/reply",
		Payload:   []byte("Ok"),
	}, nil
}

func sendMessage(subject, msg string) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC] Error sending message: %v", err)
			}
		}()
		if err := sendPushPlus(subject, msg); err != nil {
			log.Printf("[ERROR] Send PushPlus message: %v", err)
		}

	}()
}

func sendPushPlus(subject, msg string) error {
	//http://www.pushplus.plus/send?token=7dc703f6b5694ecaa36efaa293dc4a6b&title=XXX&content=XXX&template=html
	token := goconf.GetStringOrDefault("pushplus.token", "")
	if len(token) == 0 {
		return fmt.Errorf("pushplus.token cannot be empty")
	}
	reqUrl := fmt.Sprintf("http://www.pushplus.plus/send")
	payload := map[string]any{
		"token":    token,
		"title":    subject,
		"content":  msg,
		"template": "text",
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
