# Theater MSA - 교육용 Kubernetes 배포 가이드

이 프로젝트는 **교육 시연용** MSA(Microservices Architecture) 샘플 애플리케이션으로, **NaverCloud Platform**과 **NHN Cloud NKS**의 **Istio 서비스메시**를 활용한 **DestinationRule/VirtualService 기반 멀티클라우드 트래픽 관리**를 시연할 수 있도록 최적화되었습니다.

## 📋 프로젝트 개요

```
┌─────────────────────────────────────────────────────────────┐
│          Istio DestinationRule/VirtualService 기반           │
│             멀티클라우드 서비스메시 트래픽 관리               │
├─────────────────────────────────────────────────────────────┤
│  NaverCloud Platform    │    NHN Cloud NKS                  │
│  (Istio Pre-installed)  │    (Istio Pre-installed)          │
│  ┌─────────────────────┐│    ┌─────────────────────┐        │
│  │   User Service      ││    │   Movie Service     │        │
│  │   Booking Service   ││    │   API Gateway       │        │
│  │   Redis             ││    │                     │        │
│  └─────────────────────┘│    └─────────────────────┘        │
│           │              │              │                   │
│    ┌─────────────┐       │       ┌─────────────┐           │
│    │EASTWESTGATEWAY│◄────┼──────►│EASTWESTGATEWAY│          │
│    └─────────────┘       │       └─────────────┘           │
└─────────────────────────────────────────────────────────────┘
```

### 🎯 주요 특징
- **간단한 MSA 구조**: 교육용으로 복잡성 최소화
- **Istio 네이티브 트래픽 관리**: DestinationRule과 VirtualService를 통한 서비스메시 기반 로드 밸런싱
- **EASTWESTGATEWAY**: 클러스터 간 자동 서비스 디스커버리 및 투명한 멀티클러스터 통신
- **멀티클라우드 지원**: Naver Cloud + NHN Cloud 환경 최적화
- **가중치 기반 트래픽 분산**: 서비스별 차별화된 트래픽 라우팅 (User: 70%/30%, Movie: 30%/70%, Booking: 50%/50%)
- **카나리 배포 지원**: x-canary 헤더를 통한 특정 클러스터 라우팅
- **즉시 시연 가능**: 복잡한 설정 없이 빠른 배포
- **관측성 확인**: Kiali, Jaeger를 통한 트래픽 플로우 시각화
- **실제 동작 확인**: REST API 테스트 가능

## 🏗️ 아키텍처

### 마이크로서비스 구성
```
API Gateway (8080)
    ├── User Service (8081)    - 사용자 관리
    ├── Movie Service (8082)   - 영화 정보 관리  
    ├── Booking Service (8083) - 예약 관리
    └── Redis (6379)          - 데이터 저장소
```

### Istio 서비스메시 트래픽 관리
- **DestinationRule**: 클러스터별 subset 정의 및 ROUND_ROBIN 로드밸런싱
- **VirtualService**: 가중치 기반 트래픽 분산 및 카나리 배포
- **서비스별 차별화된 트래픽 비율**: 각 서비스의 특성에 맞는 클러스터 분산
- **Envoy 네이티브 처리**: 애플리케이션 코드 수정 없이 인프라 레벨 트래픽 관리

## 📁 파일 구조

```
k8s/
├── namespace.yaml                # 네임스페이스 및 설정 (Istio injection 활성화)
├── redis.yaml                   # Redis 데이터 저장소 (자동 초기 데이터)
├── user-service.yaml            # 사용자 서비스 (기본)
├── movie-service.yaml           # 영화 서비스 (기본)
├── booking-service.yaml         # 예약 서비스 (기본)
├── user-service-multicloud.yaml # 멀티클라우드 사용자 서비스 (ctx1, ctx2)
├── movie-service-multicloud.yaml # 멀티클라우드 영화 서비스 (ctx1, ctx2)  
├── booking-service-multicloud.yaml # 멀티클라우드 예약 서비스 (ctx1, ctx2)
├── api-gateway.yaml             # API 게이트웨이 (단순 프록시)
├── rbac.yaml                    # API Gateway용 서비스 계정 및 권한 설정
├── ui-configmap.yaml            # UI 파일 (Istio 설정 표시)
├── istio-destinationrules.yaml  # DestinationRule (클러스터별 subset)
├── istio-virtualservices.yaml   # VirtualService (가중치 기반 라우팅)
├── istio-gateway.yaml           # Istio Gateway (cp-gateway 사용)
├── istio-virtualservice.yaml    # 외부 접근용 VirtualService
├── deploy.yaml                  # 배포 권한 설정
├── kustomization.yaml           # 통합 배포 설정
├── build-images.sh              # Harbor 이미지 빌드 스크립트
├── update-deployment-images.sh  # Deployment YAML 이미지 태그 일괄 변경 스크립트
├── cleanup.sh                   # 샘플 배포 일괄 삭제 스크립트
└── README.md                   # 이 파일
```

