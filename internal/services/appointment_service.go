package services

import (
	"context"
	"strings"

	"github.com/hsinyuko1111/gospital/internal/models"
	"github.com/hsinyuko1111/gospital/internal/repository"
)

// AppointmentService handles appointment business logic
type AppointmentService struct {
	appointmentRepo *repository.AppointmentRepository
	patientRepo     *repository.PatientRepository
	doctorRepo      *repository.DoctorRepository
}

// NewAppointmentService creates a new appointment service
func NewAppointmentService(
	appointmentRepo *repository.AppointmentRepository,
	patientRepo *repository.PatientRepository,
	doctorRepo *repository.DoctorRepository,
) *AppointmentService {
	return &AppointmentService{
		appointmentRepo: appointmentRepo,
		patientRepo:     patientRepo,
		doctorRepo:      doctorRepo,
	}
}

// CreateAppointment creates a new appointment
func (s *AppointmentService) CreateAppointment(ctx context.Context, appointment *models.AppointmentCreate) (*models.Appointment, error) {
	// Validate patient exists
	_, err := s.patientRepo.GetByID(ctx, appointment.PatientID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrPatientNotFound
		}
		return nil, err
	}

	// Validate doctor exists and is active
	doctor, err := s.doctorRepo.GetByID(ctx, appointment.DoctorID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrDoctorNotFound
		}
		return nil, err
	}

	// Check if doctor is available (status = active)
	if doctor.Status != "active" {
		return nil, ErrInvalidAppointmentData
	}

	// Create appointment (repository will check time slot availability)
	createdAppointment, err := s.appointmentRepo.Create(ctx, appointment)
	if err != nil {
		if strings.Contains(err.Error(), "not available") {
			return nil, ErrAppointmentSlotTaken
		}
		return nil, err
	}

	return createdAppointment, nil
}

// GetAppointmentByID retrieves an appointment by ID
func (s *AppointmentService) GetAppointmentByID(ctx context.Context, appointmentID string) (*models.Appointment, error) {
	appointment, err := s.appointmentRepo.GetByID(ctx, appointmentID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrAppointmentNotFound
		}
		return nil, err
	}
	return appointment, nil
}

// ListAppointments retrieves appointments with optional filters and pagination
func (s *AppointmentService) ListAppointments(ctx context.Context, filters map[string]interface{}, page, limit int) ([]models.Appointment, *models.Pagination, error) {
	// Set default pagination values
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	return s.appointmentRepo.List(ctx, filters, page, limit)
}

// UpdateAppointment updates or reschedules an appointment
func (s *AppointmentService) UpdateAppointment(ctx context.Context, appointmentID string, updates *models.AppointmentUpdate) (*models.Appointment, error) {
	// Check if appointment exists
	existingAppointment, err := s.appointmentRepo.GetByID(ctx, appointmentID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrAppointmentNotFound
		}
		return nil, err
	}

	// Don't allow updates to cancelled appointments
	if existingAppointment.Status == "cancelled" {
		return nil, ErrInvalidAppointmentData
	}

	// If rescheduling, verify doctor is still active
	if updates.AppointmentDate != "" || updates.StartTime != "" {
		doctor, err := s.doctorRepo.GetByID(ctx, existingAppointment.DoctorID)
		if err != nil {
			return nil, err
		}
		if doctor.Status != "active" {
			return nil, ErrInvalidAppointmentData
		}
	}

	// Update appointment (repository will check time slot availability if rescheduling)
	updatedAppointment, err := s.appointmentRepo.Update(ctx, appointmentID, updates)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrAppointmentNotFound
		}
		if strings.Contains(err.Error(), "not available") {
			return nil, ErrAppointmentSlotTaken
		}
		return nil, err
	}

	return updatedAppointment, nil
}

// CancelAppointment cancels an appointment
func (s *AppointmentService) CancelAppointment(ctx context.Context, appointmentID string) error {
	// Check if appointment exists
	appointment, err := s.appointmentRepo.GetByID(ctx, appointmentID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return ErrAppointmentNotFound
		}
		return err
	}

	// Don't allow cancelling already cancelled or completed appointments
	if appointment.Status == "cancelled" || appointment.Status == "completed" {
		return ErrInvalidAppointmentData
	}

	// Cancel appointment
	err = s.appointmentRepo.Cancel(ctx, appointmentID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return ErrAppointmentNotFound
		}
		return err
	}

	return nil
}
