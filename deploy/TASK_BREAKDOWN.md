# Job Scheduler Deployment Plan - Detailed Task Breakdown

This document provides a detailed breakdown of all tasks from the deployment plan into smaller, actionable subtasks.

## ✅ Completed Tasks (Major Milestones)

### ✅ Phase 1: Infrastructure Setup
- Database schema created
- Kafka topic documentation complete
- Automated Kafka topic creation script created

### ✅ Phase 2: Containerization
- Dockerfile created and tested
- docker-compose.yml configured
- Environment template and validation script created

### ✅ Phase 3: Deployment Configurations
- Kubernetes manifests created
- Systemd services created
- Nginx configuration created

### ✅ Phase 4: CI/CD Pipeline
- Build pipeline complete
- **Deployment pipeline created (NEW)**

### ✅ Phase 5: Monitoring & Observability
- Health checks enhanced (DB + Kafka)
- Logging configuration documented

### ✅ Phase 6: Documentation
- Deployment guide complete
- Operations runbook complete

### ✅ Phase 7: Testing & Validation
- Deployment verification script complete

---

## 📋 Detailed Task Breakdown

### Phase 1: Infrastructure Setup

#### ✅ Task 1.1: Database Schema Creation
**Status**: Complete

**Subtasks**:
- [x] Create `deploy/sql/init_schema.sql` with jobs table
- [x] Create `deploy/sql/init_schema.sql` with edges table
- [x] Add required indexes: `idx_jobs_degree_status_runafter`, `idx_jobs_uuid`, `idx_jobs_status`, `idx_edges_source`, `idx_edges_target`
- [x] Add foreign key constraints
- [x] Document schema in `deploy/sql/README.md`
- [x] Verify schema matches `ModelJobs` and `ModelEdges` structs

**Next Steps** (Optional):
- [ ] Review index usage patterns and optimize
- [ ] Add performance monitoring queries
- [ ] Document schema migration strategy

#### ✅ Task 1.2: Kafka Topic Configuration
**Status**: Complete + Enhanced

**Subtasks**:
- [x] Document topic name configuration (`JOBS_TOPIC`)
- [x] Document partition recommendations (3-6)
- [x] Document replication factor (3 for production)
- [x] Document retention policy
- [x] Document consumer group naming (`CONSUMER_GROUP`)
- [x] Create automated topic creation script (`deploy/kafka/create-topic.sh`)
- [x] Add environment-specific defaults (dev/staging/prod)
- [x] Add dry-run mode support
- [x] Document topic management commands

**Next Steps** (Optional):
- [ ] Add topic validation in deployment script
- [ ] Create topic health check script
- [ ] Add topic monitoring queries

---

### Phase 2: Containerization

#### ✅ Task 2.1: Create Dockerfile
**Status**: Complete

**Subtasks**:
- [x] Create multi-stage Dockerfile
- [x] Configure builder stage with Go 1.24+
- [x] Configure runtime stage with minimal base
- [x] Support all commands: `api-server`, `scheduler first-time`, `scheduler retry`, `consumer`
- [x] Expose port 8000 for API server
- [x] Configure non-root user for security
- [x] Optimize image size

**Next Steps** (Optional):
- [ ] Test multi-architecture builds (ARM64)
- [ ] Optimize layer caching
- [ ] Add security scanning

#### ✅ Task 2.2: Create docker-compose.yml
**Status**: Complete

**Subtasks**:
- [x] Configure API server service
- [x] Configure scheduler service
- [x] Configure consumer service
- [x] Add PostgreSQL service
- [x] Add Kafka service (with Zookeeper)
- [x] Configure environment variables
- [x] Configure networks for service communication
- [x] Add health checks for each service
- [x] Configure volume mounts for SQL init scripts

**Next Steps** (Testing Required):
- [ ] Test all services start correctly
- [ ] Verify service dependencies (postgres → kafka → services)
- [ ] Test health checks work correctly
- [ ] Test network connectivity between services
- [ ] Test environment variable inheritance
- [ ] Test horizontal scaling of consumers

