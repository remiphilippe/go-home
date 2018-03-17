package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/golang/glog"
	goh4 "github.com/remiphilippe/go-h4"
)

type Annotation struct {
	IP         string            `json:"ip"`
	Attributes map[string]string `json:"attributes"`
}

type Filter struct {
	Type    string              `json:"type"`
	Filters []map[string]string `json:"filters"`
}

type Search struct {
	Filter *Filter `json:"filter"`
	Scope  string  `json:"scopeName,omitempty"`
	Limit  int     `json:"limit"`
	Offset int     `json:"offset,omitempty"`
}

type Tetration struct {
	h4 *goh4.H4
}

func TetrationFactory(config *Config) *Tetration {

	h := new(Tetration)

	h.h4 = new(goh4.H4)
	h.h4.Secret = config.APISecret
	h.h4.Key = config.APIKey
	h.h4.Endpoint = config.APIEndpoint
	h.h4.Verify = config.APIVerify
	h.h4.Prefix = "/openapi/v1"

	return h
}

func (h *Tetration) getSensorIP(hostname string) []map[string]string {
	f0 := make(map[string]string)
	f0["type"] = "contains"
	f0["field"] = "hostname"
	f0["value"] = hostname

	filter := &Filter{
		Type:    "and",
		Filters: []map[string]string{f0},
	}

	search := Search{
		Limit:  100,
		Filter: filter,
	}
	jsonStr, err := json.Marshal(&search)
	if err != nil {
		glog.Errorf("Error Marshalling search %s", err)
	}
	postResp := h.h4.Post("/inventory/search", fmt.Sprintf("%s", jsonStr))

	/*
	   (map[string]interface {}) (len=1) {
	    (string) (len=7) "results": ([]interface {}) (len=1 cap=1) {
	     (map[string]interface {}) (len=24) {
	      (string) (len=10) "os_version": (string) (len=3) "7.4",
	      (string) (len=15) "tags_scope_name": ([]interface {}) (len=4 cap=4) {
	       (string) (len=7) "Default",
	       (string) (len=13) "Default:SJC15",
	       (string) (len=25) "Default:SJC15:Attack Demo",
	       (string) (len=32) "Default:SJC15:Lab Infrastructure"
	      },
	      (string) (len=9) "user_role": (interface {}) <nil>,
	      (string) (len=8) "user_vpc": (string) (len=7) "esx-dmz",
	      (string) (len=12) "address_type": (string) (len=4) "IPV4",
	      (string) (len=2) "os": (string) (len=6) "CentOS",
	      (string) (len=12) "user_TA_zeus": (interface {}) <nil>,
	      (string) (len=19) "user_classification": (string) (len=6) "public",
	      (string) (len=2) "ip": (string) (len=12) "172.20.0.190",
	      (string) (len=7) "netmask": (string) (len=13) "255.255.255.0",
	      (string) (len=16) "tags_is_internal": (string) (len=4) "true",
	      (string) (len=8) "user_app": (interface {}) <nil>,
	      (string) (len=13) "user_location": (string) (len=5) "sjc15",
	      (string) (len=10) "user_scope": (interface {}) <nil>,
	      (string) (len=6) "vrf_id": (string) (len=1) "1",
	      (string) (len=9) "host_name": (string) (len=16) "centos7-rephilip",
	      (string) (len=9) "iface_mac": (string) (len=17) "00:50:56:8f:1c:92",
	      (string) (len=13) "user_TA_cymru": (interface {}) <nil>,
	      (string) (len=14) "user_lifecycle": (interface {}) <nil>,
	      (string) (len=11) "user_public": (interface {}) <nil>,
	      (string) (len=8) "uuid_src": (string) (len=6) "SENSOR",
	      (string) (len=8) "vrf_name": (string) (len=7) "Default",
	      (string) (len=9) "host_uuid": (string) (len=40) "39640178054cca9f98539e5ea17ecf99b005b971",
	      (string) (len=10) "iface_name": (string) (len=6) "ens192"
	     }
	    }
	   }

	*/

	jsonResp := make(map[string]interface{})
	err = json.Unmarshal(postResp, &jsonResp)
	if err != nil {
		glog.Errorf("Error unmarshalling json %s", err)
	}

	var entry map[string]string
	var res []map[string]string

	for _, v := range jsonResp["results"].([]interface{}) {
		e := v.(map[string]interface{})
		entry = make(map[string]string)
		entry["scope"] = e["vrf_name"].(string)
		entry["ip"] = e["ip"].(string)
		res = append(res, entry)
	}

	return res
}

func (h *Tetration) Isolate(ip string, scope string) error {
	attributes := make(map[string]string)
	attributes["quarantine"] = "isolate"

	annotation := Annotation{
		IP:         ip,
		Attributes: attributes,
	}
	jsonStr, err := json.Marshal(&annotation)
	if err != nil {
		log.Fatal(err)
	}
	h.h4.Post(fmt.Sprintf("/inventory/tags/%s", scope), fmt.Sprintf("%s", jsonStr))

	return nil
}
