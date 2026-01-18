# Configuration Guide

Complete configuration reference for the Job Scheduler service.

## Environment Variables

Create a `.env` file in the project root or set environment variables:

### Application Configuration

```env
# Application Environment
APP_ENV=development  # or: production, staging

# API Authentication
API_KEY=your-secret-api-key

# Rate Limiting
MAX_REQUESTS_PER_MINUTE=1000  # Default: 1000 requests per minute
```

### PostgreSQL Configuration

```env
# PostgreSQL Connection
PG_DB_HOST=localhost
PG_DB_PORT=5432
PG_DB_DATABASE=job_scheduler
PG_DB_USERNAME=postgres
PG_DB_PASSWORD=postgres
PG_DB_DEBUG=true  # Enable SQL query logging
PG_DB_SSL=false   # Enable SSL connection
```

### Kafka Configuration

```env
# Kafka Brokers (comma-separated)
KAFKA_BROKERS=localhost:9092  # or: broker1:9092,broker2:9092,broker3:9092

# Kafka Version
KAFKA_VERSION=3.6.0

# Kafka Authentication (optional)
KAFKA_USERNAME=your-kafka-username
KAFKA_PASSWORD=your-kafka-password
KAFKA_SASL_MECHANISM=PLAIN  # or: SCRAM-SHA-256, SCRAM-SHA-512

# Kafka Topics
JOBS_TOPIC=jobs-topic
CONSUMER_GROUP=jobs-consumer-group
```

### Job Configuration

```env
# Scheduler Configuration
TICKER_INTERVAL=1  # Interval in minutes for scheduler to check for ready jobs

# Job Execution Timeout
JOB_EXECUTION_TIMEOUT=30  # Timeout in minutes (default: 30)

# Job Queue Configuration
JOB_IN_QUEUE_SIZE=1000  # Maximum jobs in queue (optional)
PARALLEL_JOBS=10  # Maximum parallel jobs (optional)
```

## Configuration Reference

### Application Settings

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `APP_ENV` | string | - | Application environment (development/production/staging) |
| `API_KEY` | string | - | **Required**. API key for authentication |
| `MAX_REQUESTS_PER_MINUTE` | int | 1000 | Rate limit (requests per minute) |

### Database Settings

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `PG_DB_HOST` | string | - | **Required**. PostgreSQL host |
| `PG_DB_PORT` | string | 5432 | PostgreSQL port |
| `PG_DB_DATABASE` | string | - | **Required**. Database name |
| `PG_DB_USERNAME` | string | - | **Required**. Database username |
| `PG_DB_PASSWORD` | string | - | **Required**. Database password |
| `PG_DB_DEBUG` | bool | false | Enable SQL query logging |
| `PG_DB_SSL` | bool | false | Enable SSL connection |

### Kafka Settings

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `KAFKA_BROKERS` | string[] | - | **Required**. Kafka broker addresses (comma-separated) |
| `KAFKA_VERSION` | string | - | **Required**. Kafka version (e.g., "3.6.0") |
| `KAFKA_USERNAME` | string | "" | Kafka SASL username (optional) |
| `KAFKA_PASSWORD` | string | "" | Kafka SASL password (optional) |
| `KAFKA_SASL_MECHANISM` | string | "" | SASL mechanism (PLAIN, SCRAM-SHA-256, etc.) |
| `JOBS_TOPIC` | string | - | **Required**. Kafka topic for jobs |
| `CONSUMER_GROUP` | string | - | **Required**. Consumer group ID |

### Job Settings

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `TICKER_INTERVAL` | int | 1 | Scheduler tick interval (minutes) |
| `JOB_EXECUTION_TIMEOUT` | int | 30 | Job execution timeout (minutes) |
| `JOB_IN_QUEUE_SIZE` | int | 1000 | Maximum jobs in queue (optional) |
| `PARALLEL_JOBS` | int | 10 | Maximum parallel jobs (optional) |

## Example Configuration

### Development Environment

