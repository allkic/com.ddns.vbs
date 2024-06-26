package main

import (
	"encoding/json"
	"fmt"
	alidns "github.com/alibabacloud-go/alidns-20150109/v4/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"net"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	logIp        = ""
	Type         = "AAAA"
	aliDnsServer = "alidns.cn-hangzhou.aliyuncs.com"
	Keys         = make(map[string]string)
)

func getIpv6() string {
	var ipv6 string = ""
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		return ipv6
	}
	for _, val := range addrs {
		i := regexp.MustCompile(`(\w+:){7}\w+/64`).FindString(val.String())
		if strings.Count(i, ":") == 7 {
			spl := strings.Split(i, "/")
			ipv6 = spl[0]
		}
	}
	return ipv6
}
func createClient() (result *alidns.Client, err error) {
	config := openapi.Config{
		AccessKeyId:     tea.String(Keys["accessKeyId"]),
		AccessKeySecret: tea.String(Keys["accessKeySecret"]),
	}
	config.Endpoint = tea.String(aliDnsServer)
	result = &alidns.Client{}
	result, err = alidns.NewClient(&config)
	return result, err
}
func updateDomain(client *alidns.Client, data *alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord, ipv6 string) {
	body := alidns.UpdateDomainRecordRequest{
		Type:     data.Type,
		RR:       data.RR,
		RecordId: data.RecordId,
		Value:    tea.String(ipv6),
	}
	runtime := &util.RuntimeOptions{}
	_, err := client.UpdateDomainRecordWithOptions(&body, runtime)
	if err != nil {
		println(err)
	}
}

func refreshDDNS(ipv6 string) {
	result, err := createClient()
	if err != nil {
		println(err, "000000000")
	}
	req := alidns.DescribeDomainRecordsRequest{
		DomainName: tea.String(Keys["DomainURL"]),
		Type:       tea.String(Type),
	}
	reuntime := &util.RuntimeOptions{}
	func() (e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				e = r
			}
		}()
		res, err := result.DescribeDomainRecordsWithOptions(&req, reuntime)
		if err != nil {
			return err
		}
		domains := map[string]bool{
			"test": true,
			"ssh":  true,
		}
		data := res.Body.DomainRecords.Record
		for _, item := range data {
			if item.Value != tea.String(ipv6) && domains[string(*item.RR)] {
				updateDomain(result, item, ipv6)
			}
		}
		return nil
	}()
}
func timer() {
	ipv6 := getIpv6()
	if ipv6 != logIp {
		logIp = ipv6
		file, _ := os.OpenFile("./config/auth.json", os.O_CREATE|os.O_RDWR, 0666)
		defer file.Close()
		decoder := json.NewDecoder(file)
		_ = decoder.Decode(&Keys)
		refreshDDNS(ipv6)
	}
	m := time.Second * 60
	time.Sleep(m)
	timer()
}
func main() {
	timer()
}
