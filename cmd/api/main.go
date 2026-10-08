package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	_ "github.com/go-sql-driver/mysql"

	"github.com/hsinyuko1111/gospital/internal/handlers"
	"github.com/hsinyuko1111/gospital/internal/infrastructure"
	"github.com/hsinyuko1111/gospital/internal/repository"
	"github.com/hsinyuko1111/gospital/internal/services"
)

// @title gospital
// @description API for managing hospital operations including patients, doctors, appointments, medical records, and billing
// @version 1.0.0
// @contact.name Hospital API Support
// @contact.email support@hospital.com
// @tag.name Root
// @tag.description Root-level endpoints (health checks, system status)
// @tag.name Patients
// @tag.description Patient management operations
// @tag.name Doctors
// @tag.description Doctor management operations
// @tag.name Appointments
// @tag.description Appointment scheduling and management
// @tag.name Medical Records
// @tag.description Medical records management
// @tag.name Departments
// @tag.description Hospital department information
// @tag.name Billing
// @tag.description Billing and invoice management
func main() {
	// Load configuration from environment variables (required)
	mysqlDSN := getEnvOrFatal("MYSQL_DSN")
	dynamoDBRegion := getEnvOrFatal("DYNAMODB_REGION")
	medicalRecordsTable := getEnvOrFatal("MEDICAL_RECORDS_TABLE")
	invoicesTable := getEnvOrFatal("INVOICES_TABLE")
	serverPort := getEnvOrFatal("SERVER_PORT")

	// Initialize MySQL connection
	mysqlDB, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}
	defer mysqlDB.Close()

	// Configure MySQL connection pool
	mysqlDB.SetMaxOpenConns(25)
	mysqlDB.SetMaxIdleConns(5)

	// Test MySQL connection
	if err := mysqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping MySQL: %v", err)
	}
	log.Println("Connected to MySQL successfully")

	// Create MySQL client with circuit breaker
	mysqlClient := infrastructure.NewMySQLClient(mysqlDB)

	// Initialize DynamoDB client
	ctx := context.Background()
	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(dynamoDBRegion))
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	dynamoDBClient := dynamodb.NewFromConfig(awsCfg)
	log.Println("Connected to DynamoDB successfully")

	// Create DynamoDB client with circuit breaker
	dynamoClient := infrastructure.NewDynamoDBClient(dynamoDBClient)

	// Initialize metrics collectors
	mysqlMetrics := infrastructure.NewMySQLMetricsCollector(mysqlDB)
	runtimeMetrics := infrastructure.NewRuntimeMetricsCollector()

	// Start metrics collection goroutine
	go collectMetricsPeriodically(mysqlMetrics, runtimeMetrics)

	// Initialize repositories
	patientRepo := repository.NewPatientRepository(mysqlClient)
	doctorRepo := repository.NewDoctorRepository(mysqlClient)
	appointmentRepo := repository.NewAppointmentRepository(mysqlClient)
	departmentRepo := repository.NewDepartmentRepository(mysqlClient)
	medicalRecordRepo := repository.NewMedicalRecordRepository(dynamoClient, medicalRecordsTable)
	invoiceRepo := repository.NewInvoiceRepository(dynamoClient, invoicesTable)

	// Initialize services
	patientService := services.NewPatientService(patientRepo, medicalRecordRepo, invoiceRepo)
	doctorService := services.NewDoctorService(doctorRepo, departmentRepo, appointmentRepo)
	appointmentService := services.NewAppointmentService(appointmentRepo, patientRepo, doctorRepo)
	departmentService := services.NewDepartmentService(departmentRepo)
	medicalRecordService := services.NewMedicalRecordService(medicalRecordRepo, patientRepo, doctorRepo)
	invoiceService := services.NewInvoiceService(invoiceRepo, patientRepo)

	// Initialize handlers
	rootHandler := handlers.NewRootHandler(mysqlClient, dynamoClient)
	patientHandler := handlers.NewPatientHandler(patientService)
	doctorHandler := handlers.NewDoctorHandler(doctorService)
	appointmentHandler := handlers.NewAppointmentHandler(appointmentService)
	departmentHandler := handlers.NewDepartmentHandler(departmentService)
	medicalRecordHandler := handlers.NewMedicalRecordHandler(medicalRecordService)
	invoiceHandler := handlers.NewInvoiceHandler(invoiceService)

	// Setup router
	router := SetupRouter(
		rootHandler,
		patientHandler,
		doctorHandler,
		appointmentHandler,
		departmentHandler,
		medicalRecordHandler,
		invoiceHandler,
	)

	// Start server
	serverAddr := fmt.Sprintf(":%s", serverPort)
	log.Printf("Starting server on %s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// getEnvOrFatal gets environment variable or fails with fatal error
func getEnvOrFatal(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Environment variable %s is required but not set", key)
	}
	return value
}

// collectMetricsPeriodically collects metrics at regular intervals
func collectMetricsPeriodically(mysqlMetrics *infrastructure.MySQLMetricsCollector, runtimeMetrics *infrastructure.RuntimeMetricsCollector) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		mysqlMetrics.Collect()
		runtimeMetrics.Collect()
	}
}
