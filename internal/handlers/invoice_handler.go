package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hsinyuko1111/gospital/internal/models"
	"github.com/hsinyuko1111/gospital/internal/services"
)

// InvoiceHandler handles invoice HTTP requests
type InvoiceHandler struct {
	invoiceService *services.InvoiceService
}

// NewInvoiceHandler creates a new invoice handler
func NewInvoiceHandler(invoiceService *services.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{
		invoiceService: invoiceService,
	}
}

// CreateInvoice godoc
// @Summary Create a new invoice
// @Description Creates a new billing invoice for a patient (stored in DynamoDB)
// @Tags Billing
// @Accept json
// @Produce json
// @Param invoice body models.InvoiceCreate true "Invoice to create"
// @Success 201 {object} models.Invoice
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/billing/invoices [post]
func (h *InvoiceHandler) CreateInvoice(c *gin.Context) {
	ctx := c.Request.Context()

	var invoice models.InvoiceCreate
	if err := c.ShouldBindJSON(&invoice); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid input data",
				Details: []string{err.Error()},
			},
		})
		return
	}

	createdInvoice, err := h.invoiceService.CreateInvoice(ctx, &invoice)
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

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to create invoice",
				Details: []string{err.Error()},
			},
		})
		return
	}

	c.JSON(http.StatusCreated, createdInvoice)
}

// GetInvoiceByID godoc
// @Summary Get invoice details
// @Description Retrieves detailed information about a specific invoice (DynamoDB)
// @Tags Billing
// @Accept json
// @Produce json
// @Param invoiceId path string true "Unique invoice identifier"
// @Success 200 {object} models.Invoice
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/billing/invoices/{invoiceId} [get]
func (h *InvoiceHandler) GetInvoiceByID(c *gin.Context) {
	ctx := c.Request.Context()
	invoiceID := c.Param("invoiceId")

	invoice, err := h.invoiceService.GetInvoiceByID(ctx, invoiceID)
	if err != nil {
		if err == services.ErrInvoiceNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "INVOICE_NOT_FOUND",
					Message: "Invoice not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to retrieve invoice",
				Details: []string{err.Error()},
			},
		})
		return
	}

	c.JSON(http.StatusOK, invoice)
}

// UpdatePaymentStatus godoc
// @Summary Update payment status
// @Description Updates the payment status of an invoice (DynamoDB)
// @Tags Billing
// @Accept json
// @Produce json
// @Param invoiceId path string true "Unique invoice identifier"
// @Param payment body models.PaymentUpdate true "Payment status update"
// @Success 200 {object} models.Invoice
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/billing/invoices/{invoiceId}/payment [put]
func (h *InvoiceHandler) UpdatePaymentStatus(c *gin.Context) {
	ctx := c.Request.Context()
	invoiceID := c.Param("invoiceId")

	var payment models.PaymentUpdate
	if err := c.ShouldBindJSON(&payment); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid input data",
				Details: []string{err.Error()},
			},
		})
		return
	}

	updatedInvoice, err := h.invoiceService.UpdatePaymentStatus(ctx, invoiceID, &payment)
	if err != nil {
		if err == services.ErrInvoiceNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "INVOICE_NOT_FOUND",
					Message: "Invoice not found",
				},
			})
			return
		}
		if err == services.ErrInvalidPaymentData {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "INVALID_PAYMENT_DATA",
					Message: "Invalid payment data",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to update payment status",
				Details: []string{err.Error()},
			},
		})
		return
	}

	c.JSON(http.StatusOK, updatedInvoice)
}
