package models

var (
	// PqUniqueViolationErrName is the name of the error code returned
	// by pq when a given record violates a unique constraint on a table
	PqUniqueViolationErrName = "unique_violation"
)
