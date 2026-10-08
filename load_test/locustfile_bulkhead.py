"""
Infrastructure Bulkhead and Auto-Scaling Test

Tests:
- Service isolation (External overload doesn't affect Internal)
- Auto-scaling (75% CPU trigger)
- CPU circuit breaker (95% overload, 85% recovery)
- ALB header routing

Both InternalTasks and ExternalTasks perform IDENTICAL operations.
Only difference: X-User-Type header (staff vs public)
This allows side-by-side comparison of service behavior.

Run: locust -f locustfile_bulkhead.py --host=http://your-alb-url --users=200 --spawn-rate=10

Multiple workers:
  # Terminal 1 - Master: locust -f locustfile_bulkhead.py --host=http://your-alb-url --master
  # Terminal 2+ - Workers: locust -f locustfile_bulkhead.py --host=http://your-alb-url --worker
"""

from locust import HttpUser, TaskSet, task
import random
from datetime import datetime, timedelta

TEST_PATIENT_IDS = []
TEST_DOCTOR_IDS = []
TEST_DEPARTMENT_IDS = [
  "DEPT-001",
  "DEPT-002",
  "DEPT-003",
  "DEPT-004",
  "DEPT-005",
  "DEPT-006",
]
TEST_APPOINTMENT_IDS = []


def generate_patient_data():
  return {
    "firstName": f"Patient{random.randint(1000, 9999)}",
    "lastName": f"Test{random.randint(1000, 9999)}",
    "dateOfBirth": "1985-03-15",
    "gender": random.choice(["male", "female"]),
    "email": f"patient{random.randint(10000, 99999)}@test.com",
    "phone": f"+1-555-{random.randint(100, 999)}-{random.randint(1000, 9999)}",
    "bloodType": "O+",
  }


def generate_doctor_data():
  return {
    "firstName": f"Dr{random.randint(1000, 9999)}",
    "lastName": f"Test{random.randint(1000, 9999)}",
    "email": f"doctor{random.randint(10000, 99999)}@hospital.com",
    "phone": f"+1-555-{random.randint(100, 999)}-{random.randint(1000, 9999)}",
    "specialty": random.choice(["Cardiology", "Pediatrics", "Orthopedics"]),
    "licenseNumber": f"MD{random.randint(100000, 999999)}",
    "departmentId": random.choice(TEST_DEPARTMENT_IDS),
    "yearsOfExperience": random.randint(5, 25),
  }


def generate_appointment_data(patient_id, doctor_id):
  date = datetime.now() + timedelta(days=random.randint(1, 30))
  hour = random.randint(8, 16)
  return {
    "patientId": patient_id,
    "doctorId": doctor_id,
    "appointmentDate": date.strftime("%Y-%m-%d"),
    "startTime": f"{hour:02d}:00:00",
    "endTime": f"{hour:02d}:30:00",
    "reason": "Load test",
  }


class InternalTasks(TaskSet):
  """Internal operations with X-User-Type: staff header"""

  @task(10)
  def create_patient(self):
    data = generate_patient_data()
    with self.client.post(
      "/api/v1/patients",
      json=data,
      headers={"X-User-Type": "staff"},
      catch_response=True,
    ) as r:
      if r.status_code == 201:
        TEST_PATIENT_IDS.append(r.json().get("patientId"))
        r.success()
      else:
        r.failure(f"Status {r.status_code}: {r.text}")

  @task(20)
  def list_patients(self):
    params = {"page": random.randint(1, 10), "limit": 20}
    with self.client.get(
      "/api/v1/patients",
      params=params,
      headers={"X-User-Type": "staff"},
      catch_response=True,
    ) as r:
      if r.status_code == 200:
        r.success()
      else:
        r.failure(f"Status {r.status_code}: {r.text}")

  @task(15)
  def get_patient(self):
    if not TEST_PATIENT_IDS:
      return
    with self.client.get(
      f"/api/v1/patients/{random.choice(TEST_PATIENT_IDS)}",
      headers={"X-User-Type": "staff"},
      name="/api/v1/patients/[id]",
      catch_response=True,
    ) as r:
      if r.status_code == 200:
        r.success()
      else:
        r.failure(f"Status {r.status_code}: {r.text}")

  @task(10)
  def create_doctor(self):
    data = generate_doctor_data()
    with self.client.post(
      "/api/v1/doctors",
      json=data,
      headers={"X-User-Type": "staff"},
      catch_response=True,
    ) as r:
      if r.status_code == 201:
        TEST_DOCTOR_IDS.append(r.json().get("doctorId"))
        r.success()
      else:
        r.failure(f"Status {r.status_code}: {r.text}")

  @task(20)
  def list_doctors(self):
    params = {"page": random.randint(1, 10), "limit": 20}
    with self.client.get(
      "/api/v1/doctors",
      params=params,
      headers={"X-User-Type": "staff"},
      catch_response=True,
    ) as r:
      if r.status_code == 200:
        r.success()
      else:
        r.failure(f"Status {r.status_code}: {r.text}")

  @task(15)
  def list_departments(self):
    with self.client.get(
      "/api/v1/departments", headers={"X-User-Type": "staff"}, catch_response=True
    ) as r:
      if r.status_code == 200:
        r.success()
      else:
        r.failure(f"Status {r.status_code}: {r.text}")

  @task(10)
  def create_appointment(self):
    if not TEST_PATIENT_IDS or not TEST_DOCTOR_IDS:
      return
    data = generate_appointment_data(
      random.choice(TEST_PATIENT_IDS), random.choice(TEST_DOCTOR_IDS)
    )
    with self.client.post(
      "/api/v1/appointments",
      json=data,
      headers={"X-User-Type": "staff"},
      catch_response=True,
    ) as r:
      if r.status_code == 201:
        TEST_APPOINTMENT_IDS.append(r.json().get("appointmentId"))
        r.success()
      elif r.status_code == 400:
        r.success()
      else:
        r.failure(f"Status {r.status_code}: {r.text}")

  @task(15)
  def list_appointments(self):
    params = {"page": random.randint(1, 10), "limit": 20}
    with self.client.get(
      "/api/v1/appointments",
      params=params,
      headers={"X-User-Type": "staff"},
      catch_response=True,
    ) as r:
      if r.status_code == 200:
        r.success()
      else:
        r.failure(f"Status {r.status_code}: {r.text}")


