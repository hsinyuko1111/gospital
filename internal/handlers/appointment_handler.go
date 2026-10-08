package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hsinyuko1111/gospital/internal/models"
	"github.com/hsinyuko1111/gospital/internal/services"
)

// AppointmentHandler handles appointment HTTP requests
type AppointmentHandler struct {
	appointmentService *services.AppointmentService
}

// NewAppointmentHandler creates a new appointment handler
func NewAppointmentHandler(appointmentService *services.AppointmentService) *AppointmentHandler {
	return &AppointmentHandler{
		appointmentService: appointmentService,
	}
}

// CreateAppointment godoc
// @Summary Book a new appointment
// @Description Creates a new appointment (MySQL)
// @Tags Appointments
// @Accept json
// @Produce json
// @Param appointment body models.AppointmentCreate true "Appointment to book"
// @Success 201 {object} models.Appointment
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/appointments [post]
func (h *AppointmentHandler) CreateAppointment(c *gin.Context) {
	ctx := c.Request.Context()

	var appointment models.AppointmentCreate
	if err := c.ShouldBindJSON(&appointment); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid input data",
				Details: []string{err.Error()},
			},
		})
		return
	}

	createdAppointment, err := h.appointmentService.CreateAppointment(ctx, &appointment)
	if err != nil {
		if err == services.ErrPatientNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "PATIENT_NOT_FOUND",
					Message: "Patient not found",
				},
			})
			return
		}
		if err == services.ErrDoctorNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "DOCTOR_NOT_FOUND",
					Message: "Doctor not found",
				},
			})
			return
		}
		if err == services.ErrAppointmentSlotTaken {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "TIME_SLOT_NOT_AVAILABLE",
					Message: "Time slot is not available",
				},
			})
			return
		}
		if err == services.ErrInvalidAppointmentData {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "INVALID_APPOINTMENT_DATA",
					Message: "Invalid appointment data",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to create appointment",
				Details: []string{err.Error()},
			},
		})
		return
	}

	c.JSON(http.StatusCreated, createdAppointment)
}

// GetAppointmentByID godoc
// @Summary Get appointment details
// @Description Retrieves detailed information about a specific appointment (MySQL)
// @Tags Appointments
// @Accept json
// @Produce json
// @Param appointmentId path string true "Unique appointment identifier"
// @Success 200 {object} models.Appointment
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/appointments/{appointmentId} [get]
func (h *AppointmentHandler) GetAppointmentByID(c *gin.Context) {
	ctx := c.Request.Context()
	appointmentID := c.Param("appointmentId")

	appointment, err := h.appointmentService.GetAppointmentByID(ctx, appointmentID)
	if err != nil {
		if err == services.ErrAppointmentNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "APPOINTMENT_NOT_FOUND",
					Message: "Appointment not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to retrieve appointment",
				Details: []string{err.Error()},
			},
		})
		return
	}

	c.JSON(http.StatusOK, appointment)
}

// ListAppointments godoc
// @Summary List appointments
// @Description Retrieves a list of appointments with optional filters (MySQL)
// @Tags Appointments
// @Accept json
// @Produce json
// @Param patientId query string false "Filter by patient"
// @Param doctorId query string false "Filter by doctor"
// @Param date query string false "Filter by appointment date" Format(date)
// @Param status query string false "Filter by appointment status" Enums(scheduled, completed, cancelled, no-show)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(20)
// @Success 200 {object} object{data=[]models.Appointment,pagination=models.Pagination}
// @Router /api/v1/appointments [get]
func (h *AppointmentHandler) ListAppointments(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse query parameters
	filters := make(map[string]interface{})
	if patientID := c.Query("patientId"); patientID != "" {
		filters["patientId"] = patientID
	}
	if doctorID := c.Query("doctorId"); doctorID != "" {
		filters["doctorId"] = doctorID
	}
	if date := c.Query("date"); date != "" {
		filters["date"] = date
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	appointments, pagination, err := h.appointmentService.ListAppointments(ctx, filters, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to retrieve appointments",
				Details: []string{err.Error()},
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       appointments,
		"pagination": pagination,
	})
}

// UpdateAppointment godoc
// @Summary Update or reschedule appointment
// @Description Updates appointment details or reschedules it (MySQL)
// @Tags Appointments
// @Accept json
// @Produce json
// @Param appointmentId path string true "Unique appointment identifier"
// @Param updates body models.AppointmentUpdate true "Appointment updates"
// @Success 200 {object} models.Appointment
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/appointments/{appointmentId} [put]
func (h *AppointmentHandler) UpdateAppointment(c *gin.Context) {
	ctx := c.Request.Context()
	appointmentID := c.Param("appointmentId")

	var updates models.AppointmentUpdate
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid input data",
				Details: []string{err.Error()},
			},
		})
		return
	}

	updatedAppointment, err := h.appointmentService.UpdateAppointment(ctx, appointmentID, &updates)
	if err != nil {
		if err == services.ErrAppointmentNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "APPOINTMENT_NOT_FOUND",
					Message: "Appointment not found",
				},
			})
			return
		}
		if err == services.ErrAppointmentSlotTaken {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "TIME_SLOT_NOT_AVAILABLE",
					Message: "Time slot is not available",
				},
			})
			return
		}
		if err == services.ErrInvalidAppointmentData {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "INVALID_APPOINTMENT_DATA",
					Message: "Invalid appointment data",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to update appointment",
				Details: []string{err.Error()},
			},
		})
		return
	}

	c.JSON(http.StatusOK, updatedAppointment)
}

// CancelAppointment godoc
// @Summary Cancel appointment
// @Description Cancels an existing appointment (MySQL)
// @Tags Appointments
// @Accept json
// @Produce json
// @Param appointmentId path string true "Unique appointment identifier"
// @Success 200 {object} object{message=string,appointmentId=string}
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/appointments/{appointmentId} [delete]
func (h *AppointmentHandler) CancelAppointment(c *gin.Context) {
	ctx := c.Request.Context()
	appointmentID := c.Param("appointmentId")

	err := h.appointmentService.CancelAppointment(ctx, appointmentID)
	if err != nil {
		if err == services.ErrAppointmentNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "APPOINTMENT_NOT_FOUND",
					Message: "Appointment not found",
				},
			})
			return
		}
		if err == services.ErrInvalidAppointmentData {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "INVALID_APPOINTMENT_DATA",
					Message: "Cannot cancel this appointment",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to cancel appointment",
				Details: []string{err.Error()},
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Appointment cancelled successfully",
		"appointmentId": appointmentID,
	})
}
