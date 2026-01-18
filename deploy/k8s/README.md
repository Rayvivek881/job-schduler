# Kubernetes Deployment

This directory contains Kubernetes manifests for deploying the Job Scheduler system.

## Files

- `api-server-deployment.yaml` - API Server deployment and service
- `scheduler-deployment.yaml` - Scheduler deployment (singleton)
- `consumer-deployment.yaml` - Consumer deployment (horizontally scalable)
- `configmap.yaml` - Non-sensitive configuration
- `secrets.yaml` - Sensitive credentials (template)
- `ingress.yaml` - Ingress configuration for API server

## Prerequisites

1. Kubernetes cluster (1.20+)
2. kubectl configured and connected to cluster
3. Image registry accessible from cluster (or load image to cluster nodes)
4. PostgreSQL database accessible from cluster
5. Kafka cluster accessible from cluster

## Quick Start

### 1. Create Secrets

**Important**: Do not commit secrets to Git. Create secrets using one of these methods:

```bash
# Method 1: From literals (recommended)
kubectl create secret generic job-scheduler-secrets \
  --from-literal=API_KEY='your-production-api-key' \
  --from-literal=PG_DB_PASSWORD='your-database-password' \
  --from-literal=KAFKA_USERNAME='kafka-user' \
  --from-literal=KAFKA_PASSWORD='kafka-password' \
  --from-literal=KAFKA_SASL_MECHANISM='SCRAM-SHA-256'

# Method 2: From .env file
kubectl create secret generic job-scheduler-secrets \
  --from-env-file=.env.production
```

### 2. Update ConfigMap

Edit `configmap.yaml` with your configuration values:

```bash
# Edit configmap.yaml, then apply
kubectl apply -f deploy/k8s/configmap.yaml
```

### 3. Build and Push Docker Image

```bash
# Build image
docker build -t job-scheduler:latest .

# Tag for your registry
docker tag job-scheduler:latest your-registry.io/job-scheduler:v1.0.0

# Push to registry
docker push your-registry.io/job-scheduler:v1.0.0

# Update image in deployment files or use imagePullSecrets
```

### 4. Update Image Reference

Edit deployment files to use your image:

```yaml
# In each deployment file, update:
image: your-registry.io/job-scheduler:v1.0.0
```

### 5. Deploy All Components

```bash
# Deploy in order
kubectl apply -f deploy/k8s/configmap.yaml
kubectl apply -f deploy/k8s/secrets.yaml
kubectl apply -f deploy/k8s/api-server-deployment.yaml
kubectl apply -f deploy/k8s/scheduler-deployment.yaml
kubectl apply -f deploy/k8s/consumer-deployment.yaml
kubectl apply -f deploy/k8s/ingress.yaml
```

### 6. Verify Deployment

```bash
# Check pods
kubectl get pods -l app=job-scheduler

# Check services
kubectl get svc -l app=job-scheduler

# Check logs
kubectl logs -l component=api-server
kubectl logs -l component=scheduler
kubectl logs -l component=consumer

# Test health endpoint
kubectl port-forward svc/job-scheduler-api 8000:8000
curl http://localhost:8000/health
```

## Scaling

### Scale API Server

```bash
kubectl scale deployment job-scheduler-api --replicas=3
```

### Scale Consumers

```bash
# Scale to 5 consumers for higher throughput
kubectl scale deployment job-scheduler-consumer --replicas=5
```

**Note**: Ensure Kafka topic has enough partitions (at least equal to consumer count).

### Scheduler

**DO NOT scale scheduler** - it must remain a singleton (1 replica) to avoid duplicate job scheduling.

## Updating Configuration

### Update ConfigMap

```bash
# Edit configmap.yaml
kubectl apply -f deploy/k8s/configmap.yaml

# Restart pods to pick up changes
kubectl rollout restart deployment job-scheduler-api
kubectl rollout restart deployment job-scheduler-scheduler
kubectl rollout restart deployment job-scheduler-consumer
```

### Update Secrets

```bash
# Update secret
kubectl create secret generic job-scheduler-secrets \
  --from-literal=API_KEY='new-api-key' \
  --dry-run=client -o yaml | kubectl apply -f -

# Restart pods
kubectl rollout restart deployment job-scheduler-api
kubectl rollout restart deployment job-scheduler-scheduler
kubectl rollout restart deployment job-scheduler-consumer
```

## Monitoring

### View Pod Logs

```bash
# All components
kubectl logs -l app=job-scheduler --tail=100

# Specific component
kubectl logs -l component=api-server --tail=100 -f
kubectl logs -l component=scheduler --tail=100 -f
kubectl logs -l component=consumer --tail=100 -f
```

### Check Resource Usage

```bash
kubectl top pods -l app=job-scheduler
kubectl top nodes
```

### Describe Resources

```bash
kubectl describe deployment job-scheduler-api
kubectl describe pod <pod-name>
```

## Troubleshooting

### Pods Not Starting

```bash
# Check pod status
kubectl get pods -l app=job-scheduler

# Describe pod for events
kubectl describe pod <pod-name>

# Check logs
kubectl logs <pod-name>
```

### Health Check Failures

```bash
# Check health endpoint manually
kubectl port-forward svc/job-scheduler-api 8000:8000
curl http://localhost:8000/health
```

### Connection Issues

Verify ConfigMap and Secrets:

```bash
kubectl get configmap job-scheduler-config -o yaml
kubectl get secret job-scheduler-secrets -o yaml
```

### Image Pull Errors

Ensure:
1. Image is pushed to registry
2. ImagePullSecrets configured (if using private registry)
3. Cluster nodes can access registry

## Cleanup

```bash
# Delete all resources
kubectl delete -f deploy/k8s/

# Or delete individually
kubectl delete deployment job-scheduler-api
kubectl delete deployment job-scheduler-scheduler
kubectl delete deployment job-scheduler-consumer
kubectl delete service job-scheduler-api
kubectl delete ingress job-scheduler-ingress
kubectl delete configmap job-scheduler-config
kubectl delete secret job-scheduler-secrets
```
