"""
Database Circuit Breaker and Connection Pool Test

SETUP: Reduce MySQL connection pool to 5 in main.go:
  mysqlDB.SetMaxOpenConns(5)

Tests:
- MySQL connection pool exhaustion
- MySQL circuit breaker (60% failure threshold)
- DynamoDB throttling and circuit breaker
- Database recovery behavior

Run:
  # MySQL only:  locust -f locustfile_database.py --users=50 --spawn-rate=10 MySQLStressUser
  # DynamoDB only: locust -f locustfile_database.py --users=50 --spawn-rate=10 DynamoDBStressUser
  # Both: locust -f locustfile_database.py --users=100 --spawn-rate=10
"""

from locust import HttpUser, TaskSet, task
import random
from datetime import datetime, timedelta

TEST_PATIENT_IDS = []
TEST_DOCTOR_IDS = []
TEST_DEPARTMENT_IDS = ["DEPT-001", "DEPT-002", "DEPT-003"]
TEST_RECORD_IDS = []
TEST_INVOICE_IDS = []


def gen_patient():
  return {
    "firstName": f"P{random.randint(1000, 9999)}",
    "lastName": f"T{random.randint(1000, 9999)}",
    "dateOfBirth": "1985-03-15",
    "gender": "male",
    "email": f"p{random.randint(10000, 99999)}@test.com",
    "phone": f"+1-555-{random.randint(100, 999)}-{random.randint(1000, 9999)}",
    "bloodType": "O+",
  }


def gen_doctor():
  return {
    "firstName": f"D{random.randint(1000, 9999)}",
    "lastName": f"T{random.randint(1000, 9999)}",
    "email": f"d{random.randint(10000, 99999)}@test.com",
    "phone": f"+1-555-{random.randint(100, 999)}-{random.randint(1000, 9999)}",
    "specialty": random.choice(["Cardiology", "Pediatrics"]),
    "licenseNumber": f"MD{random.randint(100000, 999999)}",
    "departmentId": random.choice(TEST_DEPARTMENT_IDS),
  }


def gen_appointment(pid, did):
  date = datetime.now() + timedelta(days=random.randint(1, 30))
  h = random.randint(8, 16)
  return {
    "patientId": pid,
    "doctorId": did,
    "appointmentDate": date.strftime("%Y-%m-%d"),
    "startTime": f"{h:02d}:00:00",
    "endTime": f"{h:02d}:30:00",
    "reason": "Test",
  }


def gen_record(pid, did):
  return {
    "patientId": pid,
    "doctorId": did,
    "visitDate": datetime.now().strftime("%Y-%m-%dT%H:%M:%SZ"),
    "chiefComplaint": "Test",
    "diagnosis": "Test",
    "vitalSigns": {"bloodPressure": "120/80", "heartRate": 72},
  }


def gen_invoice(pid):
  today = datetime.now().strftime("%Y-%m-%d")
  due = (datetime.now() + timedelta(days=30)).strftime("%Y-%m-%d")
  return {
    "patientId": pid,
    "invoiceDate": today,
    "dueDate": due,
    "items": [
      {"description": "Test", "quantity": 1, "unitPrice": 100.0, "amount": 100.0}
    ],
  }


class MySQLTasks(TaskSet):
  """MySQL operations - connection pool stress"""

  @task(15)
  def create_patient(self):
    with self.client.post(
      "/api/v1/patients",
      json=gen_patient(),
      headers={"X-User-Type": "public"},
      catch_response=True,
    ) as r:
      if r.status_code == 201:
        TEST_PATIENT_IDS.append(r.json().get("patientId"))
        r.success()
      elif r.status_code in [500, 503]:
        r.success()

  @task(25)
  def list_patients(self):
    with self.client.get(
      f"/api/v1/patients?page={random.randint(1, 20)}&limit=50",
      headers={"X-User-Type": "public"},
      catch_response=True,
    ) as r:
      if r.status_code in [200, 500, 503]:
        r.success()

  @task(20)
  def get_patient(self):
    if not TEST_PATIENT_IDS:
      return
    with self.client.get(
      f"/api/v1/patients/{random.choice(TEST_PATIENT_IDS)}",
      headers={"X-User-Type": "public"},
      name="/api/v1/patients/[id]",
      catch_response=True,
    ) as r:
      if r.status_code in [200, 404, 500, 503]:
        r.success()

  @task(12)
  def create_doctor(self):
    with self.client.post(
      "/api/v1/doctors",
      json=gen_doctor(),
      headers={"X-User-Type": "staff"},
      catch_response=True,
    ) as r:
      if r.status_code == 201:
        TEST_DOCTOR_IDS.append(r.json().get("doctorId"))
        r.success()
      elif r.status_code in [500, 503]:
        r.success()

  @task(20)
  def list_doctors(self):
    with self.client.get(
      f"/api/v1/doctors?page={random.randint(1, 20)}&limit=50",
      headers={"X-User-Type": "public"},
      catch_response=True,
    ) as r:
      if r.status_code in [200, 500, 503]:
        r.success()

  @task(15)
  def create_appointment(self):
    if not TEST_PATIENT_IDS or not TEST_DOCTOR_IDS:
      return
    data = gen_appointment(
      random.choice(TEST_PATIENT_IDS), random.choice(TEST_DOCTOR_IDS)
    )
    with self.client.post(
      "/api/v1/appointments",
      json=data,
      headers={"X-User-Type": "public"},
      catch_response=True,
    ) as r:
      if r.status_code in [201, 400, 500, 503]:
        r.success()

  @task(25)
  def list_appointments(self):
    with self.client.get(
      f"/api/v1/appointments?page={random.randint(1, 20)}&limit=50",
      headers={"X-User-Type": "public"},
      catch_response=True,
    ) as r:
      if r.status_code in [200, 500, 503]:
        r.success()


