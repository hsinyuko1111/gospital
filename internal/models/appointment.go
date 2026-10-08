package models

import "time"

// Appointment represents an appointment in the system
// @Description Appointment details including patient, doctor, date, and time information
type Appointment struct {
	AppointmentID   string    `json:"appointmentId" example:"APT-01JCG3X8P2ABCDEFGHIJKLMNOP"`
	PatientID       string    `json:"patientId" example:"PAT-01JCG3X8P2ABCDEFGHIJKLMNOP"`
	PatientName     string    `json:"patientName,omitempty" example:"John Doe"`
	DoctorID        string    `json:"doctorId" example:"DOC-01JCG3X8P2ABCDEFGHIJKLMNOP"`
	DoctorName      string    `json:"doctorName,omitempty" example:"Dr. Sarah Smith"`
	DepartmentID    string    `json:"departmentId,omitempty" example:"DEPT-001"`
	AppointmentDate string    `json:"appointmentDate" example:"2025-11-15"`
	StartTime       string    `json:"startTime" example:"10:00:00"`
	EndTime         string    `json:"endTime" example:"10:30:00"`
	Reason          string    `json:"reason" example:"Regular checkup"`
	Status          string    `json:"status" enums:"scheduled,completed,cancelled,no-show" example:"scheduled"`
	Notes           string    `json:"notes,omitempty" example:"Patient requested morning appointment"`
	CreatedAt       time.Time `json:"createdAt" example:"2025-11-09T10:00:00Z"`
	UpdatedAt       time.Time `json:"updatedAt" example:"2025-11-09T10:00:00Z"`
}

// AppointmentCreate represents the data needed to create a new appointment
// @Description Data required to book a new appointment
type AppointmentCreate struct {
	PatientID       string `json:"patientId" binding:"required" example:"PAT-01JCG3X8P2ABCDEFGHIJKLMNOP"`
	DoctorID        string `json:"doctorId" binding:"required" example:"DOC-01JCG3X8P2ABCDEFGHIJKLMNOP"`
	AppointmentDate string `json:"appointmentDate" binding:"required" example:"2025-11-15"`
	StartTime       string `json:"startTime" binding:"required" example:"10:00:00"`
	EndTime         string `json:"endTime" binding:"required" example:"10:30:00"`
	Reason          string `json:"reason" binding:"required" example:"Regular checkup"`
	Notes           string `json:"notes,omitempty" example:"Patient requested morning appointment"`
}

// AppointmentUpdate represents the data that can be updated for an appointment
// @Description Data for updating or rescheduling an existing appointment
type AppointmentUpdate struct {
	AppointmentDate string `json:"appointmentDate,omitempty" example:"2025-11-15"`
	StartTime       string `json:"startTime,omitempty" example:"10:00:00"`
	EndTime         string `json:"endTime,omitempty" example:"10:30:00"`
	Reason          string `json:"reason,omitempty" example:"Regular checkup"`
	Status          string `json:"status,omitempty" enums:"scheduled,completed,cancelled,no-show" example:"scheduled"`
	Notes           string `json:"notes,omitempty" example:"Patient requested morning appointment"`
}
