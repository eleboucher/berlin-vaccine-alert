package sources

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"text/template"
	"time"

	"github.com/eleboucher/covid/vaccines"
)

// ArkonoPlatzPfizer holds the information for fetching the information for the
// https://medico-leopoldplatz.de/ website
type ArkonoPlatzPfizer struct {
	resultSendLastAt time.Time
	lastResult       []*vaccines.Result
}

type APfizerResponse struct {
	Data             []string      `json:"Data"`
	Success          bool          `json:"Success"`
	Error            interface{}   `json:"Error"`
	ValidationErrors []interface{} `json:"ValidationErrors"`
}

type APfizerRequest struct {
	CalendarID  int64  `json:"calendarId"`
	ServiceID   int64  `json:"serviceId"`
	PersonCount int64  `json:"personCount"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
}

// Name return the name of the source
func (a *ArkonoPlatzPfizer) Name() string {
	return "ArkonoPlatz"
}

// Fetch fetches all the available appointment and filter then and return the results
func (a *ArkonoPlatzPfizer) Fetch() ([]*vaccines.Result, error) {
	url := "https://praxis-arkonaplatz.termin-direkt.de/rest-v2/api/Calendars/4/DaysWithFreeIntervals"

	reqPayload := ARequest{
		CalendarID:  4,
		ServiceID:   5,
		PersonCount: 1,
		StartDate:   time.Now().Format(time.RFC3339Nano),
		EndDate:     time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339Nano),
	}
	payload, err := json.Marshal(&reqPayload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, strings.NewReader(string(payload)))
	if err != nil {
		return nil, err
	}
	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var resp AResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}

	var ret vaccines.Result
	ret.VaccineName = vaccines.AstraZeneca
	message, err := a.formatMessage(ret)
	if err != nil {
		return nil, err
	}
	ret.Message = message

	return []*vaccines.Result{&ret}, nil
}

func (a *ArkonoPlatzPfizer) formatMessage(result vaccines.Result) (string, error) {
	t, err := template.New("message").Parse(tArkonoPlatz)
	if err != nil {
		return "", err
	}
	var tpl bytes.Buffer
	err = t.Execute(&tpl, result)
	if err != nil {
		return "", err
	}
	return tpl.String(), nil
}

// ShouldSendResult check if the result should be send now
func (a *ArkonoPlatzPfizer) ShouldSendResult(result []*vaccines.Result) bool {
	if !reflect.DeepEqual(a.lastResult, result) && a.resultSendLastAt.Before(time.Now().Add(-1*time.Minute)) {
		return true
	}
	if a.resultSendLastAt.Before(time.Now().Add(-10 * time.Minute)) {
		return true
	}
	return false
}

// ResultSentNow set that the appointment has been sent
func (a *ArkonoPlatzPfizer) ResultSentNow(result []*vaccines.Result) {
	a.resultSendLastAt = time.Now()
	a.lastResult = result
}