## 📋 사전 요구사항 및 제약조건

### 필수 환경 조건
- **Kubernetes 클러스터**: 2개 (NaverCloud Platform, NHN Cloud NKS)
- **Istio 사전 설치**: 각 클러스터에 Istio가 설치되어 있어야 함
- **EASTWESTGATEWAY 구성**: 클러스터 간 통신을 위해 사전 구성되어 있어야 함
- **cp-gateway 존재**: `istio-system` 네임스페이스에 cp-gateway가 구성되어 있어야 함
- **Harbor Registry**: 컨테이너 이미지 저장소 (harbor.{{DOMAIN}} 형태)
- **Docker**: 이미지 빌드 및 푸시를 위한 Docker 엔진

### 제약 조건

#### 1. 클러스터 Context 명명 규칙
```bash
# 필수: kubectl context 이름을 다음과 같이 설정해야 함
kubectl config rename-context <original-context-1> ctx1
kubectl config rename-context <original-context-2> ctx2

# 확인
kubectl config get-contexts
```

#### 2. 노드 라벨링 요구사항
```bash
# ctx1 클러스터의 모든 노드에 라벨 필수 적용
kubectl label nodes <node-name> cluster-name=ctx1 --context=ctx1

# ctx2 클러스터의 모든 노드에 라벨 필수 적용  
kubectl label nodes <node-name> cluster-name=ctx2 --context=ctx2
```

#### 3. 네트워크 접근 요구사항
- **외부 도메인**: `theater.{{DOMAIN}}` 형태로 환경별 설정 필요
- **포트 개방**: 80, 443 포트가 외부에서 접근 가능해야 함
- **DNS 해결**: 설정한 도메인이 해결 가능해야 함

#### 4. 서비스 배포 제약사항
- **고정 배포**: nodeAffinity로 서비스별 클러스터 고정 배포
  - ctx1: User Service, API Gateway (cp-gateway 위치)
  - ctx2: Movie Service, Booking Service
- **Redis**: 단일 Redis 서비스 (preferredAffinity로 클러스터 배치)
- **초기 데이터**: Redis 시작 시 자동으로 3명의 사용자 데이터 생성
- **네임스페이스**: 모든 리소스는 `theater-msa` 네임스페이스에 배포

#### 5. Istio 설정 요구사항
- **VirtualService**: `istio-system` 네임스페이스에 배포해야 함
- **Gateway**: 기존 `cp-gateway` 재사용 (새로 생성하지 않음)
- **호스트명**: `theater.{{DOMAIN}}` 템플릿 형태로 환경별 설정

#### 6. 권한 요구사항
```bash
# 각 클러스터에서 다음 권한이 필요함
- pods, services, deployments: get, list, create, update, delete
- namespaces: get, list, create
- virtualservices, destinationrules: get, list, create, update, delete
- nodes: get, list, patch (라벨링용)
```

## 🚀 배포 방법 (상세)

### 1. 사전 준비

#### Harbor Registry 설정 (이미지 저장소)
Harbor에 repo 설정 
  /theater-msa 생성

#### 도메인 설정

##### 방법 2: 수동 설정

```bash
export DOMAIN="27.96.156.180.nip.io"
```

#### 클러스터 연결 확인
```bash
# kubectl 명령어 확인
kubectl version --client

# 클러스터 연결 확인 (각각)
kubectl config use-context ctx1
kubectl cluster-info

kubectl config use-context ctx2  
kubectl cluster-info
```

### 2. 이미지 빌드 및 Registry 업로드

