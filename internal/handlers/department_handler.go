package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hsinyuko1111/gospital/internal/models"
	"github.com/hsinyuko1111/gospital/internal/services"
)

// DepartmentHandler handles department HTTP requests
type DepartmentHandler struct {
	departmentService *services.DepartmentService
}

// NewDepartmentHandler creates a new department handler
func NewDepartmentHandler(departmentService *services.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{
		departmentService: departmentService,
	}
}

// GetAllDepartments godoc
// @Summary List all departments
// @Description Retrieves a list of all hospital departments (MySQL)
// @Tags Departments
// @Accept json
// @Produce json
// @Success 200 {object} object{data=[]models.Department}
// @Router /api/v1/departments [get]
func (h *DepartmentHandler) GetAllDepartments(c *gin.Context) {
	ctx := c.Request.Context()

	departments, err := h.departmentService.GetAllDepartments(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to retrieve departments",
				Details: []string{err.Error()},
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": departments,
	})
}

// GetDepartmentByID godoc
// @Summary Get department details
// @Description Retrieves detailed information about a specific department (MySQL)
// @Tags Departments
// @Accept json
// @Produce json
// @Param departmentId path string true "Unique department identifier"
// @Success 200 {object} models.Department
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/departments/{departmentId} [get]
func (h *DepartmentHandler) GetDepartmentByID(c *gin.Context) {
	ctx := c.Request.Context()
	departmentID := c.Param("departmentId")

	department, err := h.departmentService.GetDepartmentByID(ctx, departmentID)
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
				Message: "Failed to retrieve department",
				Details: []string{err.Error()},
			},
		})
		return
	}

	c.JSON(http.StatusOK, department)
}

// GetDoctorsInDepartment godoc
// @Summary List doctors in department
// @Description Retrieves all doctors belonging to a specific department (MySQL)
// @Tags Departments
// @Accept json
// @Produce json
// @Param departmentId path string true "Unique department identifier"
// @Success 200 {object} object{departmentId=string,departmentName=string,doctors=[]models.Doctor}
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/departments/{departmentId}/doctors [get]
func (h *DepartmentHandler) GetDoctorsInDepartment(c *gin.Context) {
	ctx := c.Request.Context()
	departmentID := c.Param("departmentId")

	doctors, err := h.departmentService.GetDoctorsInDepartment(ctx, departmentID)
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
				Message: "Failed to retrieve doctors",
				Details: []string{err.Error()},
			},
		})
		return
	}

	// Get department info for response
	department, _ := h.departmentService.GetDepartmentByID(ctx, departmentID)

	response := gin.H{
		"departmentId": departmentID,
		"doctors":      doctors,
	}

	if department != nil {
		response["departmentName"] = department.Name
	}

	c.JSON(http.StatusOK, response)
}
