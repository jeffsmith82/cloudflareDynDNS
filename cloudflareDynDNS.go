package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var (
	httpClient = http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 10,
		},
		Timeout: time.Duration(10 * time.Second),
	}
)

type ips struct {
	IP string `json:"ip"`
}

func main() {

	zoneID := flag.String("zoneid", "", "cloudflare zoneid")
	apikey := flag.String("apikey", "", "cloudflare apikey")
	recordname := flag.String("recordname", "", "cloudflare A or AAAA recordname")
	flag.Parse()

	fmt.Println(zoneID)
	fmt.Println(apikey)

	// Get the IP address from ipify
	ip := getIP()

	//Grab the ID of the DNS reocrd we want from cloudflare
	id := getARecordID()
	//Use the IP address we have got to update a record in cloudflare
	if len(id) == 0 {
		//We dont have a record so we update it
		createRecord()
	} else {
		//We have a record so we must update it
		updateRecords()
	}

}

func getARecordID() string {

}

func createRecord() {
}

func updateRecords() {
}

func getIP() ips {

	//request our IP address from
	req, err := http.NewRequest("GET", "https://api.ipify.org?format=json", nil)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatal("Non 200 status code from api.ipify.org" + string(resp.StatusCode))
	}

	ipBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Could not read body from api.ipify.orgr" + err.Error())
	}

	ip := ips{}
	json.Unmarshal(ipBody, &ip)
	return ip
}
