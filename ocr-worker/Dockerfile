FROM python:3.10-slim

# Bỏ đệm output để log in ra terminal ngay lập tức (Rất quan trọng khi dùng Docker)
ENV PYTHONUNBUFFERED=1

WORKDIR /app

# CHỈ giữ lại poppler-utils để thư viện pdf2image có thể cắt file PDF thành ảnh.
# Đã xóa bỏ hoàn toàn tesseract, thư viện C++ và libgl1 (OpenCV) giúp image siêu nhẹ!
RUN apt-get update && apt-get install -y --no-install-recommends \
    poppler-utils \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements file (thay thế cho hệ thống Poetry cũ)
COPY requirements.txt ./

# Cài đặt thư viện Python cực nhanh và không lưu cache thừa
RUN pip install --no-cache-dir -r requirements.txt

# Copy toàn bộ thư mục code lõi vào container
COPY src/ ./src/

# Chạy trực tiếp file worker bằng Python thuần (Không qua lớp trung gian Poetry nữa)
ENTRYPOINT ["python", "-m", "src.worker"]