#### Harbor Registry에 이미지 업로드
```bash
# 1. Harbor 로그인 (사전에 Harbor 계정 필요)
podman login harbor.${DOMAIN}

# 2. 모든 서비스 이미지 빌드 및 푸시 (자동화)
./build-images.sh ${DOMAIN}

```

#### 개별 이미지 빌드 (수동)
```bash
# 상위 디렉토리로 이동
cd ..

# 각 서비스별 이미지 빌드
docker build -t harbor.${DOMAIN}/theater-msa/user-service:latest ./services/user-service/
docker build -t harbor.${DOMAIN}/theater-msa/movie-service:latest ./services/movie-service/
docker build -t harbor.${DOMAIN}/theater-msa/booking-service:latest ./services/booking-service/
docker build -t harbor.${DOMAIN}/theater-msa/api-gateway:latest ./api-gateway/

# 각 이미지 푸시
docker push harbor.${DOMAIN}/theater-msa/user-service:latest
docker push harbor.${DOMAIN}/theater-msa/movie-service:latest
docker push harbor.${DOMAIN}/theater-msa/booking-service:latest
docker push harbor.${DOMAIN}/theater-msa/api-gateway:latest

# k8s 디렉토리로 돌아가기
cd k8s/
```

#### Deployment YAML 이미지 태그 업데이트
```bash
# Harbor Registry 이미지 태그로 일괄 변경
./update-deployment-images.sh ${DOMAIN}
```

### 3. 멀티클라우드 서비스 배포 (DestinationRule/VirtualService 기반)

#### Step 1: ctx1 클러스터 (User Service + API Gateway)
```bash
# ctx1 클러스터 접속
kubectl config use-context ctx1

# 기본 리소스 배포
kubectl apply -f namespace.yaml
kubectl apply -f rbac.yaml
kubectl apply -f ui-configmap.yaml
kubectl apply -f redis.yaml

# 멀티클라우드 서비스 배포 (클러스터 라벨 포함)
kubectl apply -f user-service-multicloud.yaml
kubectl apply -f movie-service-multicloud.yaml
kubectl apply -f booking-service-multicloud.yaml
kubectl apply -f api-gateway.yaml

# Istio 트래픽 관리 설정 배포
kubectl apply -f istio-destinationrules.yaml
kubectl apply -f istio-virtualservices.yaml
kubectl apply -f istio-virtualservice.yaml  # 외부 접근용
```

#### Step 2: ctx2 클러스터 (Movie + Booking Service)  
```bash
# ctx2 클러스터 접속
kubectl config use-context ctx2

# 기본 리소스 배포
kubectl apply -f namespace.yaml
kubectl apply -f redis.yaml

# 멀티클라우드 서비스 배포 (클러스터 라벨 포함)
kubectl apply -f user-service-multicloud.yaml
kubectl apply -f movie-service-multicloud.yaml  
kubectl apply -f booking-service-multicloud.yaml

# Istio 트래픽 관리 설정 배포
kubectl apply -f istio-destinationrules.yaml
kubectl apply -f istio-virtualservices.yaml
```

#### Step 3: 전체 배포 (Kustomize 사용) - 권장
```bash
# 각 클러스터에서 실행 (모든 리소스 자동 배포)
kubectl config use-context ctx1
kubectl apply -k .

kubectl config use-context ctx2  
kubectl apply -k .
```

#### Step 4: 트래픽 분산 동작 확인
```bash
# 각 클러스터에서 Pod 분산 상태 확인
kubectl get pods -n theater-msa -o wide --show-labels

# VirtualService 가중치 설정 확인
kubectl get vs -n theater-msa -o yaml | grep -A 3 weight

# 실제 트래픽 분산 테스트
for i in {1..10}; do
  curl -s http://theater.$DOMAIN/users/ | head -1
  sleep 1
done
```

### 4. 배포 상태 확인
```bash
# 모든 Pod 상태 및 클러스터 분산 확인
kubectl get pods -n theater-msa -o wide --show-labels

# 서비스 확인
kubectl get svc -n theater-msa

# DestinationRule 배포 확인
kubectl get dr -n theater-msa

# VirtualService 배포 확인  
kubectl get vs -n theater-msa

# Istio 사이드카 주입 확인
kubectl get pods -n theater-msa -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.containers[*].name}{"\n"}{end}'

# 외부 접근용 VirtualService 확인
kubectl get vs -n istio-system theater-msa
```