```env
# .env.development
APP_ENV=development
API_KEY=dev-api-key-12345
MAX_REQUESTS_PER_MINUTE=1000

PG_DB_HOST=localhost
PG_DB_PORT=5432
PG_DB_DATABASE=job_scheduler_dev
PG_DB_USERNAME=postgres
PG_DB_PASSWORD=postgres
PG_DB_DEBUG=true
PG_DB_SSL=false

KAFKA_BROKERS=localhost:9092
KAFKA_VERSION=3.6.0
JOBS_TOPIC=jobs-topic-dev
CONSUMER_GROUP=jobs-consumer-group-dev

TICKER_INTERVAL=1
JOB_EXECUTION_TIMEOUT=30
```

### Production Environment

```env
# .env.production
APP_ENV=production
API_KEY=production-secret-api-key-change-me
MAX_REQUESTS_PER_MINUTE=5000

PG_DB_HOST=prod-db.example.com
PG_DB_PORT=5432
PG_DB_DATABASE=job_scheduler_prod
PG_DB_USERNAME=prod_user
PG_DB_PASSWORD=secure-password-here
PG_DB_DEBUG=false
PG_DB_SSL=true

KAFKA_BROKERS=kafka1.example.com:9092,kafka2.example.com:9092,kafka3.example.com:9092
KAFKA_VERSION=3.6.0
KAFKA_USERNAME=kafka_user
KAFKA_PASSWORD=kafka_password
KAFKA_SASL_MECHANISM=SCRAM-SHA-256
JOBS_TOPIC=jobs-topic-prod
CONSUMER_GROUP=jobs-consumer-group-prod

TICKER_INTERVAL=1
JOB_EXECUTION_TIMEOUT=60
PARALLEL_JOBS=50
```

## Configuration Loading

The application uses [Viper](https://github.com/spf13/viper) for configuration management:

1. **Environment Variables**: Automatically loaded via `viper.AutomaticEnv()`
2. **`.env` File**: Loaded from project root (if exists)
3. **Defaults**: Set in `conf/viper.go`

### Priority Order

1. Environment variables (highest priority)
2. `.env` file
3. Default values (lowest priority)

## Configuration Validation

The application validates configuration on startup:

- **Required Variables**: Must be set or application will fail to start
- **Type Validation**: Automatic type conversion with validation
- **Default Values**: Sensible defaults for optional variables

### Required Variables

- `API_KEY`
- `PG_DB_HOST`
- `PG_DB_DATABASE`
- `PG_DB_USERNAME`
- `PG_DB_PASSWORD`
- `KAFKA_BROKERS`
- `KAFKA_VERSION`
- `JOBS_TOPIC`
- `CONSUMER_GROUP`

## Security Best Practices

1. **API Key**: Use strong, randomly generated API keys in production
2. **Database Password**: Store in secure vault or environment variables (never in code)
3. **Kafka Credentials**: Use SASL authentication in production
4. **SSL/TLS**: Enable `PG_DB_SSL=true` for database connections in production
5. **Environment Variables**: Use `.env` files locally, environment variables in production

## Troubleshooting

### Connection Issues

**PostgreSQL Connection Failed**:
- Verify `PG_DB_HOST`, `PG_DB_PORT`, `PG_DB_DATABASE` are correct
- Check database is running and accessible
- Verify credentials in `PG_DB_USERNAME` and `PG_DB_PASSWORD`
- Check firewall rules if connecting remotely

**Kafka Connection Failed**:
- Verify `KAFKA_BROKERS` is correct (comma-separated if multiple)
- Check Kafka is running and accessible
- Verify `KAFKA_VERSION` matches Kafka server version
- Check authentication if using SASL

### Configuration Not Loading

**Environment Variables Not Recognized**:
- Ensure variable names match exactly (case-sensitive)
- Check `.env` file is in project root
- Verify environment variables are exported (if using shell)

**Default Values Not Applied**:
- Check `conf/viper.go` for default values
- Verify configuration unmarshaling is working (check logs)

---

## See Also

- [API Documentation](./API.md) - API endpoints and usage
- [DAG Guide](./DAG_GUIDE.md) - Creating and managing DAGs
- [README.md](../README.md) - Project overview
