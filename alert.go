package main

import (
	"encoding/json"
	"reflect"

	"github.com/golang/glog"
)

/*
(main.Alert) {
 Severity: (string) (len=8) "CRITICAL",
 TenantID: (int64) 0,
 AlertTime: (int64) 1521075398920,
 AlertText: (string) (len=28) "Meltdown on centos7-rephilip",
 KeyID: (string) (len=88) "39640178054cca9f98539e5ea17ecf99b005b971:0:5aa7738d755f023c45421905:28802:SIDE_CHANNEL:0",
 RootScopeID: (string) (len=24) "5aa6e27e497d4f3d90c6c38e",
 Type: (string) (len=20) "FORENSIC_RULE_ENGINE",
 EventTime: (int64) 0,
 AlertDetail: (string) (len=879) "{\"Sensor Id\":\"39640178054cca9f98539e5ea17ecf99b005b971\",\"Hostname\":\"centos7-rephilip\",\"Process Id\":28802,\"scope_id\":\"5aa6e27e497d4f3d90c6c38e\",\"forensic\":{\"Side Channel\":\"true\",\"Process Info - Command String\":\"./kaslr \",\"Process Info - Exec Path\":\"/home/rephilip/code/meltdown/kaslr\",\"Side Channel - Source - Meltdown\":\"true\"},\"profile\":{\"id\":\"5aa7739e497d4f3c972454d8\",\"name\":\"Security Demo\",\"created_at\":1520923550,\"updated_at\":1520924354,\"root_app_scope_id\":\"5aa6e27e497d4f3d90c6c38e\"},\"rule\":{\"id\":\"5aa7738d755f023c45421905\",\"name\":\"Meltdown\",\"clause_chips\":\"[{\\\"type\\\":\\\"filter\\\",\\\"facet\\\":{\\\"field\\\":\\\"event_type\\\",\\\"title\\\":\\\"Event type\\\",\\\"type\\\":\\\"STRING\\\"},\\\"operator\\\":{\\\"label\\\":\\\"\\u003d\\\",\\\"type\\\":\\\"eq\\\"},\\\"displayValue\\\":\\\"Side Channel\\\",\\\"value\\\":\\\"Side Channel\\\"}]\",\"created_at\":1520923533,\"updated_at\":1520983766,\"root_app_scope_id\":\"5aa6e27e497d4f3d90c6c38e\"}}"
}
*/

// AlertDetail - Tetration Alert Details
type AlertDetail struct {
	SensorID  string  `json:"Sensor Id"`
	ScopeID   string  `json:"scope_id"`
	Hostname  string  `json:"Hostname"`
	ProcessID float64 `json:"Process Id"`
	Forensic  string  `json:"forensic"`
}

