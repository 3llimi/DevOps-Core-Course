# Using Python slim image
FROM python:3.13-slim

# Working directory
WORKDIR /app

# Non-root user for security
RUN groupadd -r appuser && useradd -r -g appuser appuser

# Copying requirements first for better layer caching
COPY requirements.txt .

# Installing dependencies without cache to reduce image size
RUN pip install --no-cache-dir -r requirements.txt

# Copying application code
COPY app.py .

# Changing ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8000

# Runing the application
CMD ["python", "app.py"]

