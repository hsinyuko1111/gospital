package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/oklog/ulid/v2"

	"github.com/hsinyuko1111/gospital/internal/infrastructure"
	"github.com/hsinyuko1111/gospital/internal/models"
)

// InvoiceRepository handles invoice data operations
type InvoiceRepository struct {
	client    *infrastructure.DynamoDBClient
	tableName string
}

// NewInvoiceRepository creates a new invoice repository
func NewInvoiceRepository(client *infrastructure.DynamoDBClient, tableName string) *InvoiceRepository {
	return &InvoiceRepository{
		client:    client,
		tableName: tableName,
	}
}

// Create creates a new invoice
func (r *InvoiceRepository) Create(ctx context.Context, invoiceData *models.InvoiceCreate) (*models.Invoice, error) {
	// Generate invoice ID
	invoiceID := generateInvoiceID()
	now := time.Now()

	// Calculate totals
	subtotal := 0.0
	for _, item := range invoiceData.Items {
		item.Amount = float64(item.Quantity) * item.UnitPrice
		subtotal += item.Amount
	}

	tax := subtotal * 0.08 // 8% tax rate (configurable in real app)
	total := subtotal + tax

	// Create full invoice
	invoice := &models.Invoice{
		InvoiceID:     invoiceID,
		PatientID:     invoiceData.PatientID,
		AppointmentID: invoiceData.AppointmentID,
		InvoiceDate:   invoiceData.InvoiceDate,
		DueDate:       invoiceData.DueDate,
		Items:         invoiceData.Items,
		Subtotal:      subtotal,
		Tax:           tax,
		Total:         total,
		AmountPaid:    0.0,
		AmountDue:     total,
		PaymentStatus: "pending",
		Notes:         invoiceData.Notes,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Marshal to DynamoDB attribute values
	item, err := attributevalue.MarshalMap(invoice)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal invoice: %w", err)
	}

	// Put item in DynamoDB
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	return invoice, nil
}

// GetByID retrieves an invoice by ID
func (r *InvoiceRepository) GetByID(ctx context.Context, invoiceID string) (*models.Invoice, error) {
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"invoiceId": &types.AttributeValueMemberS{Value: invoiceID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("invoice not found")
	}

	var invoice models.Invoice
	err = attributevalue.UnmarshalMap(result.Item, &invoice)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal invoice: %w", err)
	}

	return &invoice, nil
}

// GetByPatientID retrieves all invoices for a patient
func (r *InvoiceRepository) GetByPatientID(ctx context.Context, patientID string, status *string) ([]models.Invoice, error) {
	// Build query input
	keyCondition := "patientId = :patientId"
	expressionAttributeValues := map[string]types.AttributeValue{
		":patientId": &types.AttributeValueMemberS{Value: patientID},
	}

	// Add status filter if provided
	var filterExpression *string
	if status != nil && *status != "" {
		filter := "paymentStatus = :status"
		filterExpression = &filter
		expressionAttributeValues[":status"] = &types.AttributeValueMemberS{Value: *status}
	}

	// Query using GSI (assuming GSI with patientId as partition key)
	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(r.tableName),
		IndexName:                 aws.String("PatientIdIndex"),
		KeyConditionExpression:    aws.String(keyCondition),
		ExpressionAttributeValues: expressionAttributeValues,
		ScanIndexForward:          aws.Bool(false), // Sort by date descending
	}

	if filterExpression != nil {
		queryInput.FilterExpression = filterExpression
	}

	result, err := r.client.Query(ctx, queryInput)
	if err != nil {
		return nil, fmt.Errorf("failed to query invoices: %w", err)
	}

	invoices := []models.Invoice{}
	err = attributevalue.UnmarshalListOfMaps(result.Items, &invoices)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal invoices: %w", err)
	}

	return invoices, nil
}

// UpdatePaymentStatus updates the payment status of an invoice
func (r *InvoiceRepository) UpdatePaymentStatus(ctx context.Context, invoiceID string, payment *models.PaymentUpdate) (*models.Invoice, error) {
	// Build update expression
	updateParts := []string{
		"paymentStatus = :status",
		"updatedAt = :updatedAt",
	}
	expressionAttributeValues := map[string]types.AttributeValue{
		":status":    &types.AttributeValueMemberS{Value: payment.PaymentStatus},
		":updatedAt": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
	}

	if payment.PaymentMethod != "" {
		updateParts = append(updateParts, "paymentMethod = :paymentMethod")
		expressionAttributeValues[":paymentMethod"] = &types.AttributeValueMemberS{Value: payment.PaymentMethod}
	}

	if payment.AmountPaid > 0 {
		updateParts = append(updateParts, "amountPaid = :amountPaid")
		expressionAttributeValues[":amountPaid"] = &types.AttributeValueMemberN{Value: fmt.Sprintf("%.2f", payment.AmountPaid)}

		// Get current invoice to calculate amountDue
		currentInvoice, err := r.GetByID(ctx, invoiceID)
		if err != nil {
			return nil, err
		}

		amountDue := currentInvoice.Total - payment.AmountPaid
		updateParts = append(updateParts, "amountDue = :amountDue")
		expressionAttributeValues[":amountDue"] = &types.AttributeValueMemberN{Value: fmt.Sprintf("%.2f", amountDue)}
	}

	if payment.PaymentDate != nil {
		updateParts = append(updateParts, "paymentDate = :paymentDate")
		paymentDateAttr, _ := attributevalue.Marshal(payment.PaymentDate)
		expressionAttributeValues[":paymentDate"] = paymentDateAttr
	}

	if payment.Notes != "" {
		updateParts = append(updateParts, "notes = :notes")
		expressionAttributeValues[":notes"] = &types.AttributeValueMemberS{Value: payment.Notes}
	}

	updateExpression := "SET " + joinUpdateParts(updateParts)

	// Update item in DynamoDB
	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"invoiceId": &types.AttributeValueMemberS{Value: invoiceID},
		},
		UpdateExpression:          aws.String(updateExpression),
		ExpressionAttributeValues: expressionAttributeValues,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update payment status: %w", err)
	}

	// Return updated invoice
	return r.GetByID(ctx, invoiceID)
}

// Helper function to generate invoice ID using ULID
func generateInvoiceID() string {
	return fmt.Sprintf("INV-%s", ulid.Make().String())
}