### 5. 애플리케이션 접근

#### cp-gateway를 통한 접근 (권장)
```bash
# 외부 도메인으로 직접 접근 (도메인은 환경별 설정값 사용)
http://theater.{{DOMAIN}}

# 실제 접근 예시 (도메인 치환 후)
http://theater.27.96.156.180.nip.io

# 개별 서비스 API 접근
http://theater.{{DOMAIN}}/users/
http://theater.{{DOMAIN}}/movies/
http://theater.{{DOMAIN}}/bookings/
```

#### 로컬 테스트용 포트 포워딩
```bash
# API Gateway 직접 접근
kubectl port-forward svc/api-gateway 8080:8080 -n theater-msa

# 브라우저에서 접근
# http://localhost:8080
```

## 🧪 시연 시나리오

### 1. Istio 서비스메시 확인
```bash
# Envoy 사이드카 주입 확인
kubectl get pods -n theater-msa -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.containers[*].name}{"\n"}{end}'

# Istio 프록시 상태 확인
kubectl exec -n theater-msa deployment/user-service -c istio-proxy -- pilot-agent request GET stats/prometheus | grep envoy_cluster

# 서비스메시 구성 확인
istioctl proxy-config cluster deployment/user-service.theater-msa
```

### 2. EASTWESTGATEWAY를 통한 멀티클러스터 통신 확인
```bash
# 클라우드별 노드 라벨 확인
kubectl get nodes --show-labels | grep cloud-provider

# 서비스별 Pod 분산 상태 확인 (각 클러스터에서)
kubectl get pods -n theater-msa -o wide

# EASTWESTGATEWAY 상태 확인
kubectl get svc istio-eastwestgateway -n istio-system

# 멀티클러스터 서비스 디스커버리 확인
istioctl proxy-config endpoints deployment/user-service.theater-msa

# 원격 클러스터 서비스 접근 확인 (자동 프록시 경유)
kubectl exec -n theater-msa deployment/user-service -- curl http://movie-service.theater-msa.svc.cluster.local:8082/
```

### 3. API 테스트 (cp-gateway 경유)
```bash
# 환경별 도메인 설정 확인
echo "http://theater.$DOMAIN"

# 사용자 생성
curl -X POST -H "Content-Type: application/json" \
  -d '{"name":"홍길동","email":"hong@example.com"}' \
  http://theater.$DOMAIN/users/

# 영화 추가
curl -X POST -H "Content-Type: application/json" \
  -d '{"title":"어벤져스","genre":"액션","year":2019}' \
  http://theater.$DOMAIN/movies/

# 예약 생성
curl -X POST -H "Content-Type: application/json" \
  -d '{"userId":"user-id","movieId":"movie-id","seats":2}' \
  http://theater.$DOMAIN/bookings/

# VirtualService 라우팅 확인
kubectl get vs -n istio-system theater-msa
```

### 4. 관측성 도구 확인
```bash
# Kiali 대시보드 접근 (사전 설치된 경우)
kubectl port-forward svc/kiali 20001:20001 -n istio-system

# Jaeger 추적 확인 (사전 설치된 경우)
kubectl port-forward svc/jaeger 16686:16686 -n istio-system

# Prometheus 메트릭 확인
kubectl port-forward svc/prometheus 9090:9090 -n istio-system
```

