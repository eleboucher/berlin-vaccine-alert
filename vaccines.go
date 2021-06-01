package main

import (
	"errors"
	"strings"
)

const (
	AstraZeneca       = "astrazeneca"
	JohnsonAndJohnson = "johnson"
	Biontech          = "biontech"
	Pfizer            = "pfizer"
	Moderna           = "moderna"
	VaccinationCenter = "vaccination_center"
)

var ErrVaccineNotFound = errors.New("vaccine not found")

var vaccines = []string{
	AstraZeneca,
	JohnsonAndJohnson,
	Biontech,
	Pfizer,
	Moderna,
}

func getVaccineName(name string) (string, error) {
	for _, vaccine := range vaccines {
		if strings.Contains(strings.ToLower(name), vaccine) {
			return vaccine, nil
		}
	}
	return "", ErrVaccineNotFound
}
