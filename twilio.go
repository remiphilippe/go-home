package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/golang/glog"
)

// Twilio Interact with Twilio Service
type Twilio struct {
	URL   string
	Token string
	SID   string
	From  string
	To    string
}

// NewTwilio Factory for Twilio
func NewTwilio(token, sid, from string) *Twilio {
	t := new(Twilio)
	t.Token = token
	t.SID = sid
	t.From = from
	t.URL = "https://api.twilio.com/2010-04-01/Accounts/" + sid + "/Messages.json"

	return t
}

// SendSMS Send a SMS via Twilio
func (t *Twilio) SendSMS(to, message string) error {
	sms := url.Values{}
	sms.Set("To", to)
	sms.Set("From", t.From)
	sms.Set("Body", message)

	smsReader := *strings.NewReader(sms.Encode())
	client := &http.Client{}
	req, _ := http.NewRequest("POST", t.URL, &smsReader)
	req.SetBasicAuth(t.SID, t.Token)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var data map[string]interface{}
		decoder := json.NewDecoder(resp.Body)
		twilioErr := decoder.Decode(&data)
		if twilioErr == nil {
			glog.Infof("SID: %s\n", data["sid"])
		} else {
			glog.Errorf("Twilio Error: %s\n", twilioErr)
			return err
		}
	} else {
		glog.Errorf("Status: %s Error: %+v\n", resp.Status, err)
		return err
	}

	return nil
}
