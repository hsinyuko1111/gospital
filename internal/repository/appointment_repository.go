package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/hsinyuko1111/gospital/internal/infrastructure"
	"github.com/hsinyuko1111/gospital/internal/models"
)

// AppointmentRepository handles appointment data operations
type AppointmentRepository struct {
	client *infrastructure.MySQLClient
}

// NewAppointmentRepository creates a new appointment repository
func NewAppointmentRepository(client *infrastructure.MySQLClient) *AppointmentRepository {
	return &AppointmentRepository{client: client}
}

// Create creates a new appointment
func (r *AppointmentRepository) Create(ctx context.Context, appointment *models.AppointmentCreate) (*models.Appointment, error) {
	// Generate appointment ID
	appointmentID := generateAppointmentID()

	// Check if time slot is available
	var count int
	checkQuery := `
		SELECT COUNT(*) FROM appointments
		WHERE doctor_id = ?
		  AND appointment_date = ?
		  AND start_time = ?
		  AND status != 'cancelled'
	`
	err := r.client.QueryRowContext(ctx, checkQuery, appointment.DoctorID, appointment.AppointmentDate, appointment.StartTime).Scan(&count)
	if err != nil {
		return nil, err
	}

	if count > 0 {
		return nil, fmt.Errorf("time slot not available")
	}

	now := time.Now()

	query := `
		INSERT INTO appointments (
			appointment_id, patient_id, doctor_id, appointment_date,
			start_time, end_time, reason, status, notes,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, 'scheduled', ?, ?, ?)
	`

	_, err = r.client.ExecContext(
		ctx,
		query,
		appointmentID,
		appointment.PatientID,
		appointment.DoctorID,
		appointment.AppointmentDate,
		appointment.StartTime,
		appointment.EndTime,
		appointment.Reason,
		appointment.Notes,
		now,
		now,
	)

	if err != nil {
		return nil, err
	}

	// Return the created appointment
	return r.GetByID(ctx, appointmentID)
}

// GetByID retrieves an appointment by ID
func (r *AppointmentRepository) GetByID(ctx context.Context, appointmentID string) (*models.Appointment, error) {
	query := `
		SELECT 
			a.appointment_id,
			a.patient_id,
			CONCAT(p.first_name, ' ', p.last_name) as patient_name,
			a.doctor_id,
			CONCAT('Dr. ', d.first_name, ' ', d.last_name) as doctor_name,
			d.department_id,
			a.appointment_date,
			a.start_time,
			a.end_time,
			a.reason,
			a.status,
			a.notes,
			a.created_at,
			a.updated_at
		FROM appointments a
		JOIN patients p ON a.patient_id = p.patient_id
		JOIN doctors d ON a.doctor_id = d.doctor_id
		WHERE a.appointment_id = ?
	`

	var apt models.Appointment
	var notes sql.NullString
	var departmentID sql.NullString

	err := r.client.QueryRowContext(ctx, query, appointmentID).Scan(
		&apt.AppointmentID,
		&apt.PatientID,
		&apt.PatientName,
		&apt.DoctorID,
		&apt.DoctorName,
		&departmentID,
		&apt.AppointmentDate,
		&apt.StartTime,
		&apt.EndTime,
		&apt.Reason,
		&apt.Status,
		&notes,
		&apt.CreatedAt,
		&apt.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("appointment not found")
	}
	if err != nil {
		return nil, err
	}

	if departmentID.Valid {
		apt.DepartmentID = departmentID.String
	}
	if notes.Valid {
		apt.Notes = notes.String
	}

	return &apt, nil
}

