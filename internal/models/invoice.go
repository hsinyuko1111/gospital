package models

import "time"

// InvoiceItem represents a line item in an invoice
// @Description Individual line item in a billing invoice
type InvoiceItem struct {
	Description string  `json:"description" dynamodbav:"description" example:"Consultation"`
	Quantity    int     `json:"quantity" dynamodbav:"quantity" example:"1"`
	UnitPrice   float64 `json:"unitPrice" dynamodbav:"unitPrice" example:"150.00"`
	Amount      float64 `json:"amount" dynamodbav:"amount" example:"150.00"`
}

// Invoice represents a billing invoice
// @Description Complete billing invoice with line items and payment information
type Invoice struct {
	InvoiceID     string        `json:"invoiceId" dynamodbav:"invoiceId" example:"INV-01JCG3X8P2ABCDEFGHIJKLMNOP"`
	PatientID     string        `json:"patientId" dynamodbav:"patientId" example:"PAT-01JCG3X8P2ABCDEFGHIJKLMNOP"`
	PatientName   string        `json:"patientName,omitempty" dynamodbav:"patientName,omitempty" example:"John Doe"`
	AppointmentID string        `json:"appointmentId,omitempty" dynamodbav:"appointmentId,omitempty" example:"APT-01JCG3X8P2ABCDEFGHIJKLMNOP"`
	InvoiceDate   string        `json:"invoiceDate" dynamodbav:"invoiceDate" example:"2025-11-08"`
	DueDate       string        `json:"dueDate" dynamodbav:"dueDate" example:"2025-12-08"`
	Items         []InvoiceItem `json:"items" dynamodbav:"items"`
	Subtotal      float64       `json:"subtotal" dynamodbav:"subtotal" example:"350.00"`
	Tax           float64       `json:"tax" dynamodbav:"tax" example:"28.00"`
	Total         float64       `json:"total" dynamodbav:"total" example:"378.00"`
	AmountPaid    float64       `json:"amountPaid" dynamodbav:"amountPaid" example:"0.00"`
	AmountDue     float64       `json:"amountDue" dynamodbav:"amountDue" example:"378.00"`
	PaymentStatus string        `json:"paymentStatus" dynamodbav:"paymentStatus" enums:"pending,paid,overdue,cancelled" example:"pending"`
	PaymentMethod string        `json:"paymentMethod,omitempty" dynamodbav:"paymentMethod,omitempty" example:"credit_card"`
	PaymentDate   *time.Time    `json:"paymentDate,omitempty" dynamodbav:"paymentDate,omitempty" example:"2025-11-09T10:00:00Z"`
	Notes         string        `json:"notes,omitempty" dynamodbav:"notes,omitempty" example:"Payment plan arranged"`
	CreatedAt     time.Time     `json:"createdAt" dynamodbav:"createdAt" example:"2025-11-09T10:00:00Z"`
	UpdatedAt     time.Time     `json:"updatedAt" dynamodbav:"updatedAt" example:"2025-11-09T10:00:00Z"`
}

// InvoiceCreate represents the data needed to create a new invoice
// @Description Data required to create a new billing invoice
type InvoiceCreate struct {
	PatientID     string        `json:"patientId" binding:"required" example:"PAT-01JCG3X8P2ABCDEFGHIJKLMNOP"`
	AppointmentID string        `json:"appointmentId,omitempty" example:"APT-01JCG3X8P2ABCDEFGHIJKLMNOP"`
	InvoiceDate   string        `json:"invoiceDate" binding:"required" example:"2025-11-08"`
	DueDate       string        `json:"dueDate" binding:"required" example:"2025-12-08"`
	Items         []InvoiceItem `json:"items" binding:"required,min=1"`
	Notes         string        `json:"notes,omitempty" example:"Payment plan arranged"`
}

// PaymentUpdate represents the data to update payment status
// @Description Data for updating an invoice's payment status and details
type PaymentUpdate struct {
	PaymentStatus string     `json:"paymentStatus" binding:"required" enums:"pending,paid,overdue,cancelled" example:"paid"`
	PaymentMethod string     `json:"paymentMethod,omitempty" example:"credit_card"`
	AmountPaid    float64    `json:"amountPaid,omitempty" example:"378.00"`
	PaymentDate   *time.Time `json:"paymentDate,omitempty" example:"2025-11-09T10:00:00Z"`
	Notes         string     `json:"notes,omitempty" example:"Paid in full"`
}
