from fastapi import FastAPI, Request
from datetime import datetime, timezone
from fastapi.responses import JSONResponse
from starlette.exceptions import HTTPException as StarletteHTTPException
import platform
import socket
import os
import logging
import sys

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    handlers=[
        logging.StreamHandler(sys.stdout),
        logging.FileHandler("app.log"),
    ],
)

logger = logging.getLogger(__name__)

app = FastAPI()
START_TIME = datetime.now(timezone.utc)

HOST = os.getenv("HOST", "0.0.0.0")
PORT = int(os.getenv("PORT", 8000))

logger.info(f"Application starting - Host: {HOST}, Port: {PORT}")


def get_uptime():
    delta = datetime.now(timezone.utc) - START_TIME
    secs = int(delta.total_seconds())
    hrs = secs // 3600
    mins = (secs % 3600) // 60
    return {"seconds": secs, "human": f"{hrs} hours, {mins} minutes"}


@app.on_event("startup")
async def startup_event():
    logger.info("FastAPI application startup complete")
    logger.info(f"Python version: {platform.python_version()}")
    logger.info(f"Platform: {platform.system()} {platform.platform()}")
    logger.info(f"Hostname: {socket.gethostname()}")


@app.on_event("shutdown")
async def shutdown_event():
    uptime = get_uptime()
    logger.info(f"Application shutting down. Total uptime: {uptime['human']}")


@app.middleware("http")
async def log_requests(request: Request, call_next):
    start_time = datetime.now(timezone.utc)
    client_ip = request.client.host if request.client else "unknown"

    logger.info(
        f"Request started: {request.method} {request.url.path} "
        f"from {client_ip}"
    )

    try:
        response = await call_next(request)
        process_time = (
            datetime.now(timezone.utc) - start_time
        ).total_seconds()

        logger.info(
            f"Request completed: {request.method} {request.url.path} - "
            f"Status: {response.status_code} - Duration: {process_time:.3f}s"
        )

        response.headers["X-Process-Time"] = str(process_time)
        return response
    except Exception as e:
        process_time = (
            datetime.now(timezone.utc) - start_time
        ).total_seconds()
        logger.error(
            f"Request failed: {request.method} {request.url.path} - "
            f"Error: {str(e)} - Duration: {process_time:.3f}s"
        )
        raise


@app.get("/")
def home(request: Request):
    logger.debug("Home endpoint called")
    uptime = get_uptime()
    return {
        "service": {
            "name": "devops-info-service",
            "version": "1.0.0",
            "description": "DevOps course info service",
            "framework": "FastAPI",
        },
        "system": {
            "hostname": socket.gethostname(),
            "platform": platform.system(),
            "platform_version": platform.platform(),
            "architecture": platform.machine(),
            "cpu_count": os.cpu_count(),
            "python_version": platform.python_version(),
        },
        "runtime": {
            "uptime_seconds": uptime["seconds"],
            "uptime_human": uptime["human"],
            "current_time": datetime.now(timezone.utc).isoformat(),
            "timezone": "UTC",
        },
        "request": {
            "client_ip": request.client.host if request.client else "unknown",
            "user_agent": request.headers.get("user-agent", "unknown"),
            "method": request.method,
            "path": request.url.path,
        },
        "endpoints": [
            {
                "path": "/",
                "method": "GET",
                "description": "Service information",
            },
            {
                "path": "/health",
                "method": "GET",
                "description": "Health check",
            },
        ],
    }


@app.get("/health")
def health():
    logger.debug("Health check endpoint called")
    uptime = get_uptime()
    return {
        "status": "healthy",
        "timestamp": datetime.now(timezone.utc).isoformat(),
        "uptime_seconds": uptime["seconds"],
    }


@app.exception_handler(StarletteHTTPException)
async def http_exception_handler(
    request: Request, exc: StarletteHTTPException
):
    client = request.client.host if request.client else "unknown"
    logger.warning(
        f"HTTP exception: {exc.status_code} - {exc.detail} - "
        f"Path: {request.url.path} - Client: {client}"
    )
    return JSONResponse(
        status_code=exc.status_code,
        content={
            "error": exc.detail,
            "status_code": exc.status_code,
            "path": request.url.path,
        },
    )


@app.exception_handler(Exception)
async def general_exception_handler(request: Request, exc: Exception):
    client = request.client.host if request.client else "unknown"
    logger.error(
        f"Unhandled exception: {type(exc).__name__} - {str(exc)} - "
        f"Path: {request.url.path} - Client: {client}",
        exc_info=True,
    )
    return JSONResponse(
        status_code=500,
        content={
            "error": "Internal Server Error",
            "message": "An unexpected error occurred",
            "path": request.url.path,
        },
    )


if __name__ == "__main__":
    import uvicorn

    logger.info(f"Starting Uvicorn server on {HOST}:{PORT}")
    uvicorn.run(app, host=HOST, port=PORT)
