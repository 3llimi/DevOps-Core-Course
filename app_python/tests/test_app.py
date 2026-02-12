from fastapi.testclient import TestClient
from app import app

client = TestClient(app)


class TestHomeEndpoint:
    """Tests for the main / endpoint"""

    def test_home_returns_200(self):
        """Test that home endpoint returns HTTP 200 OK"""
        response = client.get("/")
        assert response.status_code == 200

    def test_home_returns_json(self):
        """Test that response is valid JSON"""
        response = client.get("/")
        data = response.json()
        assert isinstance(data, dict)

    def test_home_has_service_info(self):
        """Test that service section exists and has required fields"""
        response = client.get("/")
        data = response.json()

        assert "service" in data
        assert data["service"]["name"] == "devops-info-service"
        assert data["service"]["version"] == "1.0.0"
        assert data["service"]["framework"] == "FastAPI"

    def test_home_has_system_info(self):
        """Test that system section exists and has required fields"""
        response = client.get("/")
        data = response.json()

        assert "system" in data
        assert "hostname" in data["system"]
        assert "platform" in data["system"]
        assert "python_version" in data["system"]

    def test_home_has_runtime_info(self):
        """Test that runtime section exists"""
        response = client.get("/")
        data = response.json()

        assert "runtime" in data
        assert "uptime_seconds" in data["runtime"]
        assert "current_time" in data["runtime"]

    def test_home_has_request_info(self):
        """Test that request section exists"""
        response = client.get("/")
        data = response.json()

        assert "request" in data
        assert "method" in data["request"]
        assert data["request"]["method"] == "GET"


class TestHealthEndpoint:
    """Tests for the /health endpoint"""

    def test_health_returns_200(self):
        """Test that health endpoint returns HTTP 200 OK"""
        response = client.get("/health")
        assert response.status_code == 200

    def test_health_returns_json(self):
        """Test that response is valid JSON"""
        response = client.get("/health")
        data = response.json()
        assert isinstance(data, dict)

    def test_health_has_status(self):
        """Test that health response has status field"""
        response = client.get("/health")
        data = response.json()

        assert "status" in data
        assert data["status"] == "healthy"

    def test_health_has_timestamp(self):
        """Test that health response has timestamp"""
        response = client.get("/health")
        data = response.json()

        assert "timestamp" in data

    def test_health_has_uptime(self):
        """Test that health response has uptime"""
        response = client.get("/health")
        data = response.json()

        assert "uptime_seconds" in data
        assert isinstance(data["uptime_seconds"], int)
