package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/hsinyuko1111/gospital/internal/models"
	"github.com/hsinyuko1111/gospital/internal/repository"
)

// MedicalRecordService handles medical record business logic
type MedicalRecordService struct {
	medicalRecordRepo *repository.MedicalRecordRepository
	patientRepo       *repository.PatientRepository
	doctorRepo        *repository.DoctorRepository
}

// NewMedicalRecordService creates a new medical record service
func NewMedicalRecordService(
	medicalRecordRepo *repository.MedicalRecordRepository,
	patientRepo *repository.PatientRepository,
	doctorRepo *repository.DoctorRepository,
) *MedicalRecordService {
	return &MedicalRecordService{
		medicalRecordRepo: medicalRecordRepo,
		patientRepo:       patientRepo,
		doctorRepo:        doctorRepo,
	}
}

// CreateMedicalRecord creates a new medical record
func (s *MedicalRecordService) CreateMedicalRecord(ctx context.Context, record *models.MedicalRecordCreate) (*models.MedicalRecord, error) {
	// Validate patient exists (MySQL)
	_, err := s.patientRepo.GetByID(ctx, record.PatientID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrPatientNotFound
		}
		return nil, err
	}

	// Validate doctor exists and fetch doctor name (MySQL)
	doctor, err := s.doctorRepo.GetByID(ctx, record.DoctorID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrDoctorNotFound
		}
		return nil, err
	}

	// Create medical record in DynamoDB
	// Note: DoctorName will be populated by the repository using the doctor info we fetched
	createdRecord, err := s.medicalRecordRepo.Create(ctx, record)
	if err != nil {
		return nil, err
	}

	// Populate doctor name in the response (denormalized data for DynamoDB)
	createdRecord.DoctorName = fmt.Sprintf("Dr. %s %s", doctor.FirstName, doctor.LastName)

	return createdRecord, nil
}

// GetMedicalRecordByID retrieves a medical record by ID
func (s *MedicalRecordService) GetMedicalRecordByID(ctx context.Context, recordID string) (*models.MedicalRecord, error) {
	record, err := s.medicalRecordRepo.GetByID(ctx, recordID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrMedicalRecordNotFound
		}
		return nil, err
	}
	return record, nil
}

// GetMedicalRecordsByPatientID retrieves all medical records for a patient
func (s *MedicalRecordService) GetMedicalRecordsByPatientID(ctx context.Context, patientID string, startDate, endDate *string) ([]models.MedicalRecord, error) {
	// Validate patient exists (MySQL)
	_, err := s.patientRepo.GetByID(ctx, patientID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrPatientNotFound
		}
		return nil, err
	}

	// Get medical records from DynamoDB
	records, err := s.medicalRecordRepo.GetByPatientID(ctx, patientID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// UpdateMedicalRecord updates an existing medical record
func (s *MedicalRecordService) UpdateMedicalRecord(ctx context.Context, recordID string, updates *models.MedicalRecordUpdate) (*models.MedicalRecord, error) {
	// Check if record exists
	_, err := s.medicalRecordRepo.GetByID(ctx, recordID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrMedicalRecordNotFound
		}
		return nil, err
	}

	// Update medical record
	updatedRecord, err := s.medicalRecordRepo.Update(ctx, recordID, updates)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrMedicalRecordNotFound
		}
		return nil, err
	}

	return updatedRecord, nil
}