// Alert - Tetration Alert format
/*
	(string) (len=1339) "{\"severity\":\"CRITICAL\",\"tenant_id\":0,\"alert_time\":1521075398920,\"alert_text\":\"Meltdown on centos7-rephilip\",\"key_id\":\"39640178054cca9f98539e5ea17ecf99b005b971:0:5aa7738d755f023c45421905:28802:SIDE_CHANNEL:0\",\"root_scope_id\":\"5aa6e27e497d4f3d90c6c38e\",\"type\":\"FORENSIC_RULE_ENGINE\",\"event_time\":0,\"alert_details\":\"{\\\"Sensor Id\\\":\\\"39640178054cca9f98539e5ea17ecf99b005b971\\\",\\\"Hostname\\\":\\\"centos7-rephilip\\\",\\\"Process Id\\\":28802,\\\"scope_id\\\":\\\"5aa6e27e497d4f3d90c6c38e\\\",\\\"forensic\\\":{\\\"Side Channel\\\":\\\"true\\\",\\\"Process Info - Command String\\\":\\\"./kaslr \\\",\\\"Process Info - Exec Path\\\":\\\"/home/rephilip/code/meltdown/kaslr\\\",\\\"Side Channel - Source - Meltdown\\\":\\\"true\\\"},\\\"profile\\\":{\\\"id\\\":\\\"5aa7739e497d4f3c972454d8\\\",\\\"name\\\":\\\"Security Demo\\\",\\\"created_at\\\":1520923550,\\\"updated_at\\\":1520924354,\\\"root_app_scope_id\\\":\\\"5aa6e27e497d4f3d90c6c38e\\\"},\\\"rule\\\":{\\\"id\\\":\\\"5aa7738d755f023c45421905\\\",\\\"name\\\":\\\"Meltdown\\\",\\\"clause_chips\\\":\\\"[{\\\\\\\"type\\\\\\\":\\\\\\\"filter\\\\\\\",\\\\\\\"facet\\\\\\\":{\\\\\\\"field\\\\\\\":\\\\\\\"event_type\\\\\\\",\\\\\\\"title\\\\\\\":\\\\\\\"Event type\\\\\\\",\\\\\\\"type\\\\\\\":\\\\\\\"STRING\\\\\\\"},\\\\\\\"operator\\\\\\\":{\\\\\\\"label\\\\\\\":\\\\\\\"\\\\u003d\\\\\\\",\\\\\\\"type\\\\\\\":\\\\\\\"eq\\\\\\\"},\\\\\\\"displayValue\\\\\\\":\\\\\\\"Side Channel\\\\\\\",\\\\\\\"value\\\\\\\":\\\\\\\"Side Channel\\\\\\\"}]\\\",\\\"created_at\\\":1520923533,\\\"updated_at\\\":1520983766,\\\"root_app_scope_id\\\":\\\"5aa6e27e497d4f3d90c6c38e\\\"}}\"}"
*/
type Alert struct {
	Severity    string       `json:"severity"`
	TenantID    float64      `json:"tenant_id"`
	AlertTime   float64      `json:"alert_time"`
	AlertText   string       `json:"alert_text"`
	KeyID       string       `json:"key_id"`
	RootScopeID string       `json:"root_scope_id"`
	Type        string       `json:"type"`
	EventTime   float64      `json:"event_time"`
	AlertDetail *AlertDetail `json:"alert_details,ignore"`
}

func tagDecode(e interface{}, m map[string]interface{}) {
	r := reflect.ValueOf(e).Elem()

	glog.V(3).Infof("Total fields: %d", r.NumField())

	for i := 0; i < r.NumField(); i++ {
		valueField := r.Field(i)
		typeField := r.Type().Field(i)
		tag := string(typeField.Tag.Get("json"))

		glog.V(3).Infof("Tag: %s Type: %s", tag, reflect.TypeOf(m[tag]))

		if tag != "" {
			//value := valueField.Interface().(string)
			//fmt.Println("type:", reflect.TypeOf(m[tag]))
			//name := typeField.Name
			if m[tag] != nil {
				switch m[tag].(type) {
				case string:
					glog.V(3).Infof("Tag %s Value %s \n", tag, m[tag].(string))
					valueField.SetString(m[tag].(string))
				case int64:
					glog.V(3).Infof("Tag %s Value %d \n", tag, m[tag].(int64))
					valueField.SetInt(m[tag].(int64))
				case float64:
					glog.V(3).Infof("Tag %s Value %f \n", tag, m[tag].(float64))
					valueField.SetFloat(m[tag].(float64))
				}

			}

		}
	}
}

// UnmarshalJSON Custom JSON decoder to handle the nested elements and the escaping of Kafka.
func (a *Alert) UnmarshalJSON(b []byte) error {
	mapAlert := make(map[string]interface{})
	err := json.Unmarshal(b, &mapAlert)
	if err != nil {
		return err
	}

	glog.V(2).Infoln("Decoding Alert")
	tagDecode(a, mapAlert)

	d := new(AlertDetail)
	mapDetail := make(map[string]interface{})
	err = json.Unmarshal([]byte(mapAlert["alert_details"].(string)), &mapDetail)
	if err != nil {
		return err
	}

	glog.V(2).Infoln("Decoding Details")
	tagDecode(d, mapDetail)

	a.AlertDetail = d

	return nil
}