#### ✅ Task 2.3: Environment Template
**Status**: Complete + Enhanced

**Subtasks**:
- [x] Create `deploy/env.example` with all required variables
- [x] Document all required variables from `conf/viper.go`
- [x] Document default values
- [x] Add comments explaining each variable
- [x] Create environment validation script (`deploy/validate-env.sh`)
- [x] Validate required variables
- [x] Validate optional variables
- [x] Validate Kafka SASL configuration (conditional)

**Next Steps** (Optional):
- [ ] Add validation for variable format (e.g., URL validation)
- [ ] Add validation for numeric ranges
- [ ] Integrate validation into CI/CD pipeline

---

### Phase 3: Deployment Configurations

#### ✅ Task 3.1: Kubernetes Manifests
**Status**: Complete

**Subtasks**:
- [x] Create `api-server-deployment.yaml` (2-3 replicas)
- [x] Create `scheduler-deployment.yaml` (1 replica, singleton)
- [x] Create `consumer-deployment.yaml` (2-5 replicas, horizontally scalable)
- [x] Create `configmap.yaml` for non-sensitive configuration
- [x] Create `secrets.yaml` template for sensitive values
- [x] Create `ingress.yaml` for external routing
- [x] Configure resource limits/requests
- [x] Configure liveness and readiness probes
- [x] Document deployment in `deploy/k8s/README.md`

**Next Steps** (Testing Required):
- [ ] Test deployment with `kubectl apply`
- [ ] Verify ConfigMap and Secrets are properly referenced
- [ ] Test horizontal scaling (consumers, API server)
- [ ] Verify ingress routing works
- [ ] Test rolling updates
- [ ] Test resource limits enforcement
- [ ] Test health check probes

#### ✅ Task 3.2: Systemd Services
**Status**: Complete

**Subtasks**:
- [x] Create `job-scheduler-api.service` systemd unit
- [x] Create `job-scheduler-scheduler.service` systemd unit
- [x] Create `job-scheduler-consumer@.service` template unit
- [x] Create `install.sh` installation script
- [x] Configure working directory
- [x] Configure environment file
- [x] Configure restart policy
- [x] Document in deployment guide

**Next Steps** (Testing Required):
- [ ] Test `install.sh` script on clean system
- [ ] Verify service dependencies are correct
- [ ] Test consumer template service with multiple instances
- [ ] Test restart policies (on-failure, always)
- [ ] Test log rotation configuration

#### ✅ Task 3.3: Nginx Configuration
**Status**: Complete

**Subtasks**:
- [x] Create `deploy/nginx/job-scheduler.conf`
- [x] Configure reverse proxy for API server (port 8000)
- [x] Configure health check endpoint routing (`/health`)
- [x] Configure rate limiting (if needed)
- [x] Document SSL/TLS termination setup

**Next Steps** (Testing Required):
- [ ] Test reverse proxy configuration
- [ ] Test health check endpoint routing
- [ ] Test rate limiting configuration
- [ ] Test SSL/TLS termination (if applicable)
- [ ] Test nginx reload procedure

---

### Phase 4: CI/CD Pipeline

#### ✅ Task 4.1: Build Pipeline
**Status**: Complete

**Subtasks**:
- [x] Create `.github/workflows/build.yml`
- [x] Configure build triggers (push, PR, manual)
- [x] Set up Go 1.24+ environment
- [x] Configure dependency caching
- [x] Add test execution step
- [x] Add binary build step
- [x] Configure Docker image build
- [x] Configure image push to registry
- [x] Configure image tagging (commit SHA, latest)

**Next Steps** (Optional):
- [ ] Add security scanning for Docker images
- [ ] Add vulnerability scanning
- [ ] Add build notifications

#### ✅ Task 4.2: Deployment Pipeline
**Status**: Complete (NEW)

