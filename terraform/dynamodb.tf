# DynamoDB Table for Medical Records
resource "aws_dynamodb_table" "medical_records" {
  name         = var.medical_records_table
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "recordId"

  attribute {
    name = "recordId"
    type = "S"
  }

  attribute {
    name = "patientId"
    type = "S"
  }

  attribute {
    name = "visitDate"
    type = "S"
  }

  # Global Secondary Index for querying by patientId
  global_secondary_index {
    name            = "PatientIdIndex"
    hash_key        = "patientId"
    range_key       = "visitDate"
    projection_type = "ALL"
  }

  point_in_time_recovery {
    enabled = true
  }

  server_side_encryption {
    enabled = true
  }

  tags = {
    Name = "${var.project_name}-medical-records"
  }
}

# DynamoDB Table for Invoices
resource "aws_dynamodb_table" "invoices" {
  name         = var.invoices_table
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "invoiceId"

  attribute {
    name = "invoiceId"
    type = "S"
  }

  attribute {
    name = "patientId"
    type = "S"
  }

  attribute {
    name = "invoiceDate"
    type = "S"
  }

  # Global Secondary Index for querying by patientId
  global_secondary_index {
    name            = "PatientIdIndex"
    hash_key        = "patientId"
    range_key       = "invoiceDate"
    projection_type = "ALL"
  }

  point_in_time_recovery {
    enabled = true
  }

  server_side_encryption {
    enabled = true
  }

  tags = {
    Name = "${var.project_name}-invoices"
  }
}
