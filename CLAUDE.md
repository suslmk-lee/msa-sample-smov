# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Theater MSA (Microservices Architecture) is a multi-cloud cinema booking system demonstrating distributed microservices deployment across two Kubernetes clusters (NaverCloud and NHN Cloud) using Istio service mesh for traffic management.

### Architecture Components

- **API Gateway**: Central entry point with weighted traffic distribution to backend services
- **Microservices**: User Service, Movie Service, Booking Service
- **Multi-cluster Deployment**: 
  - CTX1 (NaverCloud): API Gateway + All services
  - CTX2 (NHN Cloud): All services (no API Gateway)
- **Istio Service Mesh**: VirtualService and DestinationRule for traffic splitting
- **Traffic Distribution**:
  - User Service: 70% CTX1, 30% CTX2
  - Movie Service: 30% CTX1, 70% CTX2  
  - Booking Service: 50% CTX1, 50% CTX2

## Development Commands

### Local Development (Docker Compose)
```bash
# Run all services locally
docker-compose up --build

# Individual service build
docker-compose up --build <service-name>
```

### Kubernetes Deployment Commands

#### Prerequisites Setup
```bash
# Set up kubectl contexts (required)
kubectl config rename-context <your-ctx1-context> ctx1
kubectl config rename-context <your-ctx2-context> ctx2

# Label cluster nodes (required for scheduling)
kubectl label nodes <node-name> cluster-name=ctx1 --context=ctx1
kubectl label nodes <node-name> cluster-name=ctx2 --context=ctx2

# Set Harbor domain (required for image registry)
export DOMAIN=27.96.156.180.nip.io  # Replace with your domain
```

#### Image Building and Registry
```bash
# Build and push all service images to Harbor registry
./build-images.sh [DOMAIN]

# Update deployment image tags
./update-deployment-images.sh [DOMAIN]
```

#### Deployment
```bash
# Deploy to both clusters (recommended)
./deploy-all.sh

# Deploy to individual clusters
./deploy-ctx1.sh  # NaverCloud cluster
./deploy-ctx2.sh  # NHN Cloud cluster

# Clean up deployments
./cleanup.sh --all
```

#### Monitoring and Debugging
```bash
# Check deployment status across clusters
kubectl get pods -n theater-msa --context=ctx1 -o wide
kubectl get pods -n theater-msa --context=ctx2 -o wide

# Check Istio traffic configuration
kubectl get vs,dr -n theater-msa --context=ctx1
kubectl get vs,dr -n theater-msa --context=ctx2

# View service logs
kubectl logs -l app=<service-name> -n theater-msa --context=<ctx1|ctx2>

# Debug pod issues
kubectl describe pod <pod-name> -n theater-msa --context=<ctx1|ctx2>
```

## Code Architecture

### Service Structure
Each microservice follows this pattern:
- `main.go`: Server setup and routing
- `handlers.go`: HTTP request handlers
- `models.go`: Data structures and business logic
- `store.go`: Redis data persistence layer
- `Dockerfile`: Container build configuration

### API Gateway Architecture
The API Gateway (`api-gateway/main.go`) implements:
- **Weighted Load Balancing**: Uses TrafficWeight struct to distribute requests
- **Kubernetes Integration**: Fetches deployment status via K8s client
- **Traffic Monitoring**: Tracks routing decisions with TrafficHistory
- **Static File Serving**: Serves UI files from ConfigMap

Key functions:
- `weightedServiceSelect()`: Implements probabilistic traffic distribution
- `getTrafficWeights()`: Returns current traffic configuration
- `getDeploymentStatus()`: Provides cluster deployment information

### Multi-Cluster Deployment Strategy

#### CTX1 (NaverCloud)
- Runs API Gateway with external access
- Hosts all services with `cluster-name=ctx1` node affinity
- Serves UI via ConfigMap mounted to API Gateway

#### CTX2 (NHN Cloud)  
- Hosts all services with `cluster-name=ctx2` node affinity
- No external access (traffic routed through CTX1)
- Participates in Istio service mesh for internal traffic

### Istio Configuration

#### VirtualService Traffic Splitting
Services use weighted routing defined in `istio-virtualservices.yaml`:
- Canary deployment support via `x-canary: true` header
- Percentage-based traffic distribution between clusters
- Fallback routing for regular traffic

#### DestinationRule Configuration
Defines service subsets (`ctx1`, `ctx2`) for traffic targeting based on cluster labels.

## Important Files

### Kubernetes Manifests
- `*-multicloud.yaml`: Multi-cluster service deployments
- `istio-virtualservices.yaml`: Traffic splitting configuration
- `istio-destinationrules.yaml`: Service subset definitions
- `ui-configmap.yaml`: Frontend UI with traffic visualization

### Scripts
- `deploy-all.sh`: Multi-cluster deployment orchestration
- `build-images.sh`: Container image build and push
- `update-deployment-images.sh`: Update image tags in manifests

### Service Code
- `api-gateway/main.go`: Central gateway with weighted routing
- `services/*/handlers.go`: REST API implementations
- `services/*/store.go`: Redis integration layer

## Traffic Visualization

The UI includes real-time traffic distribution visualization:
- 16-light signal display showing CTX1/CTX2 routing decisions
- Actual vs configured traffic ratio display
- Integration with `/traffic-weights` API endpoint

## Environment Variables

Required for deployment:
- `DOMAIN`: Harbor registry domain (e.g., `27.96.156.180.nip.io`)
- `USER_SERVICE_CTX1_WEIGHT`: User service CTX1 traffic percentage (default: 70)
- `USER_SERVICE_CTX2_WEIGHT`: User service CTX2 traffic percentage (default: 30)
- `MOVIE_SERVICE_CTX1_WEIGHT`: Movie service CTX1 traffic percentage (default: 30)
- `MOVIE_SERVICE_CTX2_WEIGHT`: Movie service CTX2 traffic percentage (default: 70)
- `BOOKING_SERVICE_CTX1_WEIGHT`: Booking service CTX1 traffic percentage (default: 50)
- `BOOKING_SERVICE_CTX2_WEIGHT`: Booking service CTX2 traffic percentage (default: 50)

## Common Issues and Solutions

### Node Affinity Issues
If pods are pending due to node constraints, verify cluster labels:
```bash
kubectl get nodes --show-labels --context=<ctx1|ctx2>
```

### Image Pull Failures
Ensure Harbor registry is accessible and images are pushed:
```bash
# Check image availability
docker pull harbor.$DOMAIN/theater-msa/<service>:latest
```

### Traffic Distribution Not Working
Verify Istio configuration and service mesh status:
```bash
istioctl proxy-config endpoints deployment/<service> --context=<ctx1|ctx2>
```

## Testing Multi-Cluster Deployment

### Basic Connectivity
```bash
# Test individual services
curl http://theater.$DOMAIN/users/
curl http://theater.$DOMAIN/movies/
curl http://theater.$DOMAIN/bookings/

# Test canary routing
curl -H 'x-canary: true' http://theater.$DOMAIN/users/
```

### Traffic Distribution Verification
```bash
# Multiple requests to observe load balancing
for i in {1..10}; do curl -s http://theater.$DOMAIN/users/ | head -1; done
```