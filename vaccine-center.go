package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var centerURL = map[string]string{
	"arena":     "https://www.doctolib.de/institut/berlin/ciz-berlin-berlin?pid=practice-158431",
	"tempelhof": "https://www.doctolib.de/institut/berlin/ciz-berlin-berlin?pid=practice-158433",
	"messe":     "https://www.doctolib.de/institut/berlin/ciz-berlin-berlin?pid=practice-158434",
	"velodrom":  "https://www.doctolib.de/institut/berlin/ciz-berlin-berlin?pid=practice-158435",
	"tegel":     "https://www.doctolib.de/institut/berlin/ciz-berlin-berlin?pid=practice-158436",
	"erika":     "https://www.doctolib.de/institut/berlin/ciz-berlin-berlin?pid=practice-158437",
}

type VMessage struct {
	Stats []StatElement `json:"stats"`
}

type StatElement struct {
	ID         string               `json:"id"`
	Name       string               `json:"name"`
	Open       bool                 `json:"open"`
	Stats      map[string]StatValue `json:"stats"`
	LastUpdate *int64               `json:"lastUpdate,omitempty"`
}

type StatValue struct {
	Percent float64 `json:"percent"`
	Count   int64   `json:"count"`
	Last    int64   `json:"last"`
}

type VaccineCenter struct {
	resultSendLastAt time.Time
}

func (v *VaccineCenter) Fetch() ([]*Result, error) {
	url := "https://api.impfstoff.link/?v=0.3"

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("robot", "1")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var resp VMessage
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	var ret []*Result

	for _, a := range resp.Stats {
		if a.Open {
			message := fmt.Sprintf("Appointment available at %s %s", a.Name, idToURL(a.ID))
			ret = append(ret, &Result{
				Message:     message,
				VaccineName: VaccinationCenter,
			})
		}
	}
	return ret, nil
}

func (v *VaccineCenter) FormatMessage(result *Result) (string, error) {
	return "", nil
}

func (v *VaccineCenter) ResultSendLastAt() time.Time {
	return v.resultSendLastAt
}

func (v *VaccineCenter) ResultSentNow() {
	v.resultSendLastAt = time.Now()
}

func idToURL(id string) string {
	return centerURL[id]
}
