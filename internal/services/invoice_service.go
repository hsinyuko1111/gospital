package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/hsinyuko1111/gospital/internal/models"
	"github.com/hsinyuko1111/gospital/internal/repository"
)

// InvoiceService handles invoice business logic
type InvoiceService struct {
	invoiceRepo *repository.InvoiceRepository
	patientRepo *repository.PatientRepository
}

// NewInvoiceService creates a new invoice service
func NewInvoiceService(
	invoiceRepo *repository.InvoiceRepository,
	patientRepo *repository.PatientRepository,
) *InvoiceService {
	return &InvoiceService{
		invoiceRepo: invoiceRepo,
		patientRepo: patientRepo,
	}
}

// CreateInvoice creates a new invoice
func (s *InvoiceService) CreateInvoice(ctx context.Context, invoice *models.InvoiceCreate) (*models.Invoice, error) {
	// Validate patient exists and fetch patient name (MySQL)
	patient, err := s.patientRepo.GetByID(ctx, invoice.PatientID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrPatientNotFound
		}
		return nil, err
	}

	// Create invoice in DynamoDB
	createdInvoice, err := s.invoiceRepo.Create(ctx, invoice)
	if err != nil {
		return nil, err
	}

	// Populate patient name in the response (denormalized data for DynamoDB)
	createdInvoice.PatientName = fmt.Sprintf("%s %s", patient.FirstName, patient.LastName)

	return createdInvoice, nil
}

// GetInvoiceByID retrieves an invoice by ID
func (s *InvoiceService) GetInvoiceByID(ctx context.Context, invoiceID string) (*models.Invoice, error) {
	invoice, err := s.invoiceRepo.GetByID(ctx, invoiceID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrInvoiceNotFound
		}
		return nil, err
	}
	return invoice, nil
}

// GetInvoicesByPatientID retrieves all invoices for a patient
func (s *InvoiceService) GetInvoicesByPatientID(ctx context.Context, patientID string, status *string) ([]models.Invoice, error) {
	// Validate patient exists (MySQL)
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

// UpdatePaymentStatus updates the payment status of an invoice
func (s *InvoiceService) UpdatePaymentStatus(ctx context.Context, invoiceID string, payment *models.PaymentUpdate) (*models.Invoice, error) {
	// Check if invoice exists
	_, err := s.invoiceRepo.GetByID(ctx, invoiceID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrInvoiceNotFound
		}
		return nil, err
	}

	// Validate payment status
	validStatuses := map[string]bool{
		"pending":   true,
		"paid":      true,
		"overdue":   true,
		"cancelled": true,
	}
	if !validStatuses[payment.PaymentStatus] {
		return nil, ErrInvalidPaymentData
	}

	// Update payment status
	updatedInvoice, err := s.invoiceRepo.UpdatePaymentStatus(ctx, invoiceID, payment)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrInvoiceNotFound
		}
		return nil, err
	}

	return updatedInvoice, nil
}
