package proxy

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

type Response struct {
	SupportsHTTPS      bool           `json:"supportsHttps"`
	Protocol           string         `json:"protocol"`
	IP                 string         `json:"ip"`
	Port               string         `json:"port"`
	Get                bool           `json:"get"`
	Post               bool           `json:"post"`
	Cookies            bool           `json:"cookies"`
	Referer            bool           `json:"referer"`
	UserAgent          bool           `json:"user-agent"`
	AnonymityLevel     int64          `json:"anonymityLevel"`
	Websites           Websites       `json:"websites"`
	Country            interface{}    `json:"country"`
	UnixTimestampMS    int64          `json:"unixTimestampMs"`
	TsChecked          int64          `json:"tsChecked"`
	UnixTimestamp      int64          `json:"unixTimestamp"`
	Curl               string         `json:"curl"`
	IPPort             string         `json:"ipPort"`
	Type               string         `json:"type"`
	Speed              float64        `json:"speed"`
	OtherProtocols     OtherProtocols `json:"otherProtocols"`
	VerifiedSecondsAgo int64          `json:"verifiedSecondsAgo"`
}

type OtherProtocols struct {
}

type Websites struct {
	Example    bool `json:"example"`
	Google     bool `json:"google"`
	Amazon     bool `json:"amazon"`
	Yelp       bool `json:"yelp"`
	GoogleMaps bool `json:"google_maps"`
}

var limiter = rate.NewLimiter(rate.Every(2*time.Second), 1)

var ctx = context.Background()

type Proxy struct {
	IPPort string
}

func (p *Proxy) Proxy() string {
	if p.IPPort == "" {
		ipPort, err := fetchProxy()
		if err != nil {
			logrus.Error(err)
		}
		p.IPPort = ipPort
	}
	return p.IPPort
}

func (p *Proxy) RenewProxy() {
	ipPort, err := fetchProxy()
	if err != nil {
		logrus.Error(err)
	}
	p.IPPort = ipPort
}

func fetchProxy() (string, error) {
	url := "https://gimmeproxy.com/api/getProxy?user-agent=true&supportsHttps=true"
	err := limiter.Wait(ctx)
	if err != nil {
		return "", err
	}
	req, _ := http.NewRequest("GET", url, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {

		return "", err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {

		return "", err
	}
	logrus.Error(string(body))
	var resp Response
	err = json.Unmarshal(body, &resp)
	if err != nil {

		return "", err
	}
	return resp.Curl, nil
}
