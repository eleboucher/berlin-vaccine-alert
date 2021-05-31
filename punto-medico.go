package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"text/template"
	"time"
)

const tPunto = "{{.Amount}} appointments for {{.VaccineName}} available https://punctum-medico.de/onlinetermine/"

var vaccines = []string{
	"AstraZeneca",
	"Johnson",
	"Biontech",
}

type PuntoMedico struct {
	resultSendLastAt time.Time
}

type TMessage struct {
	Terminsuchen          []Terminsuchen `json:"terminsuchen"`
	Termine               [][]*string    `json:"termine"`
	TermineProBezeichnung [][][]*string  `json:"termineProBezeichnung"`
}

type Terminsuchen struct {
	Name string `json:"name"`
	Nr   int64  `json:"nr"`
}

func (p *PuntoMedico) Fetch() ([]*Result, error) {
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

	var ret []*Result

	for _, a := range resp.Terminsuchen {
		// Remove second dose vaccine for now
		if !strings.Contains(a.Name, "Zweitimpfung") && isVaccine(a.Name) {
			ret = append(ret, &Result{
				VaccineName: a.Name,
				Amount:      a.Nr,
			})
		}
	}
	return ret, nil
}

func isVaccine(name string) bool {
	for _, vaccine := range vaccines {
		if strings.Contains(name, vaccine) {
			return true
		}
	}
	return false
}

func (p *PuntoMedico) ResultSendLastAt() time.Time {
	return p.resultSendLastAt
}

func (p *PuntoMedico) ResultSentNow() {
	p.resultSendLastAt = time.Now()
}

func (p *PuntoMedico) FormatMessage(result *Result) (string, error) {
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
