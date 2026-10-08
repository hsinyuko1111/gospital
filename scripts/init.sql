-- Gospital Database Initialization Script
-- Creates all tables and indexes for MySQL database

-- Create departments table
CREATE TABLE IF NOT EXISTS departments (
  department_id VARCHAR(50) PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  description TEXT,
  floor INT,
  extension VARCHAR(20),
  head_doctor_id VARCHAR(50),
  INDEX idx_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Create patients table
CREATE TABLE IF NOT EXISTS patients (
  patient_id VARCHAR(50) PRIMARY KEY,
  first_name VARCHAR(100) NOT NULL,
  last_name VARCHAR(100) NOT NULL,
  date_of_birth DATE NOT NULL,
  gender VARCHAR(10),
  email VARCHAR(255) NOT NULL,
  phone VARCHAR(20) NOT NULL,
  address JSON,
  emergency_contact JSON,
  blood_type VARCHAR(5),
  allergies JSON,
  status VARCHAR(10) NOT NULL DEFAULT 'active',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_email (email),
  INDEX idx_phone (phone),
  INDEX idx_status (status),
  INDEX idx_name (first_name, last_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Create doctors table
CREATE TABLE IF NOT EXISTS doctors (
  doctor_id VARCHAR(50) PRIMARY KEY,
  first_name VARCHAR(100) NOT NULL,
  last_name VARCHAR(100) NOT NULL,
  email VARCHAR(255) NOT NULL,
  phone VARCHAR(20) NOT NULL,
  specialty VARCHAR(100) NOT NULL,
  license_number VARCHAR(50) NOT NULL,
  department_id VARCHAR(50) NOT NULL,
  years_of_experience INT,
  qualifications JSON,
  status VARCHAR(10) NOT NULL DEFAULT 'active',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_department (department_id),
  INDEX idx_specialty (specialty),
  INDEX idx_status (status),
  FOREIGN KEY (department_id) REFERENCES departments(department_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Create appointments table
CREATE TABLE IF NOT EXISTS appointments (
  appointment_id VARCHAR(50) PRIMARY KEY,
  patient_id VARCHAR(50) NOT NULL,
  doctor_id VARCHAR(50) NOT NULL,
  appointment_date DATE NOT NULL,
  start_time TIME NOT NULL,
  end_time TIME NOT NULL,
  reason TEXT NOT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'scheduled',
  notes TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_patient (patient_id),
  INDEX idx_doctor (doctor_id),
  INDEX idx_date (appointment_date),
  INDEX idx_status (status),
  INDEX idx_doctor_date (doctor_id, appointment_date, start_time),
  FOREIGN KEY (patient_id) REFERENCES patients(patient_id),
  FOREIGN KEY (doctor_id) REFERENCES doctors(doctor_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Add foreign key to departments table (must be done after doctors table exists)
ALTER TABLE departments
  ADD CONSTRAINT fk_departments_head_doctor
  FOREIGN KEY (head_doctor_id) REFERENCES doctors(doctor_id);

-- Insert seed data for departments
INSERT INTO departments (department_id, name, description, floor, extension) VALUES
  ('DEPT-001', 'Cardiology', 'Department specializing in heart and cardiovascular conditions', 3, '3001'),
  ('DEPT-002', 'Emergency', 'Emergency medical services and urgent care', 1, '1001'),
  ('DEPT-003', 'Pediatrics', 'Medical care for infants, children, and adolescents', 2, '2001'),
  ('DEPT-004', 'Orthopedics', 'Treatment of musculoskeletal system disorders', 4, '4001'),
  ('DEPT-005', 'Radiology', 'Medical imaging and diagnostic services', 1, '1002'),
  ('DEPT-006', 'Oncology', 'Cancer treatment and care', 5, '5001')
ON DUPLICATE KEY UPDATE name=name;