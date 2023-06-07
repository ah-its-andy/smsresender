package telebot

import (
	"encoding/base64"
	"encoding/json"
)

type MessageModel struct {
	Topic  string            `json:"topic"`
	Header map[string]string `json:"header"`
	Data   []byte
}

func (model MessageModel) MarshalJSON() ([]byte, error) {
	data := make(map[string]any)
	data["topic"] = model.Topic
	data["header"] = model.Header
	if len(model.Data) > 0 {
		data["data"] = base64.StdEncoding.EncodeToString(model.Data)
	} else {
		data["data"] = ""
	}
	return json.Marshal(data)
}

func (model *MessageModel) UnmarshalJSON(data []byte) error {
	var m map[string]any
	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}
	model.Topic = m["topic"].(string)
	model.Header = make(map[string]string)
	if headerMap, ok := m["header"].(map[string]any); ok {
		for k, v := range headerMap {
			model.Header[k] = v.(string)
		}
	}

	if m["data"] != "" {
		model.Data, err = base64.StdEncoding.DecodeString(m["data"].(string))
		if err != nil {
			return err
		}
	} else {
		model.Data = nil
	}
	return nil
}
