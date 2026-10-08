package main

import (
	"time"

	"github.com/gin-gonic/gin"

	_ "github.com/hsinyuko1111/gospital/docs"
	"github.com/hsinyuko1111/gospital/internal/handlers"
	"github.com/hsinyuko1111/gospital/internal/middleware"
)

const (
	cpuOverloadThreshold = 95.0
	cpuRecoveryThreshold = 85.0
	requestTimeout       = 100 * time.Millisecond
)

func SetupRouter(
	rootHandler *handlers.RootHandler,
	patientHandler *handlers.PatientHandler,
	doctorHandler *handlers.DoctorHandler,
	appointmentHandler *handlers.AppointmentHandler,
	departmentHandler *handlers.DepartmentHandler,
	medicalRecordHandler *handlers.MedicalRecordHandler,
	invoiceHandler *handlers.InvoiceHandler,
) *gin.Engine {
	router := gin.Default()

	cpuCircuitBreaker := middleware.NewCPUCircuitBreaker(cpuOverloadThreshold, cpuRecoveryThreshold)

	router.GET("/health", rootHandler.HealthCheck)
	router.GET("/metrics", rootHandler.Metrics)
	router.GET("/swagger/*any", rootHandler.Swagger)

	api := router.Group("/api")
	api.Use(middleware.PrometheusMiddleware())
	api.Use(middleware.TimeoutMiddleware(requestTimeout))
	api.Use(middleware.CPUCircuitBreakerMiddleware(cpuCircuitBreaker))
	{
		v1 := api.Group("/v1")
		{
			patients := v1.Group("/patients")
			{
				patients.POST("", patientHandler.CreatePatient)
				patients.GET("", patientHandler.ListPatients)
				patients.GET("/:patientId", patientHandler.GetPatientByID)
				patients.PUT("/:patientId", patientHandler.UpdatePatient)
				patients.DELETE("/:patientId", patientHandler.DeactivatePatient)
				patients.GET("/:patientId/medical-records", patientHandler.GetPatientMedicalRecords)
				patients.GET("/:patientId/invoices", patientHandler.GetPatientInvoices)
			}

			doctors := v1.Group("/doctors")
			{
				doctors.POST("", doctorHandler.CreateDoctor)
				doctors.GET("", doctorHandler.ListDoctors)
				doctors.GET("/:doctorId", doctorHandler.GetDoctorByID)
				doctors.PUT("/:doctorId", doctorHandler.UpdateDoctor)
				doctors.GET("/:doctorId/availability", doctorHandler.GetDoctorAvailability)
			}

			appointments := v1.Group("/appointments")
			{
				appointments.POST("", appointmentHandler.CreateAppointment)
				appointments.GET("", appointmentHandler.ListAppointments)
				appointments.GET("/:appointmentId", appointmentHandler.GetAppointmentByID)
				appointments.PUT("/:appointmentId", appointmentHandler.UpdateAppointment)
				appointments.DELETE("/:appointmentId", appointmentHandler.CancelAppointment)
			}

			departments := v1.Group("/departments")
			{
				departments.GET("", departmentHandler.GetAllDepartments)
				departments.GET("/:departmentId", departmentHandler.GetDepartmentByID)
				departments.GET("/:departmentId/doctors", departmentHandler.GetDoctorsInDepartment)
			}

			medicalRecords := v1.Group("/medical-records")
			{
				medicalRecords.POST("", medicalRecordHandler.CreateMedicalRecord)
				medicalRecords.GET("/:recordId", medicalRecordHandler.GetMedicalRecordByID)
				medicalRecords.PUT("/:recordId", medicalRecordHandler.UpdateMedicalRecord)
			}

			billing := v1.Group("/billing")
			{
				invoices := billing.Group("/invoices")
				{
					invoices.POST("", invoiceHandler.CreateInvoice)
					invoices.GET("/:invoiceId", invoiceHandler.GetInvoiceByID)
					invoices.PUT("/:invoiceId/payment", invoiceHandler.UpdatePaymentStatus)
				}
			}
		}
	}

	return router
}
