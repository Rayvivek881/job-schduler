# Deployment Guide

Complete deployment guide for the Job Scheduler system.

## Quick Start

### Docker Compose (Development)

The fastest way to get started locally:

```bash
# 1. Copy environment template
cp deploy/env.example .env

# 2. Edit .env with your configuration (optional for local dev)
# The defaults work with docker-compose.yml

# 3. Start all services
docker-compose up -d

# 4. Verify services are running
docker-compose ps

# 5. Check logs
docker-compose logs -f api-server
docker-compose logs -f scheduler
docker-compose logs -f consumer

# 6. Test health endpoint
curl http://localhost:8000/health

# 7. Run verification script
bash deploy/verify.sh
```

### Stop Services

```bash
docker-compose down
```

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f api-server
docker-compose logs -f scheduler
docker-compose logs -f consumer
```

## Deployment Options

### Option 1: Kubernetes

**Best for**: Production environments, cloud deployments, containerized infrastructure

See [Kubernetes Deployment Guide](k8s/README.md) for complete instructions.

**Quick Steps**:
1. Build and push Docker image
2. Create ConfigMap and Secrets
3. Deploy using `kubectl apply -f deploy/k8s/`

### Option 2: EC2/VMs with Systemd

**Best for**: Traditional server deployments, AWS EC2, on-premise

**Steps**:
1. Build binary: `go build -o job-scheduler .`
2. Copy to server: `/opt/job-scheduler/`
3. Install systemd services: `sudo bash deploy/systemd/install.sh`
4. Create `.env` file: `cp deploy/env.example /opt/job-scheduler/.env`
5. Edit `.env` with production values
6. Enable and start services

See [Systemd Deployment](#systemd-deployment) section below.

### Option 3: Docker Compose (Production)

**Best for**: Single-server deployments, small-scale production

```bash
# 1. Build production image
docker build -t job-scheduler:latest .

# 2. Create .env file with production values
cp deploy/env.example .env
# Edit .env

# 3. Start services
docker-compose up -d

# 4. Verify
curl http://localhost:8000/health
```

## Prerequisites

### Required

- **PostgreSQL 14+**: Database for job storage
- **Apache Kafka 3.6.0+**: Message queue for job distribution
- **Go 1.24+** (if building from source): For compiling the binary

### Optional

- **Docker & Docker Compose**: For containerized deployment
- **Kubernetes**: For orchestrated deployment
- **Nginx**: Reverse proxy for API server

## Configuration

### Environment Variables

All configuration is done via environment variables. See [`deploy/env.example`](env.example) for complete reference.

**Required Variables**:
- `API_KEY`: API authentication key
- `PG_DB_HOST`, `PG_DB_DATABASE`, `PG_DB_USERNAME`, `PG_DB_PASSWORD`: Database connection
- `KAFKA_BROKERS`, `KAFKA_VERSION`: Kafka cluster
- `JOBS_TOPIC`, `CONSUMER_GROUP`: Kafka configuration

**Optional Variables** (with defaults):
- `TICKER_INTERVAL=1`: Scheduler tick interval (minutes)
- `JOB_EXECUTION_TIMEOUT=30`: Job timeout (minutes)
- `MAX_REQUESTS_PER_MINUTE=1000`: API rate limit

### Database Setup

1. **Create Database**:
```sql
CREATE DATABASE job_scheduler;
```

2. **Initialize Schema**:
```bash
psql -U postgres -d job_scheduler -f deploy/sql/init_schema.sql
```

3. **Verify Schema**:
```bash
psql -U postgres -d job_scheduler -c "\dt"
psql -U postgres -d job_scheduler -c "\di"
```

### Kafka Setup

1. **Create Topic** (automated):
```bash
# Using automated script (recommended)
bash deploy/kafka/create-topic.sh \
  --bootstrap-server localhost:9092 \
  --environment development

# Production example
bash deploy/kafka/create-topic.sh \
  --bootstrap-server kafka1:9092,kafka2:9092,kafka3:9092 \
  --environment production \
  --partitions 6 \
  --replication-factor 3
```

2. **Manual Topic Creation**:
```bash
kafka-topics.sh --create \
  --bootstrap-server localhost:9092 \
  --topic jobs-topic \
  --partitions 6 \
  --replication-factor 3
```

3. **Verify Topic**:
```bash
kafka-topics.sh --describe --bootstrap-server localhost:9092 --topic jobs-topic
```

See [Kafka Topic Setup Guide](kafka/TOPIC_SETUP.md) for detailed instructions.

## Systemd Deployment

### Prerequisites

- Ubuntu/Debian system (or systemd-based Linux)
- Binary built and copied to `/opt/job-scheduler/job-scheduler`
- Environment file at `/opt/job-scheduler/.env`

### Installation Steps

1. **Build Binary**:
```bash
go build -o job-scheduler .
```

2. **Copy to Server**:
```bash
scp job-scheduler user@server:/opt/job-scheduler/
scp deploy/env.example user@server:/opt/job-scheduler/.env
```

3. **Install Systemd Services**:
```bash
ssh user@server
sudo bash /opt/job-scheduler/deploy/systemd/install.sh
```

4. **Configure Environment**:
```bash
sudo nano /opt/job-scheduler/.env
# Edit with production values
```

5. **Enable and Start Services**:
```bash
# Enable services to start on boot
sudo systemctl enable job-scheduler-api
sudo systemctl enable job-scheduler-scheduler
sudo systemctl enable job-scheduler-consumer@1

