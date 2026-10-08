# Gospital

A scalable hospital management system API written in Go with MySQL and DynamoDB backends for optimized healthcare data management.

## Introduction

Healthcare API failures directly impact patient safety:

- Inaccessible medical records delay treatment decisions
- Appointment outages prevent patients from receiving care
- Billing downtime disrupts hospital operations

Using a hybrid database architecture (MySQL + DynamoDB) improves performance but introduces new failure modes. Gospital addresses these challenges through circuit breakers, bulkhead isolation, request timeouts, and auto-scaling.

## Project Structure

```
gospital/
├── cmd/
│   └── api/
│       ├── main.go                 # Application entry point
│       └── router.go               # HTTP router setup
├── internal/
│   ├── handlers/                   # HTTP request handlers
│   │   ├── appointment_handler.go
│   │   ├── department_handler.go
│   │   ├── doctor_handler.go
│   │   ├── invoice_handler.go
│   │   ├── medical_record_handler.go
│   │   ├── patient_handler.go
│   │   └── root_handler.go
│   ├── infrastructure/             # Database clients with circuit breakers
│   │   ├── dynamodb_client.go
│   │   ├── metrics.go
│   │   └── mysql_client.go
│   ├── middleware/                 # HTTP middleware
│   │   ├── cpu_circuit_breaker.go
│   │   ├── prometheus.go
│   │   └── timeout.go
│   ├── models/                     # Data models
│   │   ├── appointment.go
│   │   ├── department.go
│   │   ├── doctor.go
│   │   ├── error.go
│   │   ├── invoice.go
│   │   ├── medical_record.go
│   │   ├── pagination.go
│   │   ├── patient.go
│   │   └── timeslot.go
│   ├── repository/                 # Data access layer
│   │   ├── appointment_repository.go
│   │   ├── department_repository.go
│   │   ├── doctor_repository.go
│   │   ├── invoice_repository.go
│   │   ├── medical_record_repository.go
│   │   └── patient_repository.go
│   └── services/                   # Business logic layer
│       ├── appointment_service.go
│       ├── department_service.go
│       ├── doctor_service.go
│       ├── errors.go
│       ├── invoice_service.go
│       ├── medical_record_service.go
│       └── patient_service.go
├── load_test/                      # Locust load testing scripts
│   ├── locustfile_bulkhead.py
│   └── locustfile_database.py
├── scripts/
│   ├── dashboard.json              # Grafana dashboard configuration
│   └── init.sql                    # MySQL schema initialization
├── terraform/                      # Infrastructure as Code
│   ├── alb.tf
│   ├── docker.tf
│   ├── dynamodb.tf
│   ├── ecr.tf
│   ├── ecs.tf
│   ├── grafana.tf
│   ├── iam.tf
│   ├── outputs.tf
│   ├── prometheus.tf
│   ├── provider.tf
│   ├── rds.tf
│   ├── security_groups.tf
│   ├── service_discovery.tf
│   ├── variables.tf
│   └── vpc.tf
├── Dockerfile
├── Makefile
└── README.md
```