### 5. Istio 트래픽 관리 시연
```bash
# DestinationRule 확인 (클러스터별 subset 정의)
kubectl get destinationrules -n theater-msa
kubectl describe destinationrule user-service-dr -n theater-msa

# VirtualService 확인 (가중치 기반 트래픽 분산)
kubectl get virtualservices -n theater-msa
kubectl describe virtualservice user-service-vs -n theater-msa

# 현재 트래픽 분산 설정 확인
kubectl get vs user-service-vs -n theater-msa -o yaml | grep -A 10 weight

# 트래픽 분산 비율 실시간 변경 (User Service 예시)
kubectl patch virtualservice user-service-vs -n theater-msa --type='merge' -p='
{
  "spec": {
    "http": [{
      "match": [{
        "headers": {
          "x-canary": {"exact": "true"}
        }
      }],
      "route": [{
        "destination": {"host": "user-service", "subset": "ctx2"},
        "weight": 100
      }]
    }, {
      "route": [
        {"destination": {"host": "user-service", "subset": "ctx1"}, "weight": 90},
        {"destination": {"host": "user-service", "subset": "ctx2"}, "weight": 10}
      ]
    }]
  }
}'

# 카나리 배포 테스트 (ctx2로 100% 라우팅)
curl -H "x-canary: true" http://theater.$DOMAIN/users/

# 일반 트래픽 테스트 (가중치 분산)
curl http://theater.$DOMAIN/users/

# 서비스별 트래픽 분산 확인
kubectl get vs -n theater-msa -o custom-columns=NAME:.metadata.name,WEIGHTS:.spec.http[0].route[*].weight

# Envoy 프록시 설정 확인
istioctl proxy-config cluster deployment/user-service.theater-msa
istioctl proxy-config endpoints deployment/user-service.theater-msa

# 트래픽 분산 상태 실시간 모니터링
istioctl proxy-config listeners deployment/user-service.theater-msa --port 8081

# cp-gateway 설정 확인
kubectl get gateway cp-gateway -n istio-system -o yaml
```

## 🔧 운영 및 관리

### 상태 모니터링
```bash
# Pod 상태 실시간 확인
kubectl get pods -n theater-msa -w

# 로그 확인
kubectl logs -n theater-msa -l app=api-gateway --tail=50
kubectl logs -n theater-msa -l app=user-service --tail=50

# 리소스 사용량 확인
kubectl top pods -n theater-msa
```

### 스케일링 (시연용)
```bash
# 수동 스케일링
kubectl scale deployment user-service --replicas=3 -n theater-msa

# 스케일링 상태 확인
kubectl get pods -n theater-msa -l app=user-service
```

### 업데이트 시연
```bash
# 이미지 업데이트
kubectl set image deployment/user-service user-service=user-service:v2.0.0 -n theater-msa

# 롤아웃 상태 확인
kubectl rollout status deployment/user-service -n theater-msa
```

## 🌐 멀티클라우드 설정

### ctx1 클러스터 설정 (User + API Gateway Service)
```bash
# ctx1 클러스터 접속
kubectl config use-context ctx1

# 노드에 클러스터 라벨 추가
kubectl label nodes <node-name> cluster-name=ctx1

# ctx1에 배포될 서비스들 (cp-gateway 위치)
# - User Service (user-service.yaml)
# - API Gateway (api-gateway.yaml) 
# - Redis (shared, preferred)

# 배포 확인
kubectl get pods -n theater-msa -o wide
```

### ctx2 클러스터 설정 (Movie + Booking Service)
```bash
# ctx2 클러스터 접속
kubectl config use-context ctx2

# 노드에 클러스터 라벨 추가
kubectl label nodes <node-name> cluster-name=ctx2

# ctx2에 배포될 서비스들
# - Movie Service (movie-service.yaml)
# - Booking Service (booking-service.yaml)
# - Redis (shared, preferred)

# 서비스 분산 배치 확인
kubectl get pods -n theater-msa -o wide
```

## 🛠️ 트러블슈팅

### 제약조건 관련 문제

#### 1. Context 이름 문제
```bash
# 문제: context 이름이 ctx1, ctx2가 아닌 경우
Error: context "my-cluster" not found

# 해결: context 이름 변경
kubectl config get-contexts
kubectl config rename-context <original-name> ctx1
kubectl config rename-context <original-name> ctx2
```

#### 2. 노드 라벨 누락 문제
```bash
# 문제: Pod이 Pending 상태에서 머무는 경우
0/3 nodes are available: 3 node(s) didn't match Pod's node affinity

# 해결: 노드 라벨 확인 및 추가
kubectl get nodes --show-labels | grep cluster-name
kubectl label nodes <node-name> cluster-name=ctx1  # 또는 ctx2
```

#### 3. DestinationRule/VirtualService 설정 문제
```bash
# 문제: 트래픽이 한 클러스터로만 라우팅되는 경우
# 원인: 클러스터 라벨 불일치 또는 subset 정의 오류

# 해결: 클러스터 라벨 확인
kubectl get pods -n theater-msa --show-labels | grep cluster

# DestinationRule subset 확인
kubectl describe dr user-service-dr -n theater-msa

# VirtualService 라우팅 규칙 확인
kubectl describe vs user-service-vs -n theater-msa
```

