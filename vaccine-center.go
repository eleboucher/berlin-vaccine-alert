package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
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

// VMessage holds the information for the api response from impfstoff
type VMessage struct {
	Stats []StatElement `json:"stats"`
}

// StatElement holds the information for a specific vaccination center
type StatElement struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Open bool   `json:"open"`
	// Stats      map[string]StatValue `json:"stats"`
	LastUpdate *int64 `json:"lastUpdate,omitempty"`
}

// type StatValue struct {
// 	Percent float64 `json:"percent"`
// 	Count   int64   `json:"count"`
// 	Last    int64   `json:"last"`
// }

// VaccineCenter holds the information for fetching the information for the
// impfstoff.link/ website
type VaccineCenter struct {
	resultSendLastAt time.Time
	lastResult       []*Result
}

// Fetch fetches all the appointments and filter then and return the results
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

// ShouldSendResult check if the result should be send now
func (p *VaccineCenter) ShouldSendResult(result []*Result) bool {
	if !reflect.DeepEqual(p.lastResult, result) {
		return true
	}
	if p.resultSendLastAt.Before(time.Now().Add(-10 * time.Minute)) {
		return true
	}
	return false
}

// ResultSentNow set that the appointment has been sent
func (p *VaccineCenter) ResultSentNow(result []*Result) {
	p.resultSendLastAt = time.Now()
	p.lastResult = result
}

func idToURL(id string) string {
	return centerURL[id]
}
