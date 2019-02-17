package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

var (
	httpClient = http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 10,
		},
		Timeout: time.Duration(10 * time.Second),
	}
	badhttpClient = http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 10,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: time.Duration(10 * time.Second),
	}

	zoneID     string
	zoneName   string
	apikey     string
	recordname string
	email      string
	debug      bool
)

type DNSListResponse struct {
	Result     []DNSRecord `json:"result"`
	ResultInfo `json:"result_info"`
}

type ResultInfo struct {
	Count int `json:"count"`
	Total int `json:"total_count"`
}

type DNSRecord struct {
	ID      string `json:"id,omitempty"`
	Type    string `json:"type,omitempty"`
	Name    string `json:"name,omitempty"`
	Content string `json:"content,omitempty"`
}

type ips struct {
	IP      string `json:"ip"`
	Address string `json:"address"`
}

func main() {

	flag.StringVar(&zoneID, "zoneid", "", "cloudflare zoneid")
	flag.StringVar(&apikey, "apikey", "", "cloudflare apikey")
	flag.StringVar(&recordname, "recordname", "", "cloudflare A or AAAA recordname")
	flag.StringVar(&email, "email", "", "cloudflare email address")
	flag.StringVar(&zoneName, "zonename", "", "name of the dns zone")
	flag.BoolVar(&debug, "debug", false, "should debug logging be switched on")
	flag.Parse()

	ipv4Update()
	ipv6Update()

}

func ipv4Update() {
	ip := getIP()
	//Grab the ID of the DNS record we want from cloudflare
	//TODO Compare IP addresses and if its the same do nothing
	id, ipadd := getARecordID("A")
	if ipadd == ip.IP {
		fmt.Println("Our A record is up to date so exiting")
		return
	}

	//Use the IP address we have got to update a record in cloudflare
	if len(id) == 0 {
		//We dont have a record so we create it
		fmt.Println("Creating", recordname, "with IP Address", ip.IP)
		createRecord(ip.IP, "A")
	} else {
		//We have a record so we must update it
		fmt.Println("Updating", recordname, "with IP Address", ip.IP)
		updateRecords(id, ip.IP, "A")
	}

}

func ipv6Update() {

	// Get the IP address from ident.me
	ip6 := getIPv6()

	id6, ipadd6 := getARecordID("AAAA")
	if ipadd6 == ip6.Address {
		fmt.Println("Our AAAA records is up to date so exiting")
		return
	}

	//Use the IP address we have got to update a record in cloudflare
	if len(id6) == 0 {
		//We dont have a record so we update it
		fmt.Println("Creating", recordname, "with IP Address", ip6.Address)
		createRecord(ip6.Address, "AAAA")
	} else {
		//We have a record so we must update it
		fmt.Println("Updating", recordname, "with IP Address", ip6.Address)
		updateRecords(id6, ip6.Address, "AAAA")
	}
}

func getARecordID(recordType string) (string, string) {

	safe := url.QueryEscape(recordname + "." + zoneName)

	req, err := http.NewRequest("GET", "https://api.cloudflare.com/client/v4/zones/"+zoneID+"/dns_records?type="+recordType+"&name="+safe, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("X-Auth-Email", email)
	req.Header.Set("X-Auth-Key", apikey)
	req.Header.Add("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	ipBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Could not read body from api.ipify.orgr" + err.Error())
	}

	response := DNSListResponse{}
	json.Unmarshal(ipBody, &response)
	if response.ResultInfo.Count == 1 {
		return response.Result[0].ID, response.Result[0].Content
	}
	return "", ""

}

func createRecord(ip string, recordType string) {

	data := struct {
		Type    string `json:"type"`
		Name    string `json:"name"`
		Content string `json:"content"`
		Proxied bool   `json:"proxied"`
	}{
		recordType,
		recordname,
		ip,
		false,
	}

	stringy, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	b := bytes.NewBuffer(stringy)

	req, err := http.NewRequest("POST", "https://api.cloudflare.com/client/v4/zones/"+zoneID+"/dns_records", b)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("X-Auth-Email", email)
	req.Header.Set("X-Auth-Key", apikey)
	req.Header.Add("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	ipBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Could not read body from api.ipify.orgr" + err.Error())
	}
	if debug {
		fmt.Println(string(ipBody))
	}
}

func updateRecords(id string, ip string, recordType string) {

	data := struct {
		Type    string `json:"type"`
		Name    string `json:"name"`
		Content string `json:"content"`
		Proxied bool   `json:"proxied"`
	}{
		recordType,
		recordname,
		ip,
		false,
	}

	stringy, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	b := bytes.NewBuffer(stringy)

	req, err := http.NewRequest("PUT", "https://api.cloudflare.com/client/v4/zones/"+zoneID+"/dns_records/"+id, b)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("X-Auth-Email", email)
	req.Header.Set("X-Auth-Key", apikey)
	req.Header.Add("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	ipBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Could not read body from api.ipify.orgr" + err.Error())
	}
	if debug {
		fmt.Println(string(ipBody))
	}
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

func getIPv6() ips {

	//request our IP address from
	req, err := http.NewRequest("GET", "https://v6.ident.me/.json", nil)
	if err != nil {
		log.Fatal(err)
	}
	//This site has an invalid https cert so ignoring it for now
	resp, err := badhttpClient.Do(req)
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
