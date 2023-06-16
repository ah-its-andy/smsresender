package tastek

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

func SignIn(session *Session, host, user, password string) error {
	reqUrl := fmt.Sprintf("http://%s/goform/goform_set_cmd_process", host)
	//isTest=false&goformId=LOGIN&password=MTIzNDU2Nzg%3D&username=YWRtaW4%3D
	payload := url.Values{}
	payload.Set("isTest", "false")
	payload.Set("goformId", "LOGIN")
	payload.Set("password", base64.StdEncoding.EncodeToString([]byte(password)))
	payload.Set("username", base64.StdEncoding.EncodeToString([]byte(user)))
	resp, err := session.PostForm(reqUrl, payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		return fmt.Errorf("response code: %d", resp.StatusCode)
	}
	return nil
}

func SignOut(session *Session, host string) error {
	reqUrl := fmt.Sprintf("http://%s/goform/goform_set_cmd_process", host)
	//isTest=false&goformId=LOGIN&password=MTIzNDU2Nzg%3D&username=YWRtaW4%3D
	payload := url.Values{}
	payload.Set("isTest", "false")
	payload.Set("goformId", "LOGOUT")
	resp, err := session.PostForm(reqUrl, payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		return fmt.Errorf("response code: %d", resp.StatusCode)
	}
	return nil
}

func DelSms(session *Session, host string, msgid string) error {
	//isTest=false&goformId=DELETE_SMS&msg_id=131%3B&notCallback=true
	payload := url.Values{}
	payload.Set("isTest", "false")
	payload.Set("goformId", "DELETE_SMS")
	payload.Set("msg_id", "DELETE_SMS")
	payload.Set("notCallback", "true")
	reqUrl := fmt.Sprintf("http://%s/goform/goform_set_cmd_process", host)
	resp, err := session.PostForm(reqUrl, payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		return fmt.Errorf("response code: %d", resp.StatusCode)
	}
	return nil
}

func SmsTotal(session *Session, host string) (*SmsTotalResult, error) {
	//isTest=false&cmd=sms_data_total&page=0&data_per_page=500&mem_store=1&tags=10&order_by=order+by+id+desc&_=1686124367934
	payload := url.Values{}
	payload.Set("isTest", "false")
	payload.Set("cmd", "sms_data_total")
	payload.Set("page", "0")
	payload.Set("data_per_page", "500")
	payload.Set("mem_store", "1")
	payload.Set("tags", "10")
	payload.Set("order_by", "order by id desc")
	payload.Set("_", fmt.Sprintf("%d", time.Now().UTC().UnixMilli()))
	reqUrl := fmt.Sprintf("http://%s/goform/goform_get_cmd_process", host)

	req, err := http.NewRequest("GET", reqUrl+"?"+payload.Encode(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := session.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("response code: %d", resp.StatusCode)
	}
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	result := SmsTotalResult{}
	if err := json.Unmarshal(content, &result); err != nil {
		return nil, err
	}

	for i, msg := range result.Messages {
		if len(msg.Content) == 0 {
			msg.Content = "没有内容"
		} else {
			decodeContent := decodeMessage(msg.Content, false)
			result.Messages[i].Content = string(decodeContent)
		}
	}

	return &result, nil
}

var specialChars = []string{"000D", "000A", "0009", "0000"}
var specialCharsIgnoreWrap = []string{"0009", "0000"}

func decodeMessage(str string, ignoreWrap bool) string {
	// check if the string is empty
	if str == "" {
		return ""
	}

	// define the specials slice based on ignoreWrap
	var specials []string
	if ignoreWrap {
		specials = specialCharsIgnoreWrap
	} else {
		specials = specialChars
	}

	// create a regular expression to match hex codes
	re := regexp.MustCompile(`([A-Fa-f0-9]{1,4})`)

	// replace each hex code with a character or an empty string
	return re.ReplaceAllStringFunc(str, func(match string) string {
		// check if the match is in the specials slice
		for _, s := range specials {
			if match == s {
				return ""
			}
		}

		// convert the match to an integer
		n, err := strconv.ParseInt(match, 16, 32)

		// check for errors
		if err != nil {
			return ""
		}

		// convert the integer to a rune and return it as a string
		return string(rune(n))
	})
}

type SmsTotalResult struct {
	Messages []*SmsContent `json:"messages"`
}

type SmsContent struct {
	ID      string `json:"id"`
	Number  string `json:"number"`
	Date    string `json:"date"`
	Content string `json:"content"`
}
