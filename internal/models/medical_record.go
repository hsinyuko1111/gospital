package models

import "time"

// Prescription represents a medication prescription
// @Description Medication prescription details including dosage and duration
type Prescription struct {
	Medication string `json:"medication" dynamodbav:"medication" example:"Nitroglycerin"`
	Dosage     string `json:"dosage" dynamodbav:"dosage" example:"0.4mg"`
	Frequency  string `json:"frequency" dynamodbav:"frequency" example:"As needed"`
	Duration   string `json:"duration" dynamodbav:"duration" example:"30 days"`
}

// LabResult represents a laboratory test result
// @Description Laboratory test result with normal range comparison
type LabResult struct {
	TestName    string `json:"testName" dynamodbav:"testName" example:"Total Cholesterol"`
	Result      string `json:"result" dynamodbav:"result" example:"240 mg/dL"`
	NormalRange string `json:"normalRange" dynamodbav:"normalRange" example:"< 200 mg/dL"`
}

// VitalSigns represents patient vital signs
// @Description Patient vital signs taken during visit
type VitalSigns struct {
	BloodPressure string  `json:"bloodPressure,omitempty" dynamodbav:"bloodPressure,omitempty" example:"120/80"`
	HeartRate     int     `json:"heartRate,omitempty" dynamodbav:"heartRate,omitempty" example:"72"`
	Temperature   float64 `json:"temperature,omitempty" dynamodbav:"temperature,omitempty" example:"98.6"`
	Weight        float64 `json:"weight,omitempty" dynamodbav:"weight,omitempty" example:"175.5"`
	Height        float64 `json:"height,omitempty" dynamodbav:"height,omitempty" example:"70"`
}

// MedicalRecord represents a patient's medical record
// @Description Complete medical record including diagnosis, treatment, prescriptions, and lab results
type MedicalRecord struct {
	RecordID       string         `json:"recordId" dynamodbav:"recordId" example:"REC-01JCG3X8P2ABCDEFGHIJKLMNOP"`
	PatientID      string         `json:"patientId" dynamodbav:"patientId" example:"PAT-01JCG3X8P2ABCDEFGHIJKLMNOP"`
	DoctorID       string         `json:"doctorId" dynamodbav:"doctorId" example:"DOC-01JCG3X8P2ABCDEFGHIJKLMNOP"`
	DoctorName     string         `json:"doctorName,omitempty" dynamodbav:"doctorName,omitempty" example:"Dr. Sarah Smith"`
	AppointmentID  string         `json:"appointmentId,omitempty" dynamodbav:"appointmentId,omitempty" example:"APT-01JCG3X8P2ABCDEFGHIJKLMNOP"`
	VisitDate      time.Time      `json:"visitDate" dynamodbav:"visitDate" example:"2025-11-08T10:00:00Z"`
	ChiefComplaint string         `json:"chiefComplaint" dynamodbav:"chiefComplaint" example:"Chest pain"`
	Diagnosis      string         `json:"diagnosis,omitempty" dynamodbav:"diagnosis,omitempty" example:"Angina pectoris"`
	Symptoms       []string       `json:"symptoms,omitempty" dynamodbav:"symptoms,omitempty" example:"chest pain,shortness of breath"`
	Treatment      string         `json:"treatment,omitempty" dynamodbav:"treatment,omitempty" example:"Prescribed nitroglycerin and scheduled follow-up"`
	Prescriptions  []Prescription `json:"prescriptions,omitempty" dynamodbav:"prescriptions,omitempty"`
	LabResults     []LabResult    `json:"labResults,omitempty" dynamodbav:"labResults,omitempty"`
	VitalSigns     *VitalSigns    `json:"vitalSigns,omitempty" dynamodbav:"vitalSigns,omitempty"`
	Notes          string         `json:"notes,omitempty" dynamodbav:"notes,omitempty" example:"Patient reports chest discomfort during physical activity"`
	FollowUpDate   string         `json:"followUpDate,omitempty" dynamodbav:"followUpDate,omitempty" example:"2025-11-23"`
	CreatedAt      time.Time      `json:"createdAt" dynamodbav:"createdAt" example:"2025-11-09T10:00:00Z"`
	UpdatedAt      time.Time      `json:"updatedAt" dynamodbav:"updatedAt" example:"2025-11-09T10:00:00Z"`
}

// MedicalRecordCreate represents the data needed to create a new medical record
// @Description Data required to create a new medical record for a patient visit
type MedicalRecordCreate struct {
	PatientID      string         `json:"patientId" binding:"required" example:"PAT-01JCG3X8P2ABCDEFGHIJKLMNOP"`
	DoctorID       string         `json:"doctorId" binding:"required" example:"DOC-01JCG3X8P2ABCDEFGHIJKLMNOP"`
	AppointmentID  string         `json:"appointmentId,omitempty" example:"APT-01JCG3X8P2ABCDEFGHIJKLMNOP"`
	VisitDate      time.Time      `json:"visitDate" binding:"required" example:"2025-11-09T10:00:00Z"`
	ChiefComplaint string         `json:"chiefComplaint" binding:"required" example:"Chest pain"`
	Diagnosis      string         `json:"diagnosis,omitempty" example:"Angina pectoris"`
	Symptoms       []string       `json:"symptoms,omitempty" example:"chest pain,shortness of breath"`
	Treatment      string         `json:"treatment,omitempty" example:"Prescribed nitroglycerin and scheduled follow-up"`
	Prescriptions  []Prescription `json:"prescriptions,omitempty"`
	LabResults     []LabResult    `json:"labResults,omitempty"`
	VitalSigns     *VitalSigns    `json:"vitalSigns,omitempty"`
	Notes          string         `json:"notes,omitempty" example:"Patient reports chest discomfort during physical activity"`
	FollowUpDate   string         `json:"followUpDate,omitempty" example:"2025-11-23"`
}

// MedicalRecordUpdate represents the data that can be updated for a medical record
// @Description Data for updating an existing medical record
type MedicalRecordUpdate struct {
	Diagnosis     string         `json:"diagnosis,omitempty" example:"Angina pectoris"`
	Symptoms      []string       `json:"symptoms,omitempty" example:"chest pain,shortness of breath"`
	Treatment     string         `json:"treatment,omitempty" example:"Prescribed nitroglycerin and scheduled follow-up"`
	Prescriptions []Prescription `json:"prescriptions,omitempty"`
	LabResults    []LabResult    `json:"labResults,omitempty"`
	Notes         string         `json:"notes,omitempty" example:"Patient reports chest discomfort during physical activity"`
	FollowUpDate  string         `json:"followUpDate,omitempty" example:"2025-11-23"`
}
