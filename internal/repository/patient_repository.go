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

// PatientRepository handles patient data operations
type PatientRepository struct {
	client *infrastructure.MySQLClient
}

// NewPatientRepository creates a new patient repository
func NewPatientRepository(client *infrastructure.MySQLClient) *PatientRepository {
	return &PatientRepository{client: client}
}

// Create creates a new patient
func (r *PatientRepository) Create(ctx context.Context, patient *models.PatientCreate) (*models.Patient, error) {
	// Generate patient ID
	patientID := generatePatientID()

	// Serialize nested objects to JSON
	addressJSON, err := json.Marshal(patient.Address)
	if err != nil {
		return nil, err
	}

	emergencyContactJSON, err := json.Marshal(patient.EmergencyContact)
	if err != nil {
		return nil, err
	}

	allergiesJSON, err := json.Marshal(patient.Allergies)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	query := `
		INSERT INTO patients (
			patient_id, first_name, last_name, date_of_birth, gender,
			email, phone, address, emergency_contact, blood_type,
			allergies, status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'active', ?, ?)
	`

	_, err = r.client.ExecContext(
		ctx,
		query,
		patientID,
		patient.FirstName,
		patient.LastName,
		patient.DateOfBirth,
		patient.Gender,
		patient.Email,
		patient.Phone,
		addressJSON,
		emergencyContactJSON,
		patient.BloodType,
		allergiesJSON,
		now,
		now,
	)

	if err != nil {
		return nil, err
	}

	// Return the created patient
	return r.GetByID(ctx, patientID)
}

// GetByID retrieves a patient by ID
func (r *PatientRepository) GetByID(ctx context.Context, patientID string) (*models.Patient, error) {
	query := `
		SELECT 
			patient_id, first_name, last_name, date_of_birth, gender,
			email, phone, address, emergency_contact, blood_type,
			allergies, status, created_at, updated_at
		FROM patients
		WHERE patient_id = ?
	`

	var patient models.Patient
	var addressJSON, emergencyContactJSON, allergiesJSON []byte

	err := r.client.QueryRowContext(ctx, query, patientID).Scan(
		&patient.PatientID,
		&patient.FirstName,
		&patient.LastName,
		&patient.DateOfBirth,
		&patient.Gender,
		&patient.Email,
		&patient.Phone,
		&addressJSON,
		&emergencyContactJSON,
		&patient.BloodType,
		&allergiesJSON,
		&patient.Status,
		&patient.CreatedAt,
		&patient.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("patient not found")
	}
	if err != nil {
		return nil, err
	}

	// Deserialize JSON fields
	if len(addressJSON) > 0 {
		json.Unmarshal(addressJSON, &patient.Address)
	}
	if len(emergencyContactJSON) > 0 {
		json.Unmarshal(emergencyContactJSON, &patient.EmergencyContact)
	}
	if len(allergiesJSON) > 0 {
		json.Unmarshal(allergiesJSON, &patient.Allergies)
	}

	return &patient, nil
}

