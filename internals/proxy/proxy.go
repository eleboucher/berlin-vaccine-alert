package proxy

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
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

func GetProxy() (string, error) {

	url := "http://pubproxy.com/api/proxy?user_agent=true"

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
	var resp Response
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return "", err
	}
	return resp.Data[0].IPPort, nil
}
