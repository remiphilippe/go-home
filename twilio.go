package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang/glog"
)

// Twilio Interact with Twilio Service
type Twilio struct {
	URL      string
	Token    string
	SID      string
	From     string
	To       string
	Message  chan *Message
	MsgCount int
	Limit    int
}

// Message Message to be sent
type Message struct {
	Text string
	To   string
}

// NewTwilio Factory for Twilio
func NewTwilio(token, sid, from string, limit int) *Twilio {
	t := new(Twilio)
	t.Token = token
	t.SID = sid
	t.From = from
	t.URL = "https://api.twilio.com/2010-04-01/Accounts/" + sid

	t.Message = make(chan *Message)
	t.MsgCount = 0
	t.Limit = limit
	go t.listener()

	return t
}

func (t *Twilio) throttle() {
	timeout := 60
	timer := time.NewTimer(time.Second * time.Duration(timeout))
	<-timer.C
	t.MsgCount = 0
	glog.V(2).Infof("Timer reset after %d seconds\n", timeout)
}

func (t *Twilio) listener() {
	for {
		select {
		case msg := <-t.Message:
			if t.MsgCount < t.Limit {
				t.SendSMS(msg.To, msg.Text)
				t.MsgCount++

				go t.throttle()
			} else {
				glog.V(2).Infof("Throttle\n")
			}
		}
	}
}

// SendSMS Send a SMS via Twilio
func (t *Twilio) SendSMS(to, message string) error {
	var twilioURL string
	sms := url.Values{}
	sms.Set("To", to)
	sms.Set("From", t.From)
	sms.Set("Body", message)
	//TODO Not sure if we need this, check
	//sms.Set("MessagingServiceSid", "")

	smsReader := *strings.NewReader(sms.Encode())

	twilioURL = t.URL + "/Messages.json"

	client := &http.Client{}
	req, _ := http.NewRequest("POST", twilioURL, &smsReader)
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
