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

const tPunto = "{{.Nr}} appointments for {{.Name}} available https://punctum-medico.de/onlinetermine/"

// PuntoMedico holds the information for fetching the information for the
// punctum-medico.de website
type PuntoMedico struct {
	resultSendLastAt time.Time
	lastResult       []*vaccines.Result
}

// TMessage holds the information for the api response from punto medico
type TMessage struct {
	Terminsuchen          []Terminsuchen `json:"terminsuchen"`
	Termine               [][]*string    `json:"termine"`
	TermineProBezeichnung [][][]*string  `json:"termineProBezeichnung"`
}

// Terminsuchen holds the information of an appointment
type Terminsuchen struct {
	Name string `json:"name"`
	Nr   int64  `json:"nr"`
}

// Fetch fetches all the available appointment and filter then and return the results
func (p *PuntoMedico) Fetch() ([]*vaccines.Result, error) {
	url := "https://onlinetermine.zollsoft.de/includes/searchTermine_app_feature.php"

	payload := strings.NewReader("versichert=1&terminsuche=&uniqueident=5a72efb4d3aec")

	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("cookie", "sec_session_id=f0807917851f007bb0af2f1f4815c445")
	req.Header.Add("authority", "onlinetermine.zollsoft.de")
	req.Header.Add("accept", "*/*")
	req.Header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36")
	req.Header.Add("content-type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("sec-gpc", "1")
	req.Header.Add("origin", "https://punctum-medico.de")
	req.Header.Add("sec-fetch-site", "cross-site")
	req.Header.Add("sec-fetch-mode", "cors")
	req.Header.Add("sec-fetch-dest", "empty")
	req.Header.Add("referer", "https://punctum-medico.de/")
	req.Header.Add("accept-language", "en-GB,en-US;q=0.9,en;q=0.8")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var resp TMessage
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	var ret []*vaccines.Result

	for _, a := range resp.Terminsuchen {
		// Remove second dose vaccine for now
		if !strings.Contains(a.Name, "Zweitimpfung") {
			if vaccineName, err := vaccines.GetVaccineName(a.Name); err == nil {
				message, err := p.formatMessage(a)
				if err != nil {
					return nil, err
				}
				ret = append(ret, &vaccines.Result{
					VaccineName: vaccineName,
					Amount:      a.Nr,
					Message:     message,
				})

			}
		}
	}
	return ret, nil
}

// ShouldSendResult check if the result should be send now
func (p *PuntoMedico) ShouldSendResult(result []*vaccines.Result) bool {
	if !reflect.DeepEqual(p.lastResult, result) && p.resultSendLastAt.Before(time.Now().Add(-1*time.Minute)) {
		return true
	}
	if p.resultSendLastAt.Before(time.Now().Add(-10 * time.Minute)) {
		return true
	}
	return false
}

// ResultSentNow set that the appointment has been sent
func (p *PuntoMedico) ResultSentNow(result []*vaccines.Result) {
	p.resultSendLastAt = time.Now()
	p.lastResult = result
}

func (p *PuntoMedico) formatMessage(result Terminsuchen) (string, error) {
	t, err := template.New("message").Parse(tPunto)
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
