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
	Data  []Datum `json:"data"`
	Count int64   `json:"count"`
}

type Datum struct {
	IPPort      string  `json:"ipPort"`
	IP          string  `json:"ip"`
	Port        string  `json:"port"`
	Country     string  `json:"country"`
	LastChecked string  `json:"last_checked"`
	ProxyLevel  string  `json:"proxy_level"`
	Type        string  `json:"type"`
	Speed       string  `json:"speed"`
	Support     Support `json:"support"`
}

type Support struct {
	HTTPS     int64 `json:"https"`
	Get       int64 `json:"get"`
	Post      int64 `json:"post"`
	Cookies   int64 `json:"cookies"`
	Referer   int64 `json:"referer"`
	UserAgent int64 `json:"user_agent"`
	Google    int64 `json:"google"`
}

var limiter = rate.NewLimiter(rate.Every(time.Second), 1)

var ctx = context.Background()

func GetProxy() (string, error) {
	url := "http://pubproxy.com/api/proxy?user_agent=true&https=true"
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
	return resp.Data[0].IPPort, nil
}
