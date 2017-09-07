// wechat writer of seelog
// wechat notify use https://sc.ftqq.com
// seelog config as:
// <wechat formatid="wechat" baseurl="https://sc.ftqq.com">
//   <recipient sckey="SCU10919T101abe9bac1ae6e848dfb54ab7f9138059955f22090d5"/>
//   <recipient sckey="SCU10919T101abe9bac1ae6e848dfb54ab7f9978423889f8s94397"/>
// </wechat>

package seelog

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

var (
	ERR_WECHAT_RESPONSE = errors.New("ERR_WECHAT_RESPONSE")
)

type wechatWriter struct {
	baseUrl string
	scKeys  []string
	httpCli *http.Client
}

func NewWechatWriter(baseUrl string, scKeys []string) *wechatWriter {
	return &wechatWriter{
		baseUrl: baseUrl,
		scKeys:  scKeys,
		httpCli: newHttpClient(5, 10),
	}
}

func (wechat *wechatWriter) request(url string) error {
	var (
		resp *http.Response
		err  error
		buff []byte
	)
	if resp, err = wechat.httpCli.Get(url); err != nil {
		return err
	}
	defer resp.Body.Close()

	if buff, err = ioutil.ReadAll(resp.Body); err != nil {
		return err
	}

	//{"errno":0,"errmsg":"success","dataset":"done"}
	var resData map[string]interface{}
	if err := json.Unmarshal(buff, &resData); err != nil {
		log.Println(err, " resp:", string(buff))
		return err
	}
	errmsg := resData["errmsg"]
	if errmsg == nil {
		log.Println(ERR_WECHAT_RESPONSE, " resp:", string(buff))
		return ERR_WECHAT_RESPONSE
	}
	if errmsg.(string) != "success" {
		return errors.New(errmsg.(string))
	}
	return nil
}

func (wechat *wechatWriter) Write(data []byte) (int, error) {
	//	https://sc.ftqq.com/id.send?text=%s
	text := string(data)
	for _, sckey := range wechat.scKeys {
		url := wechat.baseUrl + "/" + sckey + ".send?text=" + text
		if err := wechat.request(url); err != nil {
			log.Println("wechat request err:", err, " url:", url)
		}
	}
	return len(data), nil
}

func newHttpClient(dialTimeout, deadlineTimeout time.Duration) *http.Client {
	c := http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(deadlineTimeout * time.Second)
				c, err := net.DialTimeout(netw, addr, time.Second*dialTimeout)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
			TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
			DisableCompression: true,
		},
	}
	return &c
}
