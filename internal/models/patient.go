package models

import "time"

// Address represents a physical address
type Address struct {
	Street  string `json:"street,omitempty" example:"123 Main St"`
	City    string `json:"city,omitempty" example:"San Francisco"`
	State   string `json:"state,omitempty" example:"CA"`
	ZipCode string `json:"zipCode,omitempty" example:"94102"`
	Country string `json:"country,omitempty" example:"USA"`
}

// EmergencyContact represents emergency contact information
type EmergencyContact struct {
	Name         string `json:"name,omitempty" example:"Jane Doe"`
	Relationship string `json:"relationship,omitempty" example:"Spouse"`
	Phone        string `json:"phone,omitempty" example:"+1-555-987-6543"`
}

// Patient represents a patient in the system
// @Description Patient information including personal details, contact information, and medical data
type Patient struct {
	PatientID        string            `json:"patientId" example:"PAT-01JCG3X8P2ABCDEFGHIJKLMNOP"`
	FirstName        string            `json:"firstName" example:"John"`
	LastName         string            `json:"lastName" example:"Doe"`
	DateOfBirth      string            `json:"dateOfBirth" example:"1985-03-15"`
	Gender           string            `json:"gender,omitempty" enums:"male,female,other" example:"male"`
	Email            string            `json:"email" example:"john.doe@email.com"`
	Phone            string            `json:"phone" example:"+1-555-123-4567"`
	Address          *Address          `json:"address,omitempty"`
	EmergencyContact *EmergencyContact `json:"emergencyContact,omitempty"`
	BloodType        string            `json:"bloodType,omitempty" example:"O+"`
	Allergies        []string          `json:"allergies,omitempty" example:"Penicillin,Peanuts"`
	Status           string            `json:"status" enums:"active,inactive" example:"active"`
	CreatedAt        time.Time         `json:"createdAt" example:"2025-11-09T10:00:00Z"`
	UpdatedAt        time.Time         `json:"updatedAt" example:"2025-11-09T10:00:00Z"`
}

// PatientCreate represents the data needed to create a new patient
// @Description Data required to register a new patient in the system
type PatientCreate struct {
	FirstName        string            `json:"firstName" binding:"required" example:"John"`
	LastName         string            `json:"lastName" binding:"required" example:"Doe"`
	DateOfBirth      string            `json:"dateOfBirth" binding:"required" example:"1985-03-15"`
	Gender           string            `json:"gender,omitempty" enums:"male,female,other" example:"male"`
	Email            string            `json:"email" binding:"required,email" example:"john.doe@email.com"`
	Phone            string            `json:"phone" binding:"required" example:"+1-555-123-4567"`
	Address          *Address          `json:"address,omitempty"`
	EmergencyContact *EmergencyContact `json:"emergencyContact,omitempty"`
	BloodType        string            `json:"bloodType,omitempty" example:"O+"`
	Allergies        []string          `json:"allergies,omitempty" example:"Penicillin,Peanuts"`
}

// PatientUpdate represents the data that can be updated for a patient
// @Description Data for updating an existing patient's information
type PatientUpdate struct {
	FirstName        string            `json:"firstName,omitempty" example:"John"`
	LastName         string            `json:"lastName,omitempty" example:"Doe"`
	Email            string            `json:"email,omitempty" example:"john.doe@email.com"`
	Phone            string            `json:"phone,omitempty" example:"+1-555-123-4567"`
	Address          *Address          `json:"address,omitempty"`
	EmergencyContact *EmergencyContact `json:"emergencyContact,omitempty"`
	Allergies        []string          `json:"allergies,omitempty" example:"Penicillin,Peanuts"`
}