**Subtasks**:
- [x] Create `.github/workflows/deploy.yml`
- [x] Configure deployment triggers (tags, manual workflow)
- [x] Add staging deployment job
- [x] Add production deployment job
- [x] Configure staging-first deployment strategy
- [x] Add smoke tests after staging
- [x] Add production deployment gates (conditional on staging success)
- [x] Add kubectl setup
- [x] Add image tag determination logic
- [x] Add deployment manifest updates
- [x] Add ConfigMap deployment
- [x] Add Secrets validation
- [x] Add rolling deployments
- [x] Add port-forwarding for smoke tests
- [x] Add rollback mechanism on failure
- [x] Add deployment notifications (placeholder)

**Next Steps** (Testing Required):
- [ ] Test staging deployment workflow
- [ ] Test smoke tests integration
- [ ] Test production deployment gates
- [ ] Test rollback mechanism
- [ ] Configure Kubernetes secrets in GitHub
- [ ] Configure API keys for staging/production
- [ ] Test manual workflow dispatch
- [ ] Test tag-based deployment

---

### Phase 5: Monitoring & Observability

#### ✅ Task 5.1: Health Check Enhancement
**Status**: Complete

**Subtasks**:
- [x] Verify database connectivity check (`PgDBConnection.Client.Ping()`)
- [x] Verify Kafka connectivity check (`KafkaService.Client().Brokers()`)
- [x] Verify structured JSON response with component status
- [x] Verify proper HTTP status codes (200 OK, 503 Service Unavailable)
- [x] Verify degraded status handling
- [x] Document health check in API docs

**Next Steps** (Optional):
- [ ] Add more detailed health metrics (connection pool usage, etc.)
- [ ] Add health check caching (optional performance improvement)
- [ ] Add health check endpoint metrics

#### ✅ Task 5.2: Logging Configuration
**Status**: Complete

**Subtasks**:
- [x] Verify zerolog usage throughout codebase
- [x] Verify structured logging format
- [x] Document log levels per environment
- [x] Document log aggregation setup (CloudWatch, ELK)
- [x] Document in `docs/MONITORING.md`

**Next Steps** (Optional):
- [ ] Configure log aggregation in production
- [ ] Set up log retention policies
- [ ] Add log parsing patterns

#### ⚠️ Task 5.3: Metrics Collection
**Status**: Not Implemented (Optional/Future Enhancement)

**Subtasks** (Future):
- [ ] Add Prometheus client library
- [ ] Implement API server metrics:
  - [ ] Request rate (`http_requests_total`)
  - [ ] Request latency (`http_request_duration_seconds`)
  - [ ] Error rate
- [ ] Implement job metrics:
  - [ ] Job status counts (`job_status_total`)
  - [ ] Job processing duration
  - [ ] Jobs processed (`jobs_processed_total`)
- [ ] Implement Kafka metrics:
  - [ ] Consumer lag (`kafka_consumer_lag`)
  - [ ] Messages consumed
- [ ] Implement database metrics:
  - [ ] Connection pool usage
  - [ ] Query duration
- [ ] Add `/metrics` endpoint
- [ ] Create Grafana dashboard configuration
- [ ] Update `docs/MONITORING.md` with metrics

---

### Phase 6: Documentation

#### ✅ Task 6.1: Deployment Guide
**Status**: Complete + Enhanced

**Subtasks**:
- [x] Create `deploy/README.md`
- [x] Add quick start guide using docker-compose
- [x] Add Kubernetes deployment instructions
- [x] Add EC2/systemd deployment instructions
- [x] Add environment variable reference
- [x] Add troubleshooting section
- [x] Add environment validation instructions
- [x] Add Kafka topic setup instructions
- [x] Update with new scripts

**Next Steps** (Optional):
- [ ] Add deployment diagrams
- [ ] Add video tutorials
- [ ] Add more troubleshooting scenarios

#### ✅ Task 6.2: Operational Runbook
**Status**: Complete

**Subtasks**:
- [x] Create `docs/OPERATIONS.md`
- [x] Document service restart procedures
- [x] Document scaling consumer instances
- [x] Document database backup and restore
- [x] Document Kafka topic management
- [x] Document common error scenarios and solutions

