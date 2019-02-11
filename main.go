package main

import (
	"sync"
	"fmt"
	"time"
	"io/ioutil"
	"strings"
	"net/http"
	"net"
	"strconv"
	"context"
)

var HttpClientPool sync.Pool
var ipChan chan string

/**
103.21.244.0/22
103.22.200.0/22
103.31.4.0/22
- 104.16.0.0/12
- 108.162.192.0/18
131.0.72.0/22
- 141.101.64.0/18
- 162.158.0.0/15
- 172.64.0.0/13
173.245.48.0/20
188.114.96.0/20
190.93.240.0/20
197.234.240.0/22
- 198.41.128.0/17
 */

func main() {
	ipChan = make(chan string, 32)
	fmt.Println(ipToInt("192.168.1.1"))
	fmt.Println(IntToIp(2147483648))
	fmt.Println(ipWithMask("198.41.128.0/17"))
	start, length := ipWithMask("198.41.128.0/17")
	var i uint32
	for j := 0; j < 32; j++ {
		go routine1()
	}
	for i = 0; i < length; i += 16 {
		ipChan <- IntToIp(start + i)
	}
}

func routine1() {
	for {
		ipStr := <-ipChan
		fmt.Println(ipStr, "	", getCdnTrace(ipStr))
	}
}

func getCdnTrace(ipStr string) (colo string) {
	c := acquireHttpClient(3 * time.Second)
	resp, err := c.Get("http://" + ipStr + "/cdn-cgi/trace")
	if err != nil {
		return
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	bodyStr := string(b)
	s := strings.Index(bodyStr, "colo=")
	colo = bodyStr[s : s+8]
	if colo != "colo=HKG" && colo != "colo=SEA" && colo != "colo=LAX" && colo != "colo=NRT" {
		colo += "FIND IT!!!"
	}
	return
}

func acquireHttpClient(timeout time.Duration) (http.Client) {
	c := HttpClientPool.Get()
	if c == nil {
		c = http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, netw, addr string) (net.Conn, error) {
					c, err := net.DialTimeout(netw, addr, time.Second*3)
					if err != nil {
						return nil, err
					}
					deadline := time.Now().Add(timeout)
					c.SetDeadline(deadline)
					return c, nil
				},
			},
		}
	}
	return c.(http.Client)
}

func putHttpClient(client http.Client) {
	HttpClientPool.Put(client)
}

func ipWithMask(ipMaskStr string) (start uint32, length uint32) {
	s := strings.Split(ipMaskStr, "/")
	if len(s) != 2 {
		return 0, 0
	}
	maskLen, _ := strconv.Atoi(s[1])
	return ipToInt(s[0]), 0xffffffff>>uint32(maskLen) + 1
}

func ipToInt(ipStr string) (ipInt uint32) {
	s := strings.Split(ipStr, ".")
	if len(s) != 4 {
		return 0
	}
	b0, _ := strconv.Atoi(s[0])
	b1, _ := strconv.Atoi(s[1])
	b2, _ := strconv.Atoi(s[2])
	b3, _ := strconv.Atoi(s[3])
	ipInt += (uint32)(b0 << 24)
	ipInt += (uint32)(b1 << 16)
	ipInt += (uint32)(b2 << 8)
	ipInt += (uint32)(b3 << 0)
	return
}

func IntToIp(ipInt uint32) (ipStr string) {
	b0 := byte(ipInt & 0xff000000 >> 24)
	b1 := byte(ipInt & 0xff0000 >> 16)
	b2 := byte(ipInt & 0xff00 >> 8)
	b3 := byte(ipInt & 0xff)
	return net.IPv4(b0, b1, b2, b3).String()
}