#### 4. VirtualService 배포 네임스페이스 문제
```bash
# 문제: 내부 서비스 VirtualService가 잘못된 네임스페이스에 배포
# 내부 서비스: theater-msa 네임스페이스
# 외부 접근: istio-system 네임스페이스

# 올바른 배포 확인
kubectl get vs -n theater-msa  # 내부 서비스 라우팅
kubectl get vs -n istio-system # 외부 Gateway 라우팅
```

#### 5. 도메인 접근 불가
```bash
# 문제: 설정한 도메인 접근 실패
curl: (6) Could not resolve host

# 해결: 도메인 설정 및 DNS 확인
echo $DOMAIN  # 도메인 변수 확인
nslookup theater.$DOMAIN
kubectl get gateway cp-gateway -n istio-system

# VirtualService 호스트명 확인
kubectl get vs theater-msa -n istio-system -o yaml | grep hosts
```

#### 6. 카나리 배포 동작 안함
```bash
# 문제: x-canary 헤더 라우팅이 동작하지 않는 경우

# 해결: VirtualService 매치 규칙 확인
kubectl get vs user-service-vs -n theater-msa -o yaml | grep -A 5 "x-canary"

# 테스트 요청
curl -v -H "x-canary: true" http://theater.$DOMAIN/users/

# Envoy 설정 확인
istioctl proxy-config route deployment/user-service.theater-msa
```

### 일반적인 문제해결
```bash
# Pod 실패 시 상세 정보 확인
kubectl describe pod <pod-name> -n theater-msa

# 이벤트 확인
kubectl get events -n theater-msa --sort-by=.metadata.creationTimestamp

# 서비스 연결 테스트
kubectl exec -it <pod-name> -n theater-msa -- wget -qO- http://redis:6379

# EASTWESTGATEWAY 상태 확인
kubectl get svc istio-eastwestgateway -n istio-system
```

### 배포 검증 체크리스트
```bash
# 1. Context 이름 확인
kubectl config current-context  # ctx1 또는 ctx2여야 함

# 2. 노드 라벨 확인
kubectl get nodes --show-labels | grep cluster-name

# 3. 서비스 배포 위치 확인
kubectl get pods -n theater-msa -o wide --show-labels

# 4. DestinationRule 배포 확인
kubectl get dr -n theater-msa
kubectl describe dr user-service-dr -n theater-msa | grep -A 10 subsets

# 5. VirtualService 배포 확인
kubectl get vs -n theater-msa  # 내부 서비스 라우팅
kubectl get vs -n istio-system # 외부 Gateway 라우팅

# 6. 트래픽 분산 설정 확인
kubectl get vs -n theater-msa -o custom-columns=NAME:.metadata.name,WEIGHTS:.spec.http[-1].route[*].weight

# 7. Envoy 사이드카 주입 확인
kubectl get pods -n theater-msa -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.containers[*].name}{"\n"}{end}'

# 8. 서비스메시 연결 확인
istioctl proxy-config endpoints deployment/user-service.theater-msa

# 9. 외부 접근 확인
curl -I http://theater.$DOMAIN

# 10. 카나리 배포 테스트
curl -H "x-canary: true" http://theater.$DOMAIN/users/
```

### 리소스 정리

#### 자동 정리 스크립트 사용 (권장)
```bash
# 현재 컨텍스트에서만 삭제
./cleanup.sh

# 모든 클러스터(ctx1, ctx2)에서 일괄 삭제
./cleanup.sh --all

# 도움말 확인
./cleanup.sh --help
```

#### 수동 정리 방법
```bash
# 각 클러스터에서 Kustomize를 사용한 일괄 삭제
kubectl config use-context ctx1
kubectl delete -k .

kubectl config use-context ctx2
kubectl delete -k .

# 또는 네임스페이스 삭제 (각 클러스터에서)
kubectl delete namespace theater-msa --context ctx1
kubectl delete namespace theater-msa --context ctx2

# 외부 VirtualService 삭제 (istio-system)
kubectl delete vs theater-msa -n istio-system --context ctx1
```