**Next Steps** (Optional):
- [ ] Add more emergency procedures
- [ ] Add incident response templates
- [ ] Add post-incident review templates

---

### Phase 7: Testing & Validation

#### ✅ Task 7.1: Deployment Verification Script
**Status**: Complete

**Subtasks**:
- [x] Create `deploy/verify.sh`
- [x] Add health check test
- [x] Add database connectivity test (from health response)
- [x] Add Kafka connectivity test (from health response)
- [x] Add DAG creation test
- [x] Add job retrieval test
- [x] Add DAG status test
- [x] Add retry functionality test
- [x] Add color-coded output
- [x] Add exit codes for CI/CD integration

**Next Steps** (Testing Required):
- [ ] Run script against local deployment
- [ ] Run script against staging deployment
- [ ] Run script against production deployment
- [ ] Verify all test cases pass
- [ ] Add performance benchmarks (optional)

#### ⚠️ Task 7.2: Load Testing
**Status**: Partially Complete (Documentation Only)

**Subtasks**:
- [x] Document load testing approach in `docs/LOAD_TESTING.md`
- [ ] Create load testing script (k6, locust, or similar)
- [ ] Test API endpoint capacity
- [ ] Test concurrent job processing
- [ ] Test consumer scalability with multiple instances
- [ ] Document load test results and recommendations
- [ ] Add load test to CI/CD pipeline (optional)

---

## 🎯 Priority Summary

### Critical Tasks (Before Production)
- ✅ **Task 4.2**: Deployment pipeline — **COMPLETE**
- ⚠️ **Task 7.1**: Verify deployment script — **Testing Required**
- ⚠️ **Task 3.1**: Test K8s deployment — **Testing Required**
- ⚠️ **Task 2.2**: Test docker-compose — **Testing Required**

### High Priority Tasks (Before Production)
- ✅ **Task 1.2**: Kafka topic configuration — **COMPLETE**
- ⚠️ **Task 3.2**: Test systemd deployment — **Testing Required**
- ✅ **Task 5.1**: Health checks — **COMPLETE**
- ✅ **Task 6.1**: Deployment guide — **COMPLETE**

### Medium Priority Tasks (Post-MVP)
- ⚠️ **Task 5.3**: Metrics collection — **Optional/Future**
- ⚠️ **Task 7.2**: Load testing — **Scripts Needed**

### Low Priority Tasks (Nice to Have)
- ⚠️ **Task 2.3**: Environment validation enhancements — **Optional**
- ⚠️ **Task 3.3**: Nginx SSL/TLS testing — **Optional**
- ⚠️ **Task 5.2**: Enhanced log aggregation — **Optional**

---

## 📊 Overall Status

- **Total Tasks**: 7 major tasks
- **Completed**: 5 tasks (71%)
- **Testing Required**: 3 tasks (43%)
- **Optional/Future**: 2 tasks (29%)

**Production Ready**: Yes, after testing critical tasks.

---

## 🚀 Immediate Next Steps

1. **Test docker-compose** (Task 2.2)
   - Start all services
   - Verify health checks
   - Test job processing end-to-end

2. **Test Kubernetes deployment** (Task 3.1)
   - Deploy to staging cluster
   - Verify all pods start correctly
   - Test scaling operations

3. **Test deployment pipeline** (Task 4.2)
   - Run staging deployment
   - Verify smoke tests
   - Test production gates

4. **Run verification script** (Task 7.1)
   - Test against all environments
   - Verify all test cases pass

---

## 📚 Reference Documents

- [Deployment Plan](../plans/job_scheduler_deployment_plan_18326e9c.plan.md) (Original)
- [Deployment Tasks Completed](DEPLOYMENT_TASKS_COMPLETED.md) (Summary)
- [Deployment Guide](README.md) (Complete Guide)
- [Kafka Topic Setup](kafka/TOPIC_SETUP.md)
- [Kubernetes Deployment](k8s/README.md)
- [Operations Runbook](../docs/OPERATIONS.md)
- [Monitoring Guide](../docs/MONITORING.md)
