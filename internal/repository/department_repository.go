package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/hsinyuko1111/gospital/internal/infrastructure"
	"github.com/hsinyuko1111/gospital/internal/models"
)

// DepartmentRepository handles department data operations
type DepartmentRepository struct {
	client *infrastructure.MySQLClient
}

// NewDepartmentRepository creates a new department repository
func NewDepartmentRepository(client *infrastructure.MySQLClient) *DepartmentRepository {
	return &DepartmentRepository{client: client}
}

// GetAll retrieves all departments
func (r *DepartmentRepository) GetAll(ctx context.Context) ([]models.Department, error) {
	query := `
		SELECT 
			d.department_id,
			d.name,
			d.description,
			d.floor,
			d.extension,
			d.head_doctor_id,
			CONCAT(doc.first_name, ' ', doc.last_name) as head_doctor_name,
			(SELECT COUNT(*) FROM doctors WHERE department_id = d.department_id) as number_of_doctors
		FROM departments d
		LEFT JOIN doctors doc ON d.head_doctor_id = doc.doctor_id
		ORDER BY d.name
	`

	rows, err := r.client.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	departments := []models.Department{}
	for rows.Next() {
		var dept models.Department
		var headDoctorID, headDoctorName sql.NullString

		err := rows.Scan(
			&dept.DepartmentID,
			&dept.Name,
			&dept.Description,
			&dept.Floor,
			&dept.Extension,
			&headDoctorID,
			&headDoctorName,
			&dept.NumberOfDoctors,
		)
		if err != nil {
			return nil, err
		}

		if headDoctorID.Valid {
			dept.HeadDoctorID = headDoctorID.String
		}
		if headDoctorName.Valid {
			dept.HeadDoctorName = headDoctorName.String
		}

		departments = append(departments, dept)
	}

	return departments, nil
}

// GetByID retrieves a department by ID
func (r *DepartmentRepository) GetByID(ctx context.Context, departmentID string) (*models.Department, error) {
	query := `
		SELECT 
			d.department_id,
			d.name,
			d.description,
			d.floor,
			d.extension,
			d.head_doctor_id,
			CONCAT(doc.first_name, ' ', doc.last_name) as head_doctor_name,
			(SELECT COUNT(*) FROM doctors WHERE department_id = d.department_id) as number_of_doctors
		FROM departments d
		LEFT JOIN doctors doc ON d.head_doctor_id = doc.doctor_id
		WHERE d.department_id = ?
	`

	var dept models.Department
	var headDoctorID, headDoctorName sql.NullString

	err := r.client.QueryRowContext(ctx, query, departmentID).Scan(
		&dept.DepartmentID,
		&dept.Name,
		&dept.Description,
		&dept.Floor,
		&dept.Extension,
		&headDoctorID,
		&headDoctorName,
		&dept.NumberOfDoctors,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("department not found")
	}
	if err != nil {
		return nil, err
	}

	if headDoctorID.Valid {
		dept.HeadDoctorID = headDoctorID.String
	}
	if headDoctorName.Valid {
		dept.HeadDoctorName = headDoctorName.String
	}

	return &dept, nil
}

// GetDoctorsByDepartmentID retrieves all doctors in a department
func (r *DepartmentRepository) GetDoctorsByDepartmentID(ctx context.Context, departmentID string) ([]models.Doctor, error) {
	// First verify department exists
	_, err := r.GetByID(ctx, departmentID)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT 
			d.doctor_id, d.first_name, d.last_name, d.email, d.phone,
			d.specialty, d.license_number, d.department_id, dept.name as department_name,
			d.years_of_experience, d.qualifications, d.status, d.created_at, d.updated_at
		FROM doctors d
		LEFT JOIN departments dept ON d.department_id = dept.department_id
		WHERE d.department_id = ?
		ORDER BY d.last_name, d.first_name
	`

	rows, err := r.client.QueryContext(ctx, query, departmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	doctors := []models.Doctor{}
	for rows.Next() {
		var doctor models.Doctor
		var qualificationsJSON []byte

		err := rows.Scan(
			&doctor.DoctorID,
			&doctor.FirstName,
			&doctor.LastName,
			&doctor.Email,
			&doctor.Phone,
			&doctor.Specialty,
			&doctor.LicenseNumber,
			&doctor.DepartmentID,
			&doctor.DepartmentName,
			&doctor.YearsOfExperience,
			&qualificationsJSON,
			&doctor.Status,
			&doctor.CreatedAt,
			&doctor.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Deserialize JSON fields
		if len(qualificationsJSON) > 0 {
			json.Unmarshal(qualificationsJSON, &doctor.Qualifications)
		}

		doctors = append(doctors, doctor)
	}

	return doctors, nil
}
