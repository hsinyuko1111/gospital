package services

import (
	"context"
	"strings"

	"github.com/hsinyuko1111/gospital/internal/models"
	"github.com/hsinyuko1111/gospital/internal/repository"
)

// PatientService handles patient business logic
type PatientService struct {
	patientRepo       *repository.PatientRepository
	medicalRecordRepo *repository.MedicalRecordRepository
	invoiceRepo       *repository.InvoiceRepository
}

// NewPatientService creates a new patient service
func NewPatientService(
	patientRepo *repository.PatientRepository,
	medicalRecordRepo *repository.MedicalRecordRepository,
	invoiceRepo *repository.InvoiceRepository,
) *PatientService {
	return &PatientService{
		patientRepo:       patientRepo,
		medicalRecordRepo: medicalRecordRepo,
		invoiceRepo:       invoiceRepo,
	}
}

// CreatePatient creates a new patient
func (s *PatientService) CreatePatient(ctx context.Context, patient *models.PatientCreate) (*models.Patient, error) {
	// Additional validation could go here
	// For example: validate email format, phone format, date of birth, etc.
	// Model binding already handles required fields

	// Create patient
	createdPatient, err := s.patientRepo.Create(ctx, patient)
	if err != nil {
		// Check for duplicate email (if database has unique constraint)
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "already exists") {
			return nil, ErrPatientAlreadyExists
		}
		return nil, err
	}

	return createdPatient, nil
}

// GetPatientByID retrieves a patient by ID
func (s *PatientService) GetPatientByID(ctx context.Context, patientID string) (*models.Patient, error) {
	patient, err := s.patientRepo.GetByID(ctx, patientID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrPatientNotFound
		}
		return nil, err
	}
	return patient, nil
}

// ListPatients retrieves patients with optional filters and pagination
func (s *PatientService) ListPatients(ctx context.Context, filters map[string]interface{}, page, limit int) ([]models.Patient, *models.Pagination, error) {
	// Set default pagination values
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	return s.patientRepo.List(ctx, filters, page, limit)
}

// UpdatePatient updates a patient's information
func (s *PatientService) UpdatePatient(ctx context.Context, patientID string, updates *models.PatientUpdate) (*models.Patient, error) {
	// Check if patient exists
	_, err := s.patientRepo.GetByID(ctx, patientID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrPatientNotFound
		}
		return nil, err
	}

	// Additional validation could go here
	// For example: validate email format, phone format, etc.

	// Update patient
	updatedPatient, err := s.patientRepo.Update(ctx, patientID, updates)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrPatientNotFound
		}
		return nil, err
	}

	return updatedPatient, nil
}

// DeactivatePatient marks a patient as inactive
func (s *PatientService) DeactivatePatient(ctx context.Context, patientID string) error {
	// Check if patient exists
	_, err := s.patientRepo.GetByID(ctx, patientID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return ErrPatientNotFound
		}
		return err
	}

	// Deactivate patient
	err = s.patientRepo.Deactivate(ctx, patientID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return ErrPatientNotFound
		}
		return err
	}

	return nil
}

// GetPatientMedicalRecords retrieves all medical records for a patient
func (s *PatientService) GetPatientMedicalRecords(ctx context.Context, patientID string, startDate, endDate *string) ([]models.MedicalRecord, error) {
	// Check if patient exists
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

// GetPatientInvoices retrieves all invoices for a patient
func (s *PatientService) GetPatientInvoices(ctx context.Context, patientID string, status *string) ([]models.Invoice, error) {
	// Check if patient exists
	_, err := s.patientRepo.GetByID(ctx, patientID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrPatientNotFound
		}
		return nil, err
	}

	// Get invoices from DynamoDB
	invoices, err := s.invoiceRepo.GetByPatientID(ctx, patientID, status)
	if err != nil {
		return nil, err
	}

	return invoices, nil
}