#### 정리 완료 확인
```bash
# 남은 리소스 확인
kubectl get all,vs,dr -n theater-msa
kubectl get vs -n istio-system theater-msa

# 네임스페이스 확인
kubectl get namespace theater-msa
```

## 📚 교육 포인트

### 1. MSA 핵심 개념
- **서비스 분리**: 각 기능별 독립적인 서비스
- **API 게이트웨이**: 단일 진입점 패턴
- **공유 데이터 저장소**: 단일 Redis를 통한 데이터 공유
- **Istio 네이티브 트래픽 분산**: DestinationRule과 VirtualService를 통한 서비스메시 기반 로드 밸런싱

### 2. Kubernetes 기본 개념
- **Pod**: 애플리케이션 실행 단위
- **Deployment**: 애플리케이션 배포 관리
- **Service**: 서비스 디스커버리
- **ConfigMap**: 설정 데이터 분리 관리 (UI 파일 포함)
- **RBAC**: 역할 기반 접근 제어 (Kubernetes API 권한)
- **ServiceAccount**: Pod의 Kubernetes API 접근 인증
- **Ingress**: 외부 접근 관리

### 3. Harbor Registry 및 이미지 관리
- **컨테이너 이미지 저장소**: 프라이빗 레지스트리를 통한 이미지 중앙 관리
- **자동화 스크립트**: build-images.sh로 일괄 이미지 빌드 및 푸시
- **배포 자동화**: update-deployment-images.sh로 YAML 이미지 태그 일괄 변경
- **백업 및 복원**: 안전한 설정 변경과 롤백 지원
- **멀티 런타임 지원**: Docker와 Podman 자동 감지 및 사용

### 4. Istio 서비스메시 개념
- **사이드카 패턴**: Envoy 프록시를 통한 투명한 네트워크 관리
- **트래픽 관리**: VirtualService, DestinationRule을 통한 세밀한 라우팅
- **보안**: mTLS 자동 적용으로 서비스간 암호화 통신
- **관측성**: 분산 추적, 메트릭, 로깅 자동 수집
- **정책 관리**: 서킷브레이커, 타임아웃, 재시도 정책

### 5. 멀티클라우드 서비스메시 (EASTWESTGATEWAY)
- **자동 서비스 디스커버리**: EASTWESTGATEWAY를 통한 클러스터 간 자동 연결
- **투명한 통신**: 애플리케이션 코드 변경 없이 멀티클러스터 통신
- **트래픽 분산**: 클라우드별 로드밸런싱 및 장애 조치
- **통합 관측성**: 전체 인프라에 걸친 통합 모니터링
- **보안 정책**: 클라우드에 관계없이 일관된 mTLS 보안

### 6. Istio 트래픽 관리 (DestinationRule & VirtualService)
- **DestinationRule 기반 클러스터 subset**: `cluster: ctx1/ctx2` 라벨을 통한 클러스터별 트래픽 분할
- **VirtualService 가중치 라우팅**: 서비스별 차별화된 트래픽 분산
  - User Service: 70% CTX1, 30% CTX2 (주요 서비스 안정성 우선)
  - Movie Service: 30% CTX1, 70% CTX2 (부하 분산 우선)
  - Booking Service: 50% CTX1, 50% CTX2 (균등 분산)
- **카나리 배포 지원**: `x-canary: true` 헤더를 통한 특정 클러스터 라우팅
- **ROUND_ROBIN 로드밸런싱**: 각 클러스터 내 Pod 간 균등 분산
- **Envoy 네이티브 처리**: 애플리케이션 수정 없이 인프라 레벨 트래픽 관리
- **동적 설정 변경**: kubectl patch를 통한 실시간 트래픽 비율 조정

## 🎓 시연 체크리스트

### 기본 배포 확인
- [ ] 클러스터 연결 확인
- [ ] Istio injection 활성화 확인
- [ ] 모든 서비스 배포 완료
- [ ] Pod에 Envoy 사이드카 주입 확인
- [ ] Istio Gateway를 통한 외부 접근 가능

### 서비스메시 기능 확인
- [ ] VirtualService 트래픽 라우팅 동작
- [ ] DestinationRule 로드밸런싱 정책 적용
- [ ] mTLS 암호화 통신 확인
- [ ] 서킷브레이커 및 재시도 정책 동작

