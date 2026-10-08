package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/hsinyuko1111/gospital/internal/infrastructure"
	"github.com/hsinyuko1111/gospital/internal/models"
)

// DoctorRepository handles doctor data operations
type DoctorRepository struct {
	client *infrastructure.MySQLClient
}

// NewDoctorRepository creates a new doctor repository
func NewDoctorRepository(client *infrastructure.MySQLClient) *DoctorRepository {
	return &DoctorRepository{client: client}
}

// Create creates a new doctor
func (r *DoctorRepository) Create(ctx context.Context, doctor *models.DoctorCreate) (*models.Doctor, error) {
	// Generate doctor ID
	doctorID := generateDoctorID()

	// Serialize qualifications to JSON
	qualificationsJSON, err := json.Marshal(doctor.Qualifications)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	query := `
		INSERT INTO doctors (
			doctor_id, first_name, last_name, email, phone,
			specialty, license_number, department_id, years_of_experience,
			qualifications, status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'active', ?, ?)
	`

	_, err = r.client.ExecContext(
		ctx,
		query,
		doctorID,
		doctor.FirstName,
		doctor.LastName,
		doctor.Email,
		doctor.Phone,
		doctor.Specialty,
		doctor.LicenseNumber,
		doctor.DepartmentID,
		doctor.YearsOfExperience,
		qualificationsJSON,
		now,
		now,
	)

	if err != nil {
		return nil, err
	}

	// Return the created doctor
	return r.GetByID(ctx, doctorID)
}

// GetByID retrieves a doctor by ID
func (r *DoctorRepository) GetByID(ctx context.Context, doctorID string) (*models.Doctor, error) {
	query := `
		SELECT 
			d.doctor_id, d.first_name, d.last_name, d.email, d.phone,
			d.specialty, d.license_number, d.department_id, dept.name as department_name,
			d.years_of_experience, d.qualifications, d.status, d.created_at, d.updated_at
		FROM doctors d
		LEFT JOIN departments dept ON d.department_id = dept.department_id
		WHERE d.doctor_id = ?
	`

	var doctor models.Doctor
	var qualificationsJSON []byte

	err := r.client.QueryRowContext(ctx, query, doctorID).Scan(
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

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("doctor not found")
	}
	if err != nil {
		return nil, err
	}

	// Deserialize JSON fields
	if len(qualificationsJSON) > 0 {
		json.Unmarshal(qualificationsJSON, &doctor.Qualifications)
	}

	return &doctor, nil
}

// List retrieves doctors with optional filters and pagination
func (r *DoctorRepository) List(ctx context.Context, filters map[string]interface{}, page, limit int) ([]models.Doctor, *models.Pagination, error) {
	// Build WHERE clause
	whereClauses := []string{}
	args := []interface{}{}

	if departmentID, ok := filters["departmentId"].(string); ok && departmentID != "" {
		whereClauses = append(whereClauses, "d.department_id = ?")
		args = append(args, departmentID)
	}

	if specialty, ok := filters["specialty"].(string); ok && specialty != "" {
		whereClauses = append(whereClauses, "d.specialty = ?")
		args = append(args, specialty)
	}

	if available, ok := filters["available"].(bool); ok {
		if available {
			whereClauses = append(whereClauses, "d.status = 'active'")
		}
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Count total items
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM doctors d %s", whereSQL)
	var totalItems int
	err := r.client.QueryRowContext(ctx, countQuery, args...).Scan(&totalItems)
	if err != nil {
		return nil, nil, err
	}

	// Calculate pagination
	totalPages := (totalItems + limit - 1) / limit
	offset := (page - 1) * limit

	// Query doctors
	query := fmt.Sprintf(`
		SELECT 
			d.doctor_id, d.first_name, d.last_name, d.email, d.phone,
			d.specialty, d.license_number, d.department_id, dept.name as department_name,
			d.years_of_experience, d.qualifications, d.status, d.created_at, d.updated_at
		FROM doctors d
		LEFT JOIN departments dept ON d.department_id = dept.department_id
		%s
		ORDER BY d.created_at DESC
		LIMIT ? OFFSET ?
	`, whereSQL)

	args = append(args, limit, offset)
	rows, err := r.client.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
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
			return nil, nil, err
		}

		// Deserialize JSON fields
		if len(qualificationsJSON) > 0 {
			json.Unmarshal(qualificationsJSON, &doctor.Qualifications)
		}

		doctors = append(doctors, doctor)
	}

	pagination := &models.Pagination{
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
		TotalItems: totalItems,
	}

	return doctors, pagination, nil
}

// Update updates a doctor's information
func (r *DoctorRepository) Update(ctx context.Context, doctorID string, updates *models.DoctorUpdate) (*models.Doctor, error) {
	// Build SET clause dynamically
	setClauses := []string{"updated_at = ?"}
	args := []interface{}{time.Now()}

	if updates.Email != "" {
		setClauses = append(setClauses, "email = ?")
		args = append(args, updates.Email)
	}

	if updates.Phone != "" {
		setClauses = append(setClauses, "phone = ?")
		args = append(args, updates.Phone)
	}

	if updates.Specialty != "" {
		setClauses = append(setClauses, "specialty = ?")
		args = append(args, updates.Specialty)
	}

	if updates.DepartmentID != "" {
		setClauses = append(setClauses, "department_id = ?")
		args = append(args, updates.DepartmentID)
	}

	if updates.YearsOfExperience > 0 {
		setClauses = append(setClauses, "years_of_experience = ?")
		args = append(args, updates.YearsOfExperience)
	}

	if updates.Qualifications != nil {
		qualificationsJSON, _ := json.Marshal(updates.Qualifications)
		setClauses = append(setClauses, "qualifications = ?")
		args = append(args, qualificationsJSON)
	}

	if updates.Status != "" {
		setClauses = append(setClauses, "status = ?")
		args = append(args, updates.Status)
	}

	// Add doctor ID to args
	args = append(args, doctorID)

	query := fmt.Sprintf(`
		UPDATE doctors
		SET %s
		WHERE doctor_id = ?
	`, strings.Join(setClauses, ", "))

	result, err := r.client.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("doctor not found")
	}

	// Return the updated doctor
	return r.GetByID(ctx, doctorID)
}

// Helper function to generate doctor ID using ULID
func generateDoctorID() string {
	return fmt.Sprintf("DOC-%s", ulid.Make().String())
}