# Start services
sudo systemctl start job-scheduler-api
sudo systemctl start job-scheduler-scheduler
sudo systemctl start job-scheduler-consumer@1
```

6. **Check Status**:
```bash
sudo systemctl status job-scheduler-api
sudo systemctl status job-scheduler-scheduler
sudo systemctl status job-scheduler-consumer@1
```

### Scaling Consumers

To add more consumers (horizontal scaling):

```bash
# Enable and start additional consumer instances
sudo systemctl enable job-scheduler-consumer@2
sudo systemctl start job-scheduler-consumer@2

sudo systemctl enable job-scheduler-consumer@3
sudo systemctl start job-scheduler-consumer@3
```

### Nginx Configuration

If exposing API server via Nginx:

1. **Copy Nginx Config**:
```bash
sudo cp deploy/nginx/job-scheduler.conf /etc/nginx/sites-available/job-scheduler
sudo ln -s /etc/nginx/sites-available/job-scheduler /etc/nginx/sites-enabled/
```

2. **Edit Configuration**:
```bash
sudo nano /etc/nginx/sites-available/job-scheduler
# Update server_name and SSL certificate paths
```

3. **Test and Reload**:
```bash
sudo nginx -t
sudo systemctl reload nginx
```

## Environment Validation

Before deployment, validate your `.env` file:

```bash
# Validate .env file completeness
bash deploy/validate-env.sh

# Or specify a different .env file
bash deploy/validate-env.sh .env.production
```

This script checks:
- All required environment variables are set
- Optional variables are documented
- Kafka SASL configuration is complete (if username is set)

## Verification

### Run Verification Script

After deployment, run the verification script:

```bash
# Local
bash deploy/verify.sh

# Remote (if API is accessible)
API_URL=http://your-api-host:8000 bash deploy/verify.sh
```

### Manual Verification

1. **Health Check**:
```bash
curl http://localhost:8000/health
```

2. **Create Test DAG**:
```bash
curl -X POST http://localhost:8000/jobs/bulk-insert/complete-graph \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '[
    {
      "uuid": "test-job-1",
      "job_title": "Test Job",
      "job_type": "test",
      "data": {"test": true},
      "edges": []
    }
  ]'
```

3. **Check Job Status**:
```bash
curl http://localhost:8000/jobs/test-job-1 \
  -H "X-API-Key: your-api-key"
```

## Troubleshooting

### Service Not Starting

**Check logs**:
```bash
# Systemd
sudo journalctl -u job-scheduler-api -f

# Docker
docker-compose logs api-server

# Kubernetes
kubectl logs -l component=api-server
```

**Common Issues**:
- Database not accessible: Check `PG_DB_HOST` and credentials
- Kafka not accessible: Check `KAFKA_BROKERS` and network connectivity
- Port already in use: Change port or stop conflicting service

### Health Check Failing

**Check components**:
1. Database: `psql -h $PG_DB_HOST -U $PG_DB_USERNAME -d $PG_DB_DATABASE`
2. Kafka: `kafka-broker-api-versions --bootstrap-server $KAFKA_BROKERS`

### Jobs Not Processing

**Check**:
1. Scheduler running: `docker-compose ps scheduler` or `systemctl status job-scheduler-scheduler`
2. Consumer running: `docker-compose ps consumer` or `systemctl status job-scheduler-consumer@1`
3. Kafka consumer lag: See [Kafka Topic Setup Guide](kafka/TOPIC_SETUP.md)

## Scaling

### Horizontal Scaling

**API Server**: Scale to multiple replicas behind load balancer
- Kubernetes: `kubectl scale deployment job-scheduler-api --replicas=3`
- Docker Compose: Add multiple service instances with different ports
- Systemd: Not recommended (use load balancer instead)

**Consumers**: Scale by adding more consumer instances
- Kubernetes: `kubectl scale deployment job-scheduler-consumer --replicas=5`
- Docker Compose: Scale service: `docker-compose up -d --scale consumer=5`
- Systemd: `sudo systemctl enable job-scheduler-consumer@2` (then start)

**Scheduler**: Must remain singleton (1 instance)

### Vertical Scaling

Adjust resource limits in:
- Kubernetes: Deployment resource requests/limits
- Systemd: `MemoryMax` and `LimitNOFILE` in service files
- Docker: `--memory` and `--cpus` flags

## Maintenance

### Updating Configuration

1. Edit ConfigMap/Secrets (K8s) or `.env` file (Systemd/Docker)
2. Restart services to pick up changes

**Kubernetes**:
```bash
kubectl rollout restart deployment job-scheduler-api
```

**Systemd**:
```bash
sudo systemctl restart job-scheduler-api
```

**Docker Compose**:
```bash
docker-compose restart api-server
```

### Database Migrations

Schema is managed via SQL scripts in `deploy/sql/`. To apply updates:

```bash
psql -U postgres -d job_scheduler -f deploy/sql/init_schema.sql
```

### Backup

**Database**:
```bash
pg_dump -U postgres job_scheduler > backup_$(date +%Y%m%d).sql
```

**Kafka Topics**: Use Kafka backup tools or replication

## Security Checklist

- [ ] Change default `API_KEY` to strong random value
- [ ] Use strong database passwords
- [ ] Enable SSL for database (`PG_DB_SSL=true`)
- [ ] Enable Kafka SASL authentication (production)
- [ ] Restrict network access (firewall/security groups)
- [ ] Use HTTPS for API (Nginx/Ingress with SSL)
- [ ] Rotate secrets regularly
- [ ] Keep dependencies updated

## References

- [Configuration Guide](../docs/CONFIGURATION.md)
- [API Documentation](../docs/API.md)
- [Kubernetes Deployment](k8s/README.md)
- [Kafka Setup](kafka/TOPIC_SETUP.md)
- [Monitoring Guide](../docs/MONITORING.md)
