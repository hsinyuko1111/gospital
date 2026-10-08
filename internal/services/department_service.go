package services

import (
	"context"

	"github.com/hsinyuko1111/gospital/internal/models"
	"github.com/hsinyuko1111/gospital/internal/repository"
)

// DepartmentService handles department business logic
type DepartmentService struct {
	departmentRepo *repository.DepartmentRepository
}

// NewDepartmentService creates a new department service
func NewDepartmentService(departmentRepo *repository.DepartmentRepository) *DepartmentService {
	return &DepartmentService{
		departmentRepo: departmentRepo,
	}
}

// GetAllDepartments retrieves all departments
func (s *DepartmentService) GetAllDepartments(ctx context.Context) ([]models.Department, error) {
	return s.departmentRepo.GetAll(ctx)
}

// GetDepartmentByID retrieves a department by ID
func (s *DepartmentService) GetDepartmentByID(ctx context.Context, departmentID string) (*models.Department, error) {
	department, err := s.departmentRepo.GetByID(ctx, departmentID)
	if err != nil {
		return nil, ErrDepartmentNotFound
	}
	return department, nil
}

// GetDoctorsInDepartment retrieves all doctors in a department
func (s *DepartmentService) GetDoctorsInDepartment(ctx context.Context, departmentID string) ([]models.Doctor, error) {
	doctors, err := s.departmentRepo.GetDoctorsByDepartmentID(ctx, departmentID)
	if err != nil {
		// Check if it's a "not found" error
		if err.Error() == "department not found" {
			return nil, ErrDepartmentNotFound
		}
		return nil, err
	}
	return doctors, nil
}
