package vaccines

import (
	"errors"
	"strings"
)

const (
	// AstraZeneca is the name of the astrazeneca vaccine
	AstraZeneca = "astra"
	// JohnsonAndJohnson is the name for the JohnsonAndJohnson vaccine
	JohnsonAndJohnson = "johnson"
	// Pfizer is the name for the biontech/pfizer vaccine
	Pfizer = "pfizer"
	// Biontech is the other name for the biontech/pfizer vaccine
	// USE ONLY PFIZER
	biontech = "biontech"
	// Moderna is the name for the Moderna vaccine
	Moderna = "moderna"
	// VaccinationCenter correspond to the vaccination center
	VaccinationCenter = "vaccination_center"
)

// Result holds the information for a vaccine appointment
type Result struct {
	VaccineName string
	Amount      int64
	Message     string
}

// ErrVaccineNotFound is return when the vaccine can't be found
var ErrVaccineNotFound = errors.New("vaccine not found")

var vaccines = []string{
	AstraZeneca,
	JohnsonAndJohnson,
	Pfizer,
	Moderna,
}

func GetVaccineName(name string) (string, error) {

	for _, vaccine := range vaccines {
		if strings.Contains(strings.ToLower(name), vaccine) || (vaccine == Pfizer &&
			strings.Contains(strings.ToLower(name), biontech)) {
			return vaccine, nil
		}
	}
	return "", ErrVaccineNotFound
}