class DynamoDBTasks(TaskSet):
  """DynamoDB operations - throttling stress"""

  @task(20)
  def create_medical_record(self):
    if not TEST_PATIENT_IDS or not TEST_DOCTOR_IDS:
      return
    data = gen_record(random.choice(TEST_PATIENT_IDS), random.choice(TEST_DOCTOR_IDS))
    with self.client.post(
      "/api/v1/medical-records",
      json=data,
      headers={"X-User-Type": "staff"},
      catch_response=True,
    ) as r:
      if r.status_code == 201:
        TEST_RECORD_IDS.append(r.json().get("recordId"))
        r.success()
      elif r.status_code in [404, 500, 503]:
        r.success()

  @task(15)
  def get_medical_record(self):
    if not TEST_RECORD_IDS:
      return
    with self.client.get(
      f"/api/v1/medical-records/{random.choice(TEST_RECORD_IDS)}",
      headers={"X-User-Type": "staff"},
      name="/api/v1/medical-records/[id]",
      catch_response=True,
    ) as r:
      if r.status_code in [200, 404, 500, 503]:
        r.success()

  @task(18)
  def get_patient_records(self):
    if not TEST_PATIENT_IDS:
      return
    with self.client.get(
      f"/api/v1/patients/{random.choice(TEST_PATIENT_IDS)}/medical-records",
      headers={"X-User-Type": "staff"},
      name="/api/v1/patients/[id]/medical-records",
      catch_response=True,
    ) as r:
      if r.status_code in [200, 404, 500, 503]:
        r.success()

  @task(12)
  def update_medical_record(self):
    if not TEST_RECORD_IDS:
      return
    data = {
      "diagnosis": f"Updated {random.randint(1000, 9999)}",
      "treatment": "Updated",
    }
    with self.client.put(
      f"/api/v1/medical-records/{random.choice(TEST_RECORD_IDS)}",
      json=data,
      headers={"X-User-Type": "staff"},
      name="/api/v1/medical-records/[id]",
      catch_response=True,
    ) as r:
      if r.status_code in [200, 404, 500, 503]:
        r.success()

  @task(20)
  def create_invoice(self):
    if not TEST_PATIENT_IDS:
      return
    data = gen_invoice(random.choice(TEST_PATIENT_IDS))
    with self.client.post(
      "/api/v1/billing/invoices",
      json=data,
      headers={"X-User-Type": "staff"},
      catch_response=True,
    ) as r:
      if r.status_code == 201:
        TEST_INVOICE_IDS.append(r.json().get("invoiceId"))
        r.success()
      elif r.status_code in [404, 500, 503]:
        r.success()

  @task(15)
  def get_invoice(self):
    if not TEST_INVOICE_IDS:
      return
    with self.client.get(
      f"/api/v1/billing/invoices/{random.choice(TEST_INVOICE_IDS)}",
      headers={"X-User-Type": "staff"},
      name="/api/v1/billing/invoices/[id]",
      catch_response=True,
    ) as r:
      if r.status_code in [200, 404, 500, 503]:
        r.success()

  @task(18)
  def get_patient_invoices(self):
    if not TEST_PATIENT_IDS:
      return
    with self.client.get(
      f"/api/v1/patients/{random.choice(TEST_PATIENT_IDS)}/invoices",
      headers={"X-User-Type": "staff"},
      name="/api/v1/patients/[id]/invoices",
      catch_response=True,
    ) as r:
      if r.status_code in [200, 404, 500, 503]:
        r.success()


class MySQLStressUser(HttpUser):
  """MySQL connection pool stress - exhausts 5-connection pool"""

  tasks = [MySQLTasks]
  weight = 1

  def on_start(self):
    for _ in range(10):
      r = self.client.post(
        "/api/v1/patients", json=gen_patient(), headers={"X-User-Type": "public"}
      )
      if r.status_code == 201:
        TEST_PATIENT_IDS.append(r.json().get("patientId"))
    for _ in range(5):
      r = self.client.post(
        "/api/v1/doctors", json=gen_doctor(), headers={"X-User-Type": "staff"}
      )
      if r.status_code == 201:
        TEST_DOCTOR_IDS.append(r.json().get("doctorId"))


class DynamoDBStressUser(HttpUser):
  """DynamoDB throttling stress - high write volume"""

  tasks = [DynamoDBTasks]
  weight = 1

  def on_start(self):
    for _ in range(10):
      r = self.client.post(
        "/api/v1/patients", json=gen_patient(), headers={"X-User-Type": "public"}
      )
      if r.status_code == 201:
        TEST_PATIENT_IDS.append(r.json().get("patientId"))
    for _ in range(5):
      r = self.client.post(
        "/api/v1/doctors", json=gen_doctor(), headers={"X-User-Type": "staff"}
      )
      if r.status_code == 201:
        TEST_DOCTOR_IDS.append(r.json().get("doctorId"))
