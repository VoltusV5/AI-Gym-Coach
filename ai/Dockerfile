FROM python:3.11-slim
WORKDIR /app
COPY requirements-ml-api.txt .
ARG HOST
ARG PORT
ENV HOST=${HOST} \
    PORT=${PORT}
RUN pip install --no-cache-dir -r requirements-ml-api.txt
COPY . .
EXPOSE 8000
CMD ["sh", "-c", "uvicorn main:app --host $HOST --port $PORT"]