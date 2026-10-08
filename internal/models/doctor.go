package models

import "time"

// Doctor represents a doctor in the system
// @Description Doctor information including specialty, department, and qualifications
type Doctor struct {
	DoctorID          string    `json:"doctorId" example:"DOC-01JCG3X8P2ABCDEFGHIJKLMNOP"`
	FirstName         string    `json:"firstName" example:"Sarah"`
	LastName          string    `json:"lastName" example:"Smith"`
	Email             string    `json:"email" example:"dr.smith@hospital.com"`
	Phone             string    `json:"phone" example:"+1-555-987-6543"`
	Specialty         string    `json:"specialty" example:"Cardiology"`
	LicenseNumber     string    `json:"licenseNumber" example:"MD123456"`
	DepartmentID      string    `json:"departmentId" example:"DEPT-001"`
	DepartmentName    string    `json:"departmentName,omitempty" example:"Cardiology Department"`
	YearsOfExperience int       `json:"yearsOfExperience,omitempty" example:"12"`
	Qualifications    []string  `json:"qualifications,omitempty" example:"MD,FACC,Board Certified"`
	Status            string    `json:"status" enums:"active,on-leave,inactive" example:"active"`
	CreatedAt         time.Time `json:"createdAt" example:"2025-11-09T10:00:00Z"`
	UpdatedAt         time.Time `json:"updatedAt" example:"2025-11-09T10:00:00Z"`
}

// DoctorCreate represents the data needed to create a new doctor
// @Description Data required to register a new doctor in the system
type DoctorCreate struct {
	FirstName         string   `json:"firstName" binding:"required" example:"Sarah"`
	LastName          string   `json:"lastName" binding:"required" example:"Smith"`
	Email             string   `json:"email" binding:"required,email" example:"dr.smith@hospital.com"`
	Phone             string   `json:"phone" binding:"required" example:"+1-555-987-6543"`
	Specialty         string   `json:"specialty" binding:"required" example:"Cardiology"`
	LicenseNumber     string   `json:"licenseNumber" binding:"required" example:"MD123456"`
	DepartmentID      string   `json:"departmentId" binding:"required" example:"DEPT-001"`
	YearsOfExperience int      `json:"yearsOfExperience,omitempty" example:"12"`
	Qualifications    []string `json:"qualifications,omitempty" example:"MD,FACC,Board Certified"`
}

// DoctorUpdate represents the data that can be updated for a doctor
// @Description Data for updating an existing doctor's information
type DoctorUpdate struct {
	Email             string   `json:"email,omitempty" example:"dr.smith@hospital.com"`
	Phone             string   `json:"phone,omitempty" example:"+1-555-987-6543"`
	Specialty         string   `json:"specialty,omitempty" example:"Cardiology"`
	DepartmentID      string   `json:"departmentId,omitempty" example:"DEPT-001"`
	YearsOfExperience int      `json:"yearsOfExperience,omitempty" example:"12"`
	Qualifications    []string `json:"qualifications,omitempty" example:"MD,FACC,Board Certified"`
	Status            string   `json:"status,omitempty" enums:"active,on-leave,inactive" example:"active"`
}
