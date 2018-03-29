package main

import (
	"time"

	"github.com/golang/glog"
	goh4 "github.com/remiphilippe/go-h4"
)

// Endpoint This struct just holds a sensorid, it's used for isolation
type Endpoint struct {
	SensorID string
}

// Tetration Struct that holds our Tetration object
type Tetration struct {
	h4      *goh4.H4
	Isolate chan *Endpoint
}

// NewTetration Factory to initialize the Tetration object
func NewTetration(config *Config) *Tetration {
	h := new(Tetration)

	h.h4 = new(goh4.H4)
	h.h4.Secret = config.APISecret
	h.h4.Key = config.APIKey
	h.h4.Endpoint = config.APIEndpoint
	h.h4.Verify = config.APIVerify
	h.h4.Prefix = "/openapi/v1"

	// The channel is to create a queue and avoid running too many parallel threads
	h.Isolate = make(chan *Endpoint)
	go h.listener()

	return h
}

func (h *Tetration) listener() {
	for {
		select {
		case msg := <-h.Isolate:
			// Alert will return a sensor name and not IP, we need to get the IPs from inventory
			ips := h.getSensorIP(msg.SensorID)
			// a host could have multiple IPs
			for _, v := range ips {
				// Isolate VMs by assigning tags to IP (quarantine=isolate)
				h.isolate(v["ip"], v["scope"])

				// Slow down a bit the requests, kafka can send a lot of messages
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

func (h *Tetration) getSensorIP(hostUUID string) []map[string]string {
	// More info - https://<some tetration cluster>/documentation/ui/openapi/api_inventory.html#inventory-search
	// first we need to build a filter
	f0 := goh4.Filter{}
	f0.Type = "contains"
	f0.Field = "host_uuid"
	f0.Value = hostUUID

	q := goh4.InventoryQuery{
		Limit: 100,
		Filter: &goh4.QueryFilter{
			Type:    "and",
			Filters: []interface{}{&f0},
		},
	}
	s, err := h.h4.InventorySearch(&q)
	if err != nil {
		glog.Errorf("Inventory Search Error: %s", err.Error())
	}

	var entry map[string]string
	var res []map[string]string

	for _, v := range s {
		entry = make(map[string]string)
		entry["scope"] = v.Scopes[0]
		entry["ip"] = v.IP.String()
		res = append(res, entry)
	}

	return res
}

// Isolate set an isolation tag on a VM, this tag needs to be attached to a policy
func (h *Tetration) isolate(ip string, scope string) error {
	// Create a map of attributes to apply to a workload
	attributes := make(map[string]string)
	attributes["quarantine"] = "isolate"

	// Define where we want to attach this annotation
	annotation := goh4.Annotation{
		IP:         ip,
		Attributes: attributes,
		Scope:      scope,
	}

	glog.V(1).Infof("Isolating Host %s in scope %s\n", ip, scope)
	err := h.h4.AddAnnotation(&annotation)
	if err != nil {
		glog.Errorf("Error in Annotation: %s", err.Error())
		return err
	}

	return nil
}
