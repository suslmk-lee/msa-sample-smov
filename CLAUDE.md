# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Theater MSA (Microservices Architecture) is a multi-cloud cinema booking system demonstrating distributed microservices deployment across two Kubernetes clusters (NaverCloud and NHN Cloud) using Istio service mesh for traffic management.

이 프로젝트는 K-PaaS 교육용 MSA 플랫폼으로, 실제 Istio 서비스메시의 트래픽 관리, Fault Injection, Circuit Breaker 등의 핵심 기능을 실습할 수 있는 종합 교육 환경을 제공합니다.

**2025-06-23 Self-contained Architecture 개선**: Practice 폴더의 모든 시나리오가 외부 의존성 없이 독립적으로 실행되며, DestinationRule 충돌 문제가 완전히 해결되었습니다.

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
- **Fault Injection & Circuit Breaker**: 교육용 장애 시나리오 및 자동 복구 기능
- **Real-time Traffic Visualization**: 실제 Istio 라우팅 결과 추적 및 시각화

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

#### Quick Start (교육용 권장)
```bash
# 전체 멀티클러스터 통합 배포
cd deploy/
./deploy-all.sh

# 웹 UI 접근
echo "https://theater.$DOMAIN"
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
- **Redis Deployment**: Actual Redis instance deployed here
- No external access (traffic routed through CTX1)
- Participates in Istio service mesh for internal traffic

#### Redis Multi-Cluster Architecture
- **CTX1**: Redis Service only (no endpoints) - enables multi-cluster access
- **CTX2**: Actual Redis Deployment + Service - real data storage
- **EastWestGateway**: Transparent cross-cluster Redis access for education
- **Service Mesh Native**: No proxy, pure Istio multi-cluster service discovery

### Istio Configuration

#### VirtualService Traffic Splitting
Services use weighted routing defined in `istio-virtualservices.yaml`:
- Canary deployment support via `x-canary: true` header
- Percentage-based traffic distribution between clusters
- Fallback routing for regular traffic

#### DestinationRule Configuration
Defines service subsets (`ctx1`, `ctx2`) for traffic targeting based on cluster labels.

## Important Files

### Deployment Directory (deploy/)
- `deploy-all.sh`: 멀티클러스터 통합 배포 스크립트 (교육용 권장)
- `deploy-ctx1.sh`, `deploy-ctx2.sh`: 개별 클러스터 배포
- `cleanup.sh`: 일괄 정리 스크립트
- `*-ctx1.yaml`, `*-ctx2.yaml`: 클러스터별 서비스 배포 매니페스트
- `istio-virtualservices.yaml`: 트래픽 분산 설정
- `istio-destinationrules.yaml`: 서비스 subset 정의
- `ui-configmap.yaml`: 실시간 트래픽 시각화 UI

### Practice Directory (practice/) - Self-contained Architecture
**Each scenario is completely self-contained with no external dependencies**

- `fault-injection-demo.sh`: Enhanced Fault Injection 관리 (충돌 방지, 환경 검증, 롤백 기능)
- `01-initial/`: 기본 Round Robin 설정 (독립적 패키지)
  - `destinationrules.yaml`, `virtualservices.yaml`, `kustomization.yaml`
- `02-circuit-breaker/`: Circuit Breaker 설정 (완전 독립)
  - `destinationrules.yaml`, `virtualservices.yaml` (로컬 복사본), `kustomization.yaml`
- `03-delay-fault/`: 지연 장애 시나리오 (Circuit Breaker 포함)
  - `destinationrules.yaml`, `virtualservices.yaml`, `kustomization.yaml`
- `04-error-fault/`: 오류 장애 시나리오 (Circuit Breaker 포함)
  - `destinationrules.yaml`, `virtualservices.yaml`, `kustomization.yaml`
- `05-block-fault/`: 클러스터 차단 시나리오 (Circuit Breaker 포함)
  - `destinationrules.yaml`, `virtualservices.yaml`, `kustomization.yaml`
- `99-scenarios/`: 복합 장애 시나리오 (Circuit Breaker 포함)
  - `destinationrules.yaml`, `multi-service-fault.yaml`, `kustomization.yaml`
- `README.md`: Self-contained 아키텍처 상세 가이드

### Service Code
- `api-gateway/main.go`: 중앙 게이트웨이 (가중치 라우팅, K8s 연동)
- `services/*/handlers.go`: REST API 구현 (실제 클러스터 정보 헤더 포함)
- `services/*/store.go`: Redis 연동 레이어

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

### DestinationRule 충돌 문제 (해결됨)
**Problem**: Subset 이름 중복으로 인한 라우팅 오류
**Solution**: Self-contained 아키텍처로 완전 해결
```bash
# 자동 충돌 방지 (fault-injection-demo.sh)
./fault-injection-demo.sh reset  # 기존 DR 정리 후 적용
./fault-injection-demo.sh setup  # 충돌 없는 Circuit Breaker 적용
```

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

### Practice 시나리오 문제 해결
```bash
# 환경 검증
./fault-injection-demo.sh status

# 시나리오 적용 실패 시
./fault-injection-demo.sh reset  # 완전 초기화

# 개별 시나리오 롤백
# rollback_scenario() 함수로 특정 장애만 해제 가능
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

# 실제 라우팅 결과 확인 (헤더 포함)
curl -v http://theater.$DOMAIN/users/ 2>&1 | grep "X-Service-Cluster"
```

## Fault Injection 실습 (Self-contained Architecture)

### Enhanced Fault Injection Features
- **DestinationRule 충돌 방지**: 자동 기존 리소스 정리
- **환경 검증**: 클러스터, 네임스페이스, 서비스 상태 사전 확인
- **시나리오별 롤백**: 개별 장애 선택적 해제
- **완전 독립 패키지**: 각 시나리오 self-contained 구조

### 교육 시나리오 실행
```bash
cd practice/

# 사용 가능한 시나리오 확인
./fault-injection-demo.sh --help

# 권장 학습 순서 (개선된 충돌 방지 로직)
./fault-injection-demo.sh reset    # 초기 상태 (기존 DR 정리 + 기본 설정)
./fault-injection-demo.sh setup    # Circuit Breaker 설정 (충돌 없는 안전한 적용)
./fault-injection-demo.sh delay    # 지연 장애 (완전 독립 패키지)
./fault-injection-demo.sh error    # 오류 장애 (완전 독립 패키지)
./fault-injection-demo.sh block    # 클러스터 차단 (완전 독립 패키지)
./fault-injection-demo.sh chaos    # 복합 장애 (완전 독립 패키지)

# 상태 확인 및 테스트 (환경 검증 포함)
./fault-injection-demo.sh status   # 현재 설정 확인 + 환경 상태 검증
./fault-injection-demo.sh test     # API 응답 테스트

# 개별 시나리오 직접 실행 (Self-contained)
kubectl apply -k practice/01-initial/        # 어디서든 안전하게 실행
kubectl apply -k practice/02-circuit-breaker/ # 외부 의존성 없음
kubectl apply -k practice/03-delay-fault/    # 완전 독립적 패키지
```

### Circuit Breaker 테스트
```bash
# 고오류율 테스트 (Circuit Breaker 트리거용)
curl -H "x-circuit-test: true" http://theater.$DOMAIN/users/

# Envoy 통계 확인
kubectl exec -n theater-msa deployment/api-gateway -- \
  curl -s localhost:15000/stats | grep user_service
```