### 멀티클라우드 기능 확인 (EASTWESTGATEWAY)
- [ ] ctx1, ctx2 클러스터 노드 라벨링 (`cluster-name=ctx1/ctx2`)
- [ ] 클러스터별 서비스 분산 배치 확인
  - [ ] ctx1: User Service, API Gateway (cp-gateway 위치)
  - [ ] ctx2: Movie Service, Booking Service  
- [ ] EASTWESTGATEWAY를 통한 자동 클러스터 간 연결
- [ ] 원격 클러스터 서비스 자동 디스커버리
- [ ] 투명한 멀티클러스터 서비스 호출 확인 (ctx1→ctx2, ctx2→ctx1)

### 관측성 도구 확인
- [ ] Kiali 서비스 토폴로지 시각화
- [ ] Jaeger 분산 추적 확인
- [ ] Prometheus 메트릭 수집 확인
- [ ] 실시간 트래픽 플로우 모니터링

## 💡 추가 학습 자료

### Istio 고급 기능 실습
- **카나리 배포**: VirtualService를 통한 점진적 배포
- **A/B 테스트**: 트래픽 분할을 통한 버전 비교
- **장애 주입**: Fault Injection을 통한 장애 복원력 테스트
- **보안 정책**: AuthorizationPolicy를 통한 세밀한 접근 제어

### EASTWESTGATEWAY 고급 시나리오
- **멀티 클러스터 메시**: EASTWESTGATEWAY를 통한 투명한 클러스터 간 연동
- **지역별 트래픽 라우팅**: 지연시간 기반 자동 라우팅
- **DR(재해복구)**: 클러스터 장애 시 EASTWESTGATEWAY를 통한 자동 failover
- **하이브리드 클라우드**: 온프레미스와 클라우드 간 투명한 연동
- **서비스 로컬리티**: 가장 가까운 클러스터의 서비스 우선 호출

### 실습 과제
1. **VirtualService 수정**: 새로운 라우팅 규칙 추가
2. **DestinationRule 최적화**: 로드밸런싱 알고리즘 변경
3. **관측성 대시보드**: Grafana 대시보드 커스터마이징
4. **보안 강화**: mTLS 정책 세부 설정

---

## ⚠️ 중요 알림

이 **Istio DestinationRule/VirtualService 기반 MSA** 시연 환경은 NaverCloud와 NHN Cloud의 **사전 설치된 Istio와 EASTWESTGATEWAY**를 활용하여 복잡한 설정 없이 즉시 **멀티클라우드 서비스메시의 트래픽 관리 핵심 기능**들을 체험할 수 있도록 구성되었습니다.

### 필수 준수사항
1. **Context 명명**: 반드시 `ctx1`, `ctx2`로 설정해야 함
2. **노드 라벨링**: 각 클러스터 노드에 `cluster-name=ctx1/ctx2` 라벨 필수
3. **클러스터 라벨**: 서비스 Pod에 `cluster: ctx1/ctx2` 라벨 필수 (DestinationRule subset 매칭용)
4. **도메인 설정**: `theater.{{DOMAIN}}` 템플릿을 환경에 맞게 치환 필요
5. **네임스페이스 구분**: 
   - 내부 서비스 트래픽 관리: `theater-msa` 네임스페이스
   - 외부 Gateway 접근: `istio-system` 네임스페이스
6. **Gateway 재사용**: 기존 `cp-gateway` 사용 (새로 생성 금지)

### Istio 네이티브 트래픽 관리 동작 원리
- **DestinationRule**: `cluster: ctx1/ctx2` 라벨을 기반으로 클러스터별 subset 정의
- **VirtualService**: 서비스별 차별화된 가중치로 트래픽 분산 (User: 70%/30%, Movie: 30%/70%, Booking: 50%/50%)
- **Envoy 프록시**: 애플리케이션 코드 수정 없이 자동 로드밸런싱 및 트래픽 분산
- **EASTWESTGATEWAY**: 클러스터 간 투명한 서비스 디스커버리 및 통신

**클러스터 간 서비스 호출 흐름 (Istio 기반):**
- API Gateway → VirtualService → DestinationRule → ctx1/ctx2 User Service
- User Service → VirtualService → DestinationRule → ctx1/ctx2 Movie Service  
- EASTWESTGATEWAY를 통한 투명한 멀티클러스터 통신