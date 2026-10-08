package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hsinyuko1111/gospital/internal/models"
	"github.com/hsinyuko1111/gospital/internal/services"
)

// DoctorHandler handles doctor HTTP requests
type DoctorHandler struct {
	doctorService *services.DoctorService
}

// NewDoctorHandler creates a new doctor handler
func NewDoctorHandler(doctorService *services.DoctorService) *DoctorHandler {
	return &DoctorHandler{
		doctorService: doctorService,
	}
}

// CreateDoctor godoc
// @Summary Register a new doctor
// @Description Creates a new doctor record in the system (MySQL)
// @Tags Doctors
// @Accept json
// @Produce json
// @Param doctor body models.DoctorCreate true "Doctor to register"
// @Success 201 {object} models.Doctor
// @Failure 400 {object} models.ErrorResponse
// @Router /api/v1/doctors [post]
func (h *DoctorHandler) CreateDoctor(c *gin.Context) {
	ctx := c.Request.Context()

	var doctor models.DoctorCreate
	if err := c.ShouldBindJSON(&doctor); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid input data",
				Details: []string{err.Error()},
			},
		})
		return
	}

	createdDoctor, err := h.doctorService.CreateDoctor(ctx, &doctor)
	if err != nil {
		if err == services.ErrDepartmentNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "DEPARTMENT_NOT_FOUND",
					Message: "Department not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to create doctor",
				Details: []string{err.Error()},
			},
		})
		return
	}

	c.JSON(http.StatusCreated, createdDoctor)
}

// GetDoctorByID godoc
// @Summary Get doctor details
// @Description Retrieves detailed information about a specific doctor (MySQL)
// @Tags Doctors
// @Accept json
// @Produce json
// @Param doctorId path string true "Unique doctor identifier"
// @Success 200 {object} models.Doctor
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/doctors/{doctorId} [get]
func (h *DoctorHandler) GetDoctorByID(c *gin.Context) {
	ctx := c.Request.Context()
	doctorID := c.Param("doctorId")

	doctor, err := h.doctorService.GetDoctorByID(ctx, doctorID)
	if err != nil {
		if err == services.ErrDoctorNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "DOCTOR_NOT_FOUND",
					Message: "Doctor not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to retrieve doctor",
				Details: []string{err.Error()},
			},
		})
		return
	}

	c.JSON(http.StatusOK, doctor)
}

// ListDoctors godoc
// @Summary List doctors
// @Description Retrieves a list of doctors with optional filters (MySQL)
// @Tags Doctors
// @Accept json
// @Produce json
// @Param departmentId query string false "Filter by department"
// @Param specialty query string false "Filter by specialty"
// @Param available query bool false "Filter by availability"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(20)
// @Success 200 {object} object{data=[]models.Doctor,pagination=models.Pagination}
// @Router /api/v1/doctors [get]
func (h *DoctorHandler) ListDoctors(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse query parameters
	filters := make(map[string]interface{})
	if departmentID := c.Query("departmentId"); departmentID != "" {
		filters["departmentId"] = departmentID
	}
	if specialty := c.Query("specialty"); specialty != "" {
		filters["specialty"] = specialty
	}
	if available := c.Query("available"); available != "" {
		filters["available"] = available == "true"
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	doctors, pagination, err := h.doctorService.ListDoctors(ctx, filters, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to retrieve doctors",
				Details: []string{err.Error()},
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       doctors,
		"pagination": pagination,
	})
}

// UpdateDoctor godoc
// @Summary Update doctor information
// @Description Updates existing doctor information (MySQL)
// @Tags Doctors
// @Accept json
// @Produce json
// @Param doctorId path string true "Unique doctor identifier"
// @Param updates body models.DoctorUpdate true "Doctor updates"
// @Success 200 {object} models.Doctor
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/doctors/{doctorId} [put]
func (h *DoctorHandler) UpdateDoctor(c *gin.Context) {
	ctx := c.Request.Context()
	doctorID := c.Param("doctorId")

	var updates models.DoctorUpdate
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

	updatedDoctor, err := h.doctorService.UpdateDoctor(ctx, doctorID, &updates)
	if err != nil {
		if err == services.ErrDoctorNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "DOCTOR_NOT_FOUND",
					Message: "Doctor not found",
				},
			})
			return
		}
		if err == services.ErrDepartmentNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "DEPARTMENT_NOT_FOUND",
					Message: "Department not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to update doctor",
				Details: []string{err.Error()},
			},
		})
		return
	}

	c.JSON(http.StatusOK, updatedDoctor)
}

// GetDoctorAvailability godoc
// @Summary Get doctor's availability schedule
// @Description Retrieves the availability schedule for a specific doctor (MySQL)
// @Tags Doctors
// @Accept json
// @Produce json
// @Param doctorId path string true "Unique doctor identifier"
// @Param startDate query string false "Start date for availability check" Format(date)
// @Param endDate query string false "End date for availability check" Format(date)
// @Success 200 {object} object{doctorId=string,availability=[]models.TimeSlot}
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/doctors/{doctorId}/availability [get]
func (h *DoctorHandler) GetDoctorAvailability(c *gin.Context) {
	ctx := c.Request.Context()
	doctorID := c.Param("doctorId")

	var startDate, endDate *string
	if sd := c.Query("startDate"); sd != "" {
		startDate = &sd
	}
	if ed := c.Query("endDate"); ed != "" {
		endDate = &ed
	}

	availability, err := h.doctorService.GetDoctorAvailability(ctx, doctorID, startDate, endDate)
	if err != nil {
		if err == services.ErrDoctorNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "DOCTOR_NOT_FOUND",
					Message: "Doctor not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to retrieve availability",
				Details: []string{err.Error()},
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"doctorId":     doctorID,
		"availability": availability,
	})
}