// List retrieves patients with optional filters and pagination
func (r *PatientRepository) List(ctx context.Context, filters map[string]interface{}, page, limit int) ([]models.Patient, *models.Pagination, error) {
	// Build WHERE clause
	whereClauses := []string{}
	args := []interface{}{}

	if name, ok := filters["name"].(string); ok && name != "" {
		whereClauses = append(whereClauses, "(first_name LIKE ? OR last_name LIKE ?)")
		searchTerm := "%" + name + "%"
		args = append(args, searchTerm, searchTerm)
	}

	if email, ok := filters["email"].(string); ok && email != "" {
		whereClauses = append(whereClauses, "email = ?")
		args = append(args, email)
	}

	if phone, ok := filters["phone"].(string); ok && phone != "" {
		whereClauses = append(whereClauses, "phone = ?")
		args = append(args, phone)
	}

	if status, ok := filters["status"].(string); ok && status != "" {
		whereClauses = append(whereClauses, "status = ?")
		args = append(args, status)
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Count total items
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM patients %s", whereSQL)
	var totalItems int
	err := r.client.QueryRowContext(ctx, countQuery, args...).Scan(&totalItems)
	if err != nil {
		return nil, nil, err
	}

	// Calculate pagination
	totalPages := (totalItems + limit - 1) / limit
	offset := (page - 1) * limit

	// Query patients
	query := fmt.Sprintf(`
		SELECT 
			patient_id, first_name, last_name, date_of_birth, gender,
			email, phone, address, emergency_contact, blood_type,
			allergies, status, created_at, updated_at
		FROM patients
		%s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, whereSQL)

	args = append(args, limit, offset)
	rows, err := r.client.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	patients := []models.Patient{}
	for rows.Next() {
		var patient models.Patient
		var addressJSON, emergencyContactJSON, allergiesJSON []byte

		err := rows.Scan(
			&patient.PatientID,
			&patient.FirstName,
			&patient.LastName,
			&patient.DateOfBirth,
			&patient.Gender,
			&patient.Email,
			&patient.Phone,
			&addressJSON,
			&emergencyContactJSON,
			&patient.BloodType,
			&allergiesJSON,
			&patient.Status,
			&patient.CreatedAt,
			&patient.UpdatedAt,
		)
		if err != nil {
			return nil, nil, err
		}

		// Deserialize JSON fields
		if len(addressJSON) > 0 {
			json.Unmarshal(addressJSON, &patient.Address)
		}
		if len(emergencyContactJSON) > 0 {
			json.Unmarshal(emergencyContactJSON, &patient.EmergencyContact)
		}
		if len(allergiesJSON) > 0 {
			json.Unmarshal(allergiesJSON, &patient.Allergies)
		}

		patients = append(patients, patient)
	}

	pagination := &models.Pagination{
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
		TotalItems: totalItems,
	}

	return patients, pagination, nil
}

// Update updates a patient's information
func (r *PatientRepository) Update(ctx context.Context, patientID string, updates *models.PatientUpdate) (*models.Patient, error) {
	// Build SET clause dynamically
	setClauses := []string{"updated_at = ?"}
	args := []interface{}{time.Now()}

	if updates.FirstName != "" {
		setClauses = append(setClauses, "first_name = ?")
		args = append(args, updates.FirstName)
	}

	if updates.LastName != "" {
		setClauses = append(setClauses, "last_name = ?")
		args = append(args, updates.LastName)
	}

	if updates.Email != "" {
		setClauses = append(setClauses, "email = ?")
		args = append(args, updates.Email)
	}

	if updates.Phone != "" {
		setClauses = append(setClauses, "phone = ?")
		args = append(args, updates.Phone)
	}

	if updates.Address != nil {
		addressJSON, _ := json.Marshal(updates.Address)
		setClauses = append(setClauses, "address = ?")
		args = append(args, addressJSON)
	}

	if updates.EmergencyContact != nil {
		emergencyContactJSON, _ := json.Marshal(updates.EmergencyContact)
		setClauses = append(setClauses, "emergency_contact = ?")
		args = append(args, emergencyContactJSON)
	}

	if updates.Allergies != nil {
		allergiesJSON, _ := json.Marshal(updates.Allergies)
		setClauses = append(setClauses, "allergies = ?")
		args = append(args, allergiesJSON)
	}

	// Add patient ID to args
	args = append(args, patientID)

	query := fmt.Sprintf(`
		UPDATE patients
		SET %s
		WHERE patient_id = ?
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
		return nil, fmt.Errorf("patient not found")
	}

	// Return the updated patient
	return r.GetByID(ctx, patientID)
}

// Deactivate marks a patient as inactive
func (r *PatientRepository) Deactivate(ctx context.Context, patientID string) error {
	query := `
		UPDATE patients
		SET status = 'inactive', updated_at = ?
		WHERE patient_id = ?
	`

	result, err := r.client.ExecContext(ctx, query, time.Now(), patientID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("patient not found")
	}

	return nil
}

// Helper function to generate patient ID using ULID
func generatePatientID() string {
	return fmt.Sprintf("PAT-%s", ulid.Make().String())
}
