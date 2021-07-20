package sources

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"
	"text/template"
	"time"

	"github.com/eleboucher/covid/internals/proxy"
	"github.com/eleboucher/covid/vaccines"
)

const tDoctolib = "{{.Amount}} appointments for {{.VaccineName}} {{.Detail}} available {{.URL}}"

type ResultDoctolib struct {
	Availabilities []*Availability `json:"availabilities,omitempty"`
	Total          int64           `json:"total"`
	NextSlot       *string         `json:"next_slot"`
}

type Availability struct {
	Date  string        `json:"date"`
	Slots []interface{} `json:"slots"`
}

// Doctolib holds the information for fetching the information for the
// doctolib website
type Doctolib struct {
	VaccineName      string             `json:"-"`
	URL              string             `json:"-"`
	Detail           string             `json:"-"`
	Delay            time.Duration      `json:"-"`
	Limit            string             `json:"limit"`
	PracticeID       string             `json:"pratice_ids"`
	AgendaID         string             `json:"agenda_ids"`
	VisitMotiveID    string             `json:"visit_motive_ids"`
	StartDate        string             `json:"start_date"`
	resultSendLastAt time.Time          `json:"-"`
	lastResult       []*vaccines.Result `json:"-"`
}

// Name return the name of the source
func (d *Doctolib) Name() string {
	return "Doctolib " + d.URL
}

// Fetch fetches all the available appointment and filter then and return the results
func (d *Doctolib) Fetch() ([]*vaccines.Result, error) {
	url := "http://doctolib-proxy.herokuapp.com/doctolib"
	var ret vaccines.Result
	startDate := time.Now()
	for {
		prx, err := proxy.GetProxy()
		if err != nil {
			return nil, err
		}
		os.Setenv("HTTP_PROXY", "http://"+prx)
		d.StartDate = startDate.Format("2006-01-02")
		d.Limit = "1000"
		payload, err := json.Marshal(d)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequest("POST", url, strings.NewReader(string(payload)))

		if err != nil {
			return nil, err
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)

		if err != nil {
			return nil, err
		}
		var resp ResultDoctolib
		err = json.Unmarshal(body, &resp)
		if err != nil {
			return nil, err
		}

		ret.VaccineName = d.VaccineName

		for _, availability := range resp.Availabilities {
			date, err := time.Parse("2006-01-02", availability.Date)
			if err != nil {
				continue
			}
			// do not show vaccine that is in more than 30 days for now
			if date.After(time.Now().Add(30 * 24 * time.Hour)) {
				continue
			}
			ret.Amount += int64(len(availability.Slots))
		}
		if resp.NextSlot != nil {
			date, err := time.Parse("2006-01-02", *resp.NextSlot)
			if err != nil {
				break
			}
			// do not show vaccine that is in more than 30 days for now
			if date.After(time.Now().Add(30 * 24 * time.Hour)) {
				break
			}
			startDate = date
		} else {
			break
		}
	}
	if ret.Amount == 0 {
		return nil, nil
	}
	message, err := d.formatMessage(ret)
	if err != nil {
		return nil, err
	}
	ret.Message = message
	return []*vaccines.Result{&ret}, nil
}

func (d *Doctolib) formatMessage(result vaccines.Result) (string, error) {
	res := struct {
		URL         string
		VaccineName string
		Amount      int64
		Detail      string
	}{
		URL:         d.URL,
		VaccineName: d.VaccineName,
		Detail:      d.Detail,
		Amount:      result.Amount,
	}

	t, err := template.New("message").Parse(tDoctolib)
	if err != nil {
		return "", err
	}
	var tpl bytes.Buffer
	err = t.Execute(&tpl, res)
	if err != nil {
		return "", err
	}
	return tpl.String(), nil
}

// ShouldSendResult check if the result should be send now
func (d *Doctolib) ShouldSendResult(result []*vaccines.Result) bool {
	if d.Delay == 0 {
		d.Delay = 1
	}
	if !reflect.DeepEqual(d.lastResult, result) && d.resultSendLastAt.Before(time.Now().Add(-d.Delay*1*time.Minute)) {
		return true
	}
	if d.resultSendLastAt.Before(time.Now().Add(d.Delay - 10*time.Minute)) {
		return true
	}
	return false
}

// ResultSentNow set that the appointment has been sent
func (d *Doctolib) ResultSentNow(result []*vaccines.Result) {
	d.resultSendLastAt = time.Now()
	d.lastResult = result
}