class ExternalTasks(TaskSet):
  """External operations with X-User-Type: public header - IDENTICAL to InternalTasks"""

  @task(10)
  def create_patient(self):
    data = generate_patient_data()
    with self.client.post(
      "/api/v1/patients",
      json=data,
      headers={"X-User-Type": "public"},
      catch_response=True,
    ) as r:
      if r.status_code == 201:
        TEST_PATIENT_IDS.append(r.json().get("patientId"))
        r.success()
      else:
        r.failure(f"Status {r.status_code}: {r.text}")

  @task(20)
  def list_patients(self):
    params = {"page": random.randint(1, 10), "limit": 20}
    with self.client.get(
      "/api/v1/patients",
      params=params,
      headers={"X-User-Type": "public"},
      catch_response=True,
    ) as r:
      if r.status_code == 200:
        r.success()
      else:
        r.failure(f"Status {r.status_code}: {r.text}")

  @task(15)
  def get_patient(self):
    if not TEST_PATIENT_IDS:
      return
    with self.client.get(
      f"/api/v1/patients/{random.choice(TEST_PATIENT_IDS)}",
      headers={"X-User-Type": "public"},
      name="/api/v1/patients/[id]",
      catch_response=True,
    ) as r:
      if r.status_code == 200:
        r.success()
      else:
        r.failure(f"Status {r.status_code}: {r.text}")

  @task(10)
  def create_doctor(self):
    data = generate_doctor_data()
    with self.client.post(
      "/api/v1/doctors",
      json=data,
      headers={"X-User-Type": "public"},
      catch_response=True,
    ) as r:
      if r.status_code == 201:
        TEST_DOCTOR_IDS.append(r.json().get("doctorId"))
        r.success()
      else:
        r.failure(f"Status {r.status_code}: {r.text}")

  @task(20)
  def list_doctors(self):
    params = {"page": random.randint(1, 10), "limit": 20}
    with self.client.get(
      "/api/v1/doctors",
      params=params,
      headers={"X-User-Type": "public"},
      catch_response=True,
    ) as r:
      if r.status_code == 200:
        r.success()
      else:
        r.failure(f"Status {r.status_code}: {r.text}")

  @task(15)
  def list_departments(self):
    with self.client.get(
      "/api/v1/departments", headers={"X-User-Type": "public"}, catch_response=True
    ) as r:
      if r.status_code == 200:
        r.success()
      else:
        r.failure(f"Status {r.status_code}: {r.text}")

  @task(10)
  def create_appointment(self):
    if not TEST_PATIENT_IDS or not TEST_DOCTOR_IDS:
      return
    data = generate_appointment_data(
      random.choice(TEST_PATIENT_IDS), random.choice(TEST_DOCTOR_IDS)
    )
    with self.client.post(
      "/api/v1/appointments",
      json=data,
      headers={"X-User-Type": "public"},
      catch_response=True,
    ) as r:
      if r.status_code == 201:
        TEST_APPOINTMENT_IDS.append(r.json().get("appointmentId"))
        r.success()
      elif r.status_code == 400:
        r.success()
      else:
        r.failure(f"Status {r.status_code}: {r.text}")

  @task(15)
  def list_appointments(self):
    params = {"page": random.randint(1, 10), "limit": 20}
    with self.client.get(
      "/api/v1/appointments",
      params=params,
      headers={"X-User-Type": "public"},
      catch_response=True,
    ) as r:
      if r.status_code == 200:
        r.success()
      else:
        r.failure(f"Status {r.status_code}: {r.text}")


class InternalUser(HttpUser):
  """Low-volume staff (weight=1) - Routes to Internal service"""

  tasks = [InternalTasks]
  weight = 1

  def on_start(self):
    for _ in range(5):
      r = self.client.post(
        "/api/v1/patients",
        json=generate_patient_data(),
        headers={"X-User-Type": "staff"},
      )
      if r.status_code == 201:
        TEST_PATIENT_IDS.append(r.json().get("patientId"))
    for _ in range(3):
      r = self.client.post(
        "/api/v1/doctors", json=generate_doctor_data(), headers={"X-User-Type": "staff"}
      )
      if r.status_code == 201:
        TEST_DOCTOR_IDS.append(r.json().get("doctorId"))


class ExternalUser(HttpUser):
  """High-volume public (weight=10) - Routes to External service"""

  tasks = [ExternalTasks]
  weight = 10

  def on_start(self):
    for _ in range(5):
      r = self.client.post(
        "/api/v1/patients",
        json=generate_patient_data(),
        headers={"X-User-Type": "public"},
      )
      if r.status_code == 201:
        TEST_PATIENT_IDS.append(r.json().get("patientId"))
    for _ in range(3):
      r = self.client.post(
        "/api/v1/doctors",
        json=generate_doctor_data(),
        headers={"X-User-Type": "public"},
      )
      if r.status_code == 201:
        TEST_DOCTOR_IDS.append(r.json().get("doctorId"))
