package main

// DoctolibConfig holds the information of the config.yml doctolib link declaration
type DoctolibConfig struct {
	URL           string  `mapstructure:"url"`
	VaccineName   string  `mapstructure:"vaccine_name"`
	PracticeID    string  `mapstructure:"practice_id"`
	AgendaID      string  `mapstructure:"agenda_id"`
	VisitMotiveID string  `mapstructure:"visit_motive_id"`
	Detail        *string `mapstructure:"detail"`
	Delay         *int    `mapstructure:"delay"`
}
