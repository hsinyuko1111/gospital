package services

import "errors"

// Common service errors
var (
	// Patient errors
	ErrPatientNotFound      = errors.New("patient not found")
	ErrPatientAlreadyExists = errors.New("patient already exists")
	ErrInvalidPatientData   = errors.New("invalid patient data")

	// Doctor errors
	ErrDoctorNotFound      = errors.New("doctor not found")
	ErrDoctorAlreadyExists = errors.New("doctor already exists")
	ErrInvalidDoctorData   = errors.New("invalid doctor data")

	// Appointment errors
	ErrAppointmentNotFound    = errors.New("appointment not found")
	ErrAppointmentSlotTaken   = errors.New("appointment slot not available")
	ErrInvalidAppointmentData = errors.New("invalid appointment data")

	// Department errors
	ErrDepartmentNotFound = errors.New("department not found")

	// Medical Record errors
	ErrMedicalRecordNotFound    = errors.New("medical record not found")
	ErrInvalidMedicalRecordData = errors.New("invalid medical record data")

	// Invoice errors
	ErrInvoiceNotFound    = errors.New("invoice not found")
	ErrInvalidInvoiceData = errors.New("invalid invoice data")
	ErrInvalidPaymentData = errors.New("invalid payment data")

	// General errors
	ErrInvalidInput = errors.New("invalid input")
	ErrInternal     = errors.New("internal server error")
)
