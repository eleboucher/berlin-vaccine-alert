package main

import (
	"errors"
	"strings"
)

const (
	// AstraZeneca is the name of the astrazeneca vaccine
	AstraZeneca = "astra"
	// JohnsonAndJohnson is the name for the JohnsonAndJohnson vaccine
	JohnsonAndJohnson = "johnson"
	// Biontech is the name for the biontech/pfizer vaccine
	Biontech = "biontech"
	// Pfizer is the other name for the biontech/pfizer vaccine
	Pfizer = "pfizer"
	// Moderna is the name for the Moderna vaccine
	Moderna = "moderna"
	// VaccinationCenter correspond to the vaccination center
	VaccinationCenter = "vaccination_center"
)

// ErrVaccineNotFound is return when the vaccine can't be found
var ErrVaccineNotFound = errors.New("vaccine not found")

var vaccines = []string{
	AstraZeneca,
	JohnsonAndJohnson,
	Pfizer,
	Moderna,
}

func getVaccineName(name string) (string, error) {

	for _, vaccine := range vaccines {
		if strings.Contains(strings.ToLower(name), vaccine) || (vaccine == Pfizer &&
			strings.Contains(strings.ToLower(name), Biontech)) {
			return vaccine, nil
		}
	}
	return "", ErrVaccineNotFound
}
