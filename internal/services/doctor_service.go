package services

import (
	"context"
	"strings"

	"github.com/hsinyuko1111/gospital/internal/models"
	"github.com/hsinyuko1111/gospital/internal/repository"
)

// DoctorService handles doctor business logic
type DoctorService struct {
	doctorRepo      *repository.DoctorRepository
	departmentRepo  *repository.DepartmentRepository
	appointmentRepo *repository.AppointmentRepository
}

// NewDoctorService creates a new doctor service
func NewDoctorService(
	doctorRepo *repository.DoctorRepository,
	departmentRepo *repository.DepartmentRepository,
	appointmentRepo *repository.AppointmentRepository,
) *DoctorService {
	return &DoctorService{
		doctorRepo:      doctorRepo,
		departmentRepo:  departmentRepo,
		appointmentRepo: appointmentRepo,
	}
}

// CreateDoctor creates a new doctor
func (s *DoctorService) CreateDoctor(ctx context.Context, doctor *models.DoctorCreate) (*models.Doctor, error) {
	// Validate department exists
	_, err := s.departmentRepo.GetByID(ctx, doctor.DepartmentID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrDepartmentNotFound
		}
		return nil, err
	}

	// Create doctor
	createdDoctor, err := s.doctorRepo.Create(ctx, doctor)
	if err != nil {
		return nil, err
	}

	return createdDoctor, nil
}

// GetDoctorByID retrieves a doctor by ID
func (s *DoctorService) GetDoctorByID(ctx context.Context, doctorID string) (*models.Doctor, error) {
	doctor, err := s.doctorRepo.GetByID(ctx, doctorID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrDoctorNotFound
		}
		return nil, err
	}
	return doctor, nil
}

// ListDoctors retrieves doctors with optional filters and pagination
func (s *DoctorService) ListDoctors(ctx context.Context, filters map[string]interface{}, page, limit int) ([]models.Doctor, *models.Pagination, error) {
	// Set default pagination values
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	return s.doctorRepo.List(ctx, filters, page, limit)
}

// UpdateDoctor updates a doctor's information
func (s *DoctorService) UpdateDoctor(ctx context.Context, doctorID string, updates *models.DoctorUpdate) (*models.Doctor, error) {
	// Check if doctor exists
	_, err := s.doctorRepo.GetByID(ctx, doctorID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrDoctorNotFound
		}
		return nil, err
	}

	// If department is being updated, validate it exists
	if updates.DepartmentID != "" {
		_, err := s.departmentRepo.GetByID(ctx, updates.DepartmentID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return nil, ErrDepartmentNotFound
			}
			return nil, err
		}
	}

	// Update doctor
	updatedDoctor, err := s.doctorRepo.Update(ctx, doctorID, updates)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrDoctorNotFound
		}
		return nil, err
	}

	return updatedDoctor, nil
}

// GetDoctorAvailability retrieves a doctor's availability schedule
func (s *DoctorService) GetDoctorAvailability(ctx context.Context, doctorID string, startDate, endDate *string) ([]models.TimeSlot, error) {
	// Check if doctor exists
	doctor, err := s.doctorRepo.GetByID(ctx, doctorID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrDoctorNotFound
		}
		return nil, err
	}

	// Check if doctor is available (status = active)
	if doctor.Status != "active" {
		return []models.TimeSlot{}, nil // Return empty availability
	}

	// Get all appointments for this doctor in the date range
	filters := map[string]interface{}{
		"doctorId": doctorID,
	}
	if startDate != nil {
		// Note: This is a simplified implementation
		// In production, you'd query by date range more efficiently
		filters["date"] = *startDate
	}

	appointments, _, err := s.appointmentRepo.List(ctx, filters, 1, 1000)
	if err != nil {
		return nil, err
	}

	// Build a map of booked time slots for quick lookup
	bookedSlots := make(map[string]bool)
	for _, apt := range appointments {
		if apt.Status != "cancelled" {
			// Create a key from date + start time
			slotKey := apt.AppointmentDate + "#" + apt.StartTime
			bookedSlots[slotKey] = true
		}
	}

	// Generate time slots
	// In a real system, this would:
	// 1. Use doctor's working hours from configuration/database
	// 2. Generate slots for the entire date range
	// 3. Account for lunch breaks, holidays, etc.

	// For now, return a simplified version showing which queried appointments are booked
	timeSlots := []models.TimeSlot{}
	for _, apt := range appointments {
		slotKey := apt.AppointmentDate + "#" + apt.StartTime
		timeSlots = append(timeSlots, models.TimeSlot{
			Date:      apt.AppointmentDate,
			StartTime: apt.StartTime,
			EndTime:   apt.EndTime,
			Available: !bookedSlots[slotKey],
		})
	}

	return timeSlots, nil
}
