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

// MedicalRecordRepository handles medical record data operations
type MedicalRecordRepository struct {
	client    *infrastructure.DynamoDBClient
	tableName string
}

// NewMedicalRecordRepository creates a new medical record repository
func NewMedicalRecordRepository(client *infrastructure.DynamoDBClient, tableName string) *MedicalRecordRepository {
	return &MedicalRecordRepository{
		client:    client,
		tableName: tableName,
	}
}

// Create creates a new medical record
func (r *MedicalRecordRepository) Create(ctx context.Context, record *models.MedicalRecordCreate) (*models.MedicalRecord, error) {
	// Generate record ID
	recordID := generateMedicalRecordID()
	now := time.Now()

	// Create full medical record
	medicalRecord := &models.MedicalRecord{
		RecordID:       recordID,
		PatientID:      record.PatientID,
		DoctorID:       record.DoctorID,
		AppointmentID:  record.AppointmentID,
		VisitDate:      record.VisitDate,
		ChiefComplaint: record.ChiefComplaint,
		Diagnosis:      record.Diagnosis,
		Symptoms:       record.Symptoms,
		Treatment:      record.Treatment,
		Prescriptions:  record.Prescriptions,
		LabResults:     record.LabResults,
		VitalSigns:     record.VitalSigns,
		Notes:          record.Notes,
		FollowUpDate:   record.FollowUpDate,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Marshal to DynamoDB attribute values
	item, err := attributevalue.MarshalMap(medicalRecord)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record: %w", err)
	}

	// Put item in DynamoDB
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create medical record: %w", err)
	}

	return medicalRecord, nil
}

// GetByID retrieves a medical record by ID
func (r *MedicalRecordRepository) GetByID(ctx context.Context, recordID string) (*models.MedicalRecord, error) {
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"recordId": &types.AttributeValueMemberS{Value: recordID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get medical record: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("medical record not found")
	}

	var record models.MedicalRecord
	err = attributevalue.UnmarshalMap(result.Item, &record)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal record: %w", err)
	}

	return &record, nil
}

// GetByPatientID retrieves all medical records for a patient
func (r *MedicalRecordRepository) GetByPatientID(ctx context.Context, patientID string, startDate, endDate *string) ([]models.MedicalRecord, error) {
	// Build query input
	keyCondition := "patientId = :patientId"
	expressionAttributeValues := map[string]types.AttributeValue{
		":patientId": &types.AttributeValueMemberS{Value: patientID},
	}

	// Add date range filter if provided
	var filterExpression *string
	if startDate != nil && endDate != nil {
		filter := "visitDate BETWEEN :startDate AND :endDate"
		filterExpression = &filter
		expressionAttributeValues[":startDate"] = &types.AttributeValueMemberS{Value: *startDate}
		expressionAttributeValues[":endDate"] = &types.AttributeValueMemberS{Value: *endDate}
	} else if startDate != nil {
		filter := "visitDate >= :startDate"
		filterExpression = &filter
		expressionAttributeValues[":startDate"] = &types.AttributeValueMemberS{Value: *startDate}
	} else if endDate != nil {
		filter := "visitDate <= :endDate"
		filterExpression = &filter
		expressionAttributeValues[":endDate"] = &types.AttributeValueMemberS{Value: *endDate}
	}

	// Query using GSI (assuming GSI with patientId as partition key)
	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(r.tableName),
		IndexName:                 aws.String("PatientIdIndex"),
		KeyConditionExpression:    aws.String(keyCondition),
		ExpressionAttributeValues: expressionAttributeValues,
		ScanIndexForward:          aws.Bool(false), // Sort by visitDate descending
	}

	if filterExpression != nil {
		queryInput.FilterExpression = filterExpression
	}

	result, err := r.client.Query(ctx, queryInput)
	if err != nil {
		return nil, fmt.Errorf("failed to query medical records: %w", err)
	}

	records := []models.MedicalRecord{}
	err = attributevalue.UnmarshalListOfMaps(result.Items, &records)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal records: %w", err)
	}

	return records, nil
}

// Update updates a medical record
func (r *MedicalRecordRepository) Update(ctx context.Context, recordID string, updates *models.MedicalRecordUpdate) (*models.MedicalRecord, error) {
	// Build update expression
	updateParts := []string{"updatedAt = :updatedAt"}
	expressionAttributeValues := map[string]types.AttributeValue{
		":updatedAt": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
	}

	if updates.Diagnosis != "" {
		updateParts = append(updateParts, "diagnosis = :diagnosis")
		expressionAttributeValues[":diagnosis"] = &types.AttributeValueMemberS{Value: updates.Diagnosis}
	}

	if updates.Symptoms != nil {
		updateParts = append(updateParts, "symptoms = :symptoms")
		symptomsAttr, _ := attributevalue.Marshal(updates.Symptoms)
		expressionAttributeValues[":symptoms"] = symptomsAttr
	}

	if updates.Treatment != "" {
		updateParts = append(updateParts, "treatment = :treatment")
		expressionAttributeValues[":treatment"] = &types.AttributeValueMemberS{Value: updates.Treatment}
	}

	if updates.Prescriptions != nil {
		updateParts = append(updateParts, "prescriptions = :prescriptions")
		prescriptionsAttr, _ := attributevalue.Marshal(updates.Prescriptions)
		expressionAttributeValues[":prescriptions"] = prescriptionsAttr
	}

	if updates.LabResults != nil {
		updateParts = append(updateParts, "labResults = :labResults")
		labResultsAttr, _ := attributevalue.Marshal(updates.LabResults)
		expressionAttributeValues[":labResults"] = labResultsAttr
	}

	if updates.Notes != "" {
		updateParts = append(updateParts, "notes = :notes")
		expressionAttributeValues[":notes"] = &types.AttributeValueMemberS{Value: updates.Notes}
	}

	if updates.FollowUpDate != "" {
		updateParts = append(updateParts, "followUpDate = :followUpDate")
		expressionAttributeValues[":followUpDate"] = &types.AttributeValueMemberS{Value: updates.FollowUpDate}
	}

	updateExpression := "SET " + joinUpdateParts(updateParts)

	// Update item in DynamoDB
	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"recordId": &types.AttributeValueMemberS{Value: recordID},
		},
		UpdateExpression:          aws.String(updateExpression),
		ExpressionAttributeValues: expressionAttributeValues,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update medical record: %w", err)
	}

	// Return updated record
	return r.GetByID(ctx, recordID)
}

// Helper function to generate medical record ID using ULID
func generateMedicalRecordID() string {
	return fmt.Sprintf("REC-%s", ulid.Make().String())
}

// Helper function to join update parts with commas
func joinUpdateParts(parts []string) string {
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += ", "
		}
		result += part
	}
	return result
}
