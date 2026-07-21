FROM debian:bookworm-slim

RUN apt-get update && apt-get install python3 python3-venv python3-pip -y

WORKDIR /app
COPY ./requirements.txt .
COPY ./templates ./templates
RUN pip install --no-cache-dir --break-system-packages -r requirements.txt
RUN pip install flask --upgrade --no-cache-dir --break-system-packages

COPY ./app.py .

EXPOSE 5000

CMD ["python3", "app.py"]
