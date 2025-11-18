package skit

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sync"

	"bufio"
	"io"
	"time"
	"unsafe"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
)

const (
	XForwardedFor = "X-Forwarded-For"
	XRealIP       = "X-Real-IP"
)

//返回本机IP地址串}
func LocalIP() string {
	addrslice, err := net.InterfaceAddrs()
	if err == nil {
		for _, addr := range addrslice {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}
	return "localhost"
}

// RemoteIp 返回远程客户端的 IP，如 192.168.1.1
func RemoteIp(req *http.Request) string {
	remoteAddr := req.RemoteAddr
	if ip := req.Header.Get(XRealIP); ip != "" {
		remoteAddr = ip
	} else if ip = req.Header.Get(XForwardedFor); ip != "" {
		remoteAddr = ip
	} else {
		remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
	}

	if remoteAddr == "::1" {
		remoteAddr = "127.0.0.1"
	}

	return remoteAddr
}

// Ip2long 将 IPv4 字符串形式转为 uint32
func Ip2long(ipstr string) uint32 {
	ip := net.ParseIP(ipstr)
	if ip == nil {
		return 0
	}
	ip = ip.To4()
	return binary.BigEndian.Uint32(ip)
}

type Client struct {
	HttpClient *http.Client
	Url        string
	Method     string
	Error      error
	Bs         []byte
	mu         sync.RWMutex
}

func NewHttpClient() (client *Client) {
	c := new(Client)
	c.HttpClient = &http.Client{}
	return c
}

func (c *Client) Get(url string) (client *Client) {
	c.mu.Lock()
	c.Method = http.MethodGet
	c.Url = url
	c.mu.Unlock()
	return c
}

func (c *Client) Post(url string) (client *Client) {
	c.mu.Lock()
	c.Method = http.MethodPost
	c.Url = url
	c.mu.Unlock()
	return c
}

func (c *Client) Send(bs []byte) (client *Client) {
	c.mu.Lock()
	c.Bs = bs
	c.mu.Unlock()
	return c
}

func (c *Client) End(v interface{}) (err error) {
	if c.Error != nil {
		return c.Error
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	req, err := http.NewRequest(c.Method, c.Url, bytes.NewReader(c.Bs))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("HTTP Request Error, StatusCode = %d", res.StatusCode)
	}
	defer res.Body.Close()
	bs, err := encoding(res) //ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	//fmt.Println(string(bs))
	err = json.Unmarshal([]byte(bs), &v)
	if err != nil {
		return err
	}
	return nil
}

func encoding(r *http.Response) (res string, err error) {
	// content-type 中会提供编码，比如 content-type="text/html;charset=utf-8"
	// html head meta 获取编码，
	// <meta http-equiv=Content-Type content="text/html;charset=utf-8"
	// 可以通过网页的头部猜测网页的编码信息。
	bufReader := bufio.NewReader(r.Body)
	bytes, _ := bufReader.Peek(1024) // 不会移动 reader 的读取位置

	e, _, _ := charset.DetermineEncoding(bytes, r.Header.Get("content-type"))

	bodyReader := transform.NewReader(bufReader, e.NewDecoder())
	content, err := ioutil.ReadAll(bodyReader)
	return *(*string)(unsafe.Pointer(&content)), err
}

/*
有关Http协议GET和POST请求的封装
*/

//发送GET请求
//url:请求地址
//response:请求返回的内容
func Get(url string) (response string) {
	client := http.Client{Timeout: 15 * time.Second}
	resp, error := client.Get(url)
	defer resp.Body.Close()
	if error != nil {
		panic(error)
	}

	var buffer [512]byte
	result := bytes.NewBuffer(nil)
	for {
		n, err := resp.Body.Read(buffer[0:])
		result.Write(buffer[0:n])
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
	}

	response = result.String()
	return
}

//发送POST请求
//url:请求地址，data:POST请求提交的数据,contentType:请求体格式，如：application/json
//content:请求放回的内容
func Post(url string, data interface{}, contentType string) (content string) {
	jsonStr, _ := json.Marshal(data)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Add("content-type", contentType)
	if err != nil {
		panic(err)
	}
	defer req.Body.Close()

	client := &http.Client{Timeout: 15 * time.Second}
	resp, error := client.Do(req)
	if error != nil {
		panic(error)
	}
	defer resp.Body.Close()

	result, _ := encoding(resp) //ioutil.ReadAll(resp.Body)
	content = string(result)
	return
}