// List retrieves appointments with optional filters and pagination
func (r *AppointmentRepository) List(ctx context.Context, filters map[string]interface{}, page, limit int) ([]models.Appointment, *models.Pagination, error) {
	// Build WHERE clause
	whereClauses := []string{}
	args := []interface{}{}

	if patientID, ok := filters["patientId"].(string); ok && patientID != "" {
		whereClauses = append(whereClauses, "a.patient_id = ?")
		args = append(args, patientID)
	}

	if doctorID, ok := filters["doctorId"].(string); ok && doctorID != "" {
		whereClauses = append(whereClauses, "a.doctor_id = ?")
		args = append(args, doctorID)
	}

	if date, ok := filters["date"].(string); ok && date != "" {
		whereClauses = append(whereClauses, "a.appointment_date = ?")
		args = append(args, date)
	}

	if status, ok := filters["status"].(string); ok && status != "" {
		whereClauses = append(whereClauses, "a.status = ?")
		args = append(args, status)
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Count total items
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM appointments a %s", whereSQL)
	var totalItems int
	err := r.client.QueryRowContext(ctx, countQuery, args...).Scan(&totalItems)
	if err != nil {
		return nil, nil, err
	}

	// Calculate pagination
	totalPages := (totalItems + limit - 1) / limit
	offset := (page - 1) * limit

	// Query appointments
	query := fmt.Sprintf(`
		SELECT 
			a.appointment_id,
			a.patient_id,
			CONCAT(p.first_name, ' ', p.last_name) as patient_name,
			a.doctor_id,
			CONCAT('Dr. ', d.first_name, ' ', d.last_name) as doctor_name,
			d.department_id,
			a.appointment_date,
			a.start_time,
			a.end_time,
			a.reason,
			a.status,
			a.notes,
			a.created_at,
			a.updated_at
		FROM appointments a
		JOIN patients p ON a.patient_id = p.patient_id
		JOIN doctors d ON a.doctor_id = d.doctor_id
		%s
		ORDER BY a.appointment_date DESC, a.start_time DESC
		LIMIT ? OFFSET ?
	`, whereSQL)

	args = append(args, limit, offset)
	rows, err := r.client.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	appointments := []models.Appointment{}
	for rows.Next() {
		var apt models.Appointment
		var notes sql.NullString
		var departmentID sql.NullString

		err := rows.Scan(
			&apt.AppointmentID,
			&apt.PatientID,
			&apt.PatientName,
			&apt.DoctorID,
			&apt.DoctorName,
			&departmentID,
			&apt.AppointmentDate,
			&apt.StartTime,
			&apt.EndTime,
			&apt.Reason,
			&apt.Status,
			&notes,
			&apt.CreatedAt,
			&apt.UpdatedAt,
		)
		if err != nil {
			return nil, nil, err
		}

		if departmentID.Valid {
			apt.DepartmentID = departmentID.String
		}
		if notes.Valid {
			apt.Notes = notes.String
		}

		appointments = append(appointments, apt)
	}

	pagination := &models.Pagination{
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
		TotalItems: totalItems,
	}

	return appointments, pagination, nil
}

// Update updates an appointment
func (r *AppointmentRepository) Update(ctx context.Context, appointmentID string, updates *models.AppointmentUpdate) (*models.Appointment, error) {
	// Build SET clause dynamically
	setClauses := []string{"updated_at = ?"}
	args := []interface{}{time.Now()}

	// If rescheduling, check availability
	if updates.AppointmentDate != "" || updates.StartTime != "" {
		// Get current appointment details
		current, err := r.GetByID(ctx, appointmentID)
		if err != nil {
			return nil, err
		}

		newDate := current.AppointmentDate
		newStartTime := current.StartTime
		doctorID := current.DoctorID

		if updates.AppointmentDate != "" {
			newDate = updates.AppointmentDate
		}
		if updates.StartTime != "" {
			newStartTime = updates.StartTime
		}

		// Check if new slot is available
		var count int
		checkQuery := `
			SELECT COUNT(*) FROM appointments
			WHERE doctor_id = ?
			  AND appointment_date = ?
			  AND start_time = ?
			  AND appointment_id != ?
			  AND status != 'cancelled'
		`
		err = r.client.QueryRowContext(ctx, checkQuery, doctorID, newDate, newStartTime, appointmentID).Scan(&count)
		if err != nil {
			return nil, err
		}

		if count > 0 {
			return nil, fmt.Errorf("time slot not available")
		}
	}

	if updates.AppointmentDate != "" {
		setClauses = append(setClauses, "appointment_date = ?")
		args = append(args, updates.AppointmentDate)
	}

	if updates.StartTime != "" {
		setClauses = append(setClauses, "start_time = ?")
		args = append(args, updates.StartTime)
	}

	if updates.EndTime != "" {
		setClauses = append(setClauses, "end_time = ?")
		args = append(args, updates.EndTime)
	}

	if updates.Reason != "" {
		setClauses = append(setClauses, "reason = ?")
		args = append(args, updates.Reason)
	}

	if updates.Status != "" {
		setClauses = append(setClauses, "status = ?")
		args = append(args, updates.Status)
	}

	if updates.Notes != "" {
		setClauses = append(setClauses, "notes = ?")
		args = append(args, updates.Notes)
	}

	// Add appointment ID to args
	args = append(args, appointmentID)

	query := fmt.Sprintf(`
		UPDATE appointments
		SET %s
		WHERE appointment_id = ?
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
		return nil, fmt.Errorf("appointment not found")
	}

	// Return the updated appointment
	return r.GetByID(ctx, appointmentID)
}

// Cancel cancels an appointment
func (r *AppointmentRepository) Cancel(ctx context.Context, appointmentID string) error {
	query := `
		UPDATE appointments
		SET status = 'cancelled', updated_at = ?
		WHERE appointment_id = ?
	`

	result, err := r.client.ExecContext(ctx, query, time.Now(), appointmentID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("appointment not found")
	}

	return nil
}

// Helper function to generate appointment ID using ULID
func generateAppointmentID() string {
	return fmt.Sprintf("APT-%s", ulid.Make().String())
}
