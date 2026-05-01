from locust import HttpUser, task, between

class WarReportUser(HttpUser):
    wait_time = between(1, 2)

    @task
    def send_report(self):
        self.client.post("/", json={
            "country": "Guatemala",
            "warplanesInAir": 10,
            "warshipsInWater": 5
        })
