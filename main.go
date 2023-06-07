package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/ah-its-andy/goconf"
	physicalfile "github.com/ah-its-andy/goconf/physicalFile"
	"github.com/ah-its-andy/smsresender/tastek"
)

func main() {
	configFilePath := flag.String("c", "/etc/smsresender/config.yml", "path to config")
	flag.Parse()
	// initialize on application startup
	goconf.Init(func(b goconf.Builder) {
		b.AddSource(physicalfile.Yaml(*configFilePath)).AddSource(goconf.EnvironmentVariable(""))
	})

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
