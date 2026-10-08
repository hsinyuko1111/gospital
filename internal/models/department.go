package models

// Department represents a hospital department
// @Description Hospital department information including head doctor and staff count
type Department struct {
	DepartmentID    string `json:"departmentId" example:"DEPT-001"`
	Name            string `json:"name" example:"Cardiology"`
	Description     string `json:"description,omitempty" example:"Department specializing in heart and cardiovascular conditions"`
	Floor           int    `json:"floor,omitempty" example:"3"`
	Extension       string `json:"extension,omitempty" example:"3001"`
	HeadDoctorID    string `json:"headDoctorId,omitempty" example:"DOC-01JCG3X8P2ABCDEFGHIJKLMNOP"`
	HeadDoctorName  string `json:"headDoctorName,omitempty" example:"Dr. Sarah Smith"`
	NumberOfDoctors int    `json:"numberOfDoctors,omitempty" example:"8"`
}
