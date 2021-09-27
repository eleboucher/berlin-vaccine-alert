package sources

import (
	"bytes"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"text/template"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/eleboucher/berlin-vaccine-alert/vaccines"
)

const tMedicoLeopoldPlatz = "{{.Amount}} appointments for {{.Name}} available call 0304579790"

var regexMedicoLeopoldPlatz = regexp.MustCompile(`Impftermine COVID-19 mit (\w+): (\d+)`)

type resultMedicoLeopoldPlatz struct {
	Amount int
	Name   string
}

// MedicoLeopoldPlatz holds the information for fetching the information for the
// https://medico-leopoldplatz.de/ website
type MedicoLeopoldPlatz struct {
	resultSendLastAt time.Time
	lastResult       []*vaccines.Result
}

// Name return the name of the source
func (m *MedicoLeopoldPlatz) Name() string {
	return "Medico LeopoldPlatz"
}

// Fetch fetches all the available appointment and filter then and return the results
func (m *MedicoLeopoldPlatz) Fetch() ([]*vaccines.Result, error) {
	var ret []*vaccines.Result
	res, err := http.Get("https://medico-leopoldplatz.de/corona-covid-19-impfung/")
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	doc.Find(".elementor-element-290b411 p").
		Each(
			func(i int, s *goquery.Selection) {
				// For each item found, get the band and title
				vaccine := s.Text()
				matches := regexMedicoLeopoldPlatz.FindAllStringSubmatch(vaccine, -1)
				if len(matches) == 0 {
					return
				}
				name := matches[0][1]
				if vaccineName, err := vaccines.GetVaccineName(name); err == nil {
					amount, err := strconv.Atoi(matches[0][2])
					if err != nil {
						return
					}
					message, err := m.formatMessage(
						resultMedicoLeopoldPlatz{
							Name:   name,
							Amount: amount,
						},
					)
					if err != nil {
						return
					}
					ret = append(ret, &vaccines.Result{
						VaccineName: vaccineName,
						Amount:      int64(amount),
						Message:     message,
					})

				}

			},
		)
	return ret, nil
}

func (m *MedicoLeopoldPlatz) formatMessage(result resultMedicoLeopoldPlatz) (string, error) {
	t, err := template.New("message").Parse(tMedicoLeopoldPlatz)
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
func (m *MedicoLeopoldPlatz) ShouldSendResult(result []*vaccines.Result) bool {
	if !reflect.DeepEqual(m.lastResult, result) && m.resultSendLastAt.Before(time.Now().Add(-1*time.Minute)) {
		return true
	}
	if m.resultSendLastAt.Before(time.Now().Add(-10 * time.Minute)) {
		return true
	}
	return false
}

// ResultSentNow set that the appointment has been sent
func (m *MedicoLeopoldPlatz) ResultSentNow(result []*vaccines.Result) {
	m.resultSendLastAt = time.Now()
	m.lastResult = result
}
