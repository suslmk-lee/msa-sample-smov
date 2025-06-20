# MSA Sample SMOV - 개발 히스토리

## 프로젝트 개요
Theater Management System의 멀티클라우드 MSA(Microservices Architecture) 데모 애플리케이션

### 아키텍처
- **CTX1 (NaverCloud)**: API Gateway + 3개 서비스 (user, movie, booking)
- **CTX2 (NHN Cloud)**: 3개 서비스 (user, movie, booking)
- **Istio Service Mesh**: VirtualService와 DestinationRule을 통한 트래픽 분산
- **트래픽 분산 정책**: 
  - User Service: 70% CTX1, 30% CTX2
  - Movie Service: 30% CTX1, 70% CTX2
  - Booking Service: 50% CTX1, 50% CTX2

## 개발 진행 상황

### 2025-06-20

#### 1. Git 리셋 및 기본 배포 설정
- `git reset --hard HEAD~1`로 이전 커밋으로 복구
- 멀티클라우드 배포 아키텍처 재구성

#### 2. 누락된 YAML 파일 생성
**문제**: 배포 스크립트 실행 시 여러 파일 누락
- `user-service-multicloud.yaml` - 사용자 서비스 멀티클라우드 배포
- `movie-service-multicloud.yaml` - 영화 서비스 멀티클라우드 배포  
- `booking-service-multicloud.yaml` - 예약 서비스 멀티클라우드 배포
- `istio-virtualservice.yaml` - 외부 트래픽 라우팅 설정

**해결**: 각 서비스별로 ctx1, ctx2에 모두 배포되도록 multicloud YAML 파일 생성

#### 3. Istio 게이트웨이 설정 수정
**문제**: `istio-gateway.yaml` 파일이 비어있어 "no objects passed to apply" 오류
**해결**: 파일을 `istio-gateway.yaml.disabled`로 변경 (기존 cp-gateway 사용)

#### 4. API Gateway 배포 정책 수정
**문제**: API Gateway가 ctx2 노드에도 스케줄링되려 함
**해결**: `requiredDuringSchedulingIgnoredDuringExecution`으로 변경하여 ctx1에만 배포

#### 5. 트래픽 시각화 UI 구현
**기능 요구사항**: 
- 사용자 목록 상단에 CTX1, CTX2 트래픽 분산 시각화
- 16개 신호등으로 실시간 트래픽 표시
- VirtualService 설정값과 실제 트래픽 비율 표시

**구현 단계**:
1. **기본 신호등 UI 추가** - CTX1, CTX2 각각 16개 신호등
2. **트래픽 시뮬레이션** - 70%/30% 확률로 클러스터 선택
3. **실시간 비율 계산** - 최근 100건 요청 기반 비율 표시
4. **UI 레이아웃 최적화** - 2줄 배치, 모던한 디자인
5. **성능 최적화** - dataInitialized 플래그, Promise.all 병렬 처리

#### 6. UI 개선 및 최적화
**개선사항**:
- 신호등 크기 30% 축소
- 통계 박스와 신호등 박스 높이 통일
- "VirtualService 설정" → "트래픽설정"으로 텍스트 변경
- 실제 VirtualService 설정값 로드 기능 구현

#### 7. 실제 트래픽 설정값 연동
**구현**:
- `/traffic-weights` API 엔드포인트 활용
- `loadVirtualServiceConfig()` 함수로 실제 가중치 로드
- API Gateway의 TrafficWeight 구조체에서 데이터 가져오기
- 오류 시 기본값(70% : 30%) 사용

#### 8. 영화목록 및 예약내역 트래픽 시각화 확장
**구현사항**:
- 모든 서비스 섹션에 동일한 트래픽 시각화 적용
- 서비스별 독립적인 신호등 및 통계 관리
- 영화 서비스: 30% CTX1, 70% CTX2 설정값 연동
- 예약 서비스: 50% CTX1, 50% CTX2 설정값 연동

#### 9. 데이터 중복 생성 문제 해결
**문제**: 사용자목록 새로고침 시 데이터가 계속 증가
**해결**:
- 견고한 초기화 로직 구현 (`initializationPromise` 사용)
- Redis에서 올바른 JSON 형식으로 초기 데이터 저장
- UI 초기화 로직에서 중복 데이터 생성 방지

#### 10. Redis 아키텍처 최적화
**변경사항**:
- Redis Deployment: CTX2에만 배포
- Redis Service: 양쪽 클러스터에 존재하여 멀티클러스터 접근
- 초기 데이터 형식 수정: `user:ID` 형태의 JSON 문자열로 저장

#### 11. **🎯 실제 Istio 라우팅 추적 구현 (주요 개선)**
**기존 문제**: JavaScript 시뮬레이션으로 가짜 트래픽 분산 표시
**해결 방안**: 실제 Istio 라우팅 결과 추적 시스템 구현

**구현 내용**:
1. **서비스별 라우팅 정보 헤더 추가**:
   ```go
   // 각 마이크로서비스에서 응답 헤더에 실제 클러스터 정보 포함
   w.Header().Set("X-Service-Cluster", getClusterName())
   w.Header().Set("X-Pod-Name", os.Getenv("HOSTNAME"))
   w.Header().Set("X-Service-Name", "user-service")
   
   func getClusterName() string {
       // 환경변수 또는 파드명에서 클러스터 정보 추출
       if cluster := os.Getenv("CLUSTER_NAME"); cluster != "" {
           return cluster
       }
       hostname := os.Getenv("HOSTNAME")
       if strings.Contains(hostname, "ctx1") {
           return "ctx1"
       } else if strings.Contains(hostname, "ctx2") {
           return "ctx2"
       }
       return "unknown"
   }
   ```

2. **UI에서 실제 라우팅 결과 추적**:
   ```javascript
   async function loadUsers() {
       const response = await fetch('/users/');
       const users = await response.json();
       
       // 실제 Istio 라우팅 결과 추적
       const routedCluster = response.headers.get('X-Service-Cluster');
       const podName = response.headers.get('X-Pod-Name');
       const serviceName = response.headers.get('X-Service-Name');
       
       console.log(`실제 라우팅 결과 - 서비스: ${serviceName}, 클러스터: ${routedCluster}, 파드: ${podName}`);
       
       if (routedCluster) {
           updateTrafficVisualization('user', routedCluster);
       }
   }
   ```

3. **시뮬레이션 로직 완전 제거**: 가짜 확률 계산 제거, 100% 실제 라우팅 결과 기반

**교육적 가치 향상**:
- ✅ **진정한 서비스 메시 동작 시연**: 실제 Istio VirtualService 라우팅 결과 표시
- ✅ **투명한 트래픽 흐름**: 각 요청이 실제로 어느 클러스터/파드로 라우팅되었는지 추적
- ✅ **실시간 검증**: VirtualService 설정이 실제로 작동하는지 눈으로 확인
- ✅ **실무 적용성**: 실제 프로덕션 환경에서 사용할 수 있는 모니터링 방식

### 주요 파일 변경사항

#### `/k8s/ui-configmap.yaml`
```javascript
// VirtualService 설정 로드
async function loadVirtualServiceConfig() {
    try {
        const response = await fetch('/traffic-weights');
        const weights = await response.json();
        
        if (weights) {
            const ctx1Weight = weights.UserServiceCtx1Weight || 70;
            const ctx2Weight = weights.UserServiceCtx2Weight || 30;
            document.getElementById('vs-ratio').textContent = `${ctx1Weight}% : ${ctx2Weight}%`;
            console.log('트래픽 가중치 로드됨:', weights);
        }
    } catch (error) {
        console.log('트래픽 가중치 로드 실패, 기본값 사용:', error);
    }
}
```

#### `/api-gateway/main.go`
```go
// TrafficWeight represents service traffic distribution
type TrafficWeight struct {
    UserServiceCtx1Weight    int
    UserServiceCtx2Weight    int
    MovieServiceCtx1Weight   int
    MovieServiceCtx2Weight   int
    BookingServiceCtx1Weight int
    BookingServiceCtx2Weight int
}

// getTrafficWeights returns current traffic weight configuration
func getTrafficWeights(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(trafficWeights)
}
```

### 현재 배포 현황 (2025-06-20 최종)
- **API Gateway**: CTX1에만 배포 (정상)
- **Redis**: CTX2에만 배포, 양쪽 클러스터에서 접근 가능
- **User/Movie/Booking Services**: CTX1, CTX2 모두 배포 (정상)
- **트래픽 시각화**: 실제 Istio 라우팅 결과 기반 실시간 표시
- **설정 연동**: 실제 VirtualService 가중치 값과 실제 라우팅 결과 모두 표시

#### 실제 테스트 결과
```bash
# User Service 요청 결과
< X-Service-Cluster: ctx2
< X-Pod-Name: user-service-ctx2-754bc8dd6f-kghtc
< X-Service-Name: user-service

# Movie Service 요청 결과  
< X-Service-Cluster: ctx2
< X-Pod-Name: movie-service-ctx2-54d9dbffc4-9vmjb
< X-Service-Name: movie-service

# Booking Service 요청 결과
< X-Service-Cluster: ctx1
< X-Pod-Name: booking-service-ctx1-5498cbb9cf-s25hr
< X-Service-Name: booking-service
```

### 성능 최적화
1. **데이터 초기화 최적화**: 중복 데이터 생성 방지
2. **병렬 요청 처리**: Promise.all로 동시 요청
3. **UI 응답성 개선**: 불필요한 DOM 조작 최소화

### 완료된 주요 마일스톤
- [x] **멀티클라우드 MSA 기본 아키텍처 구축**
- [x] **Istio 서비스 메시 트래픽 분산 구현**
- [x] **전체 서비스 트래픽 시각화 완료** (User, Movie, Booking)
- [x] **실제 Istio 라우팅 추적 시스템 구현** ⭐
- [x] **Redis 멀티클러스터 아키텍처 최적화**
- [x] **데이터 중복 생성 문제 해결**
- [x] **실시간 VirtualService 설정값 연동**

### 향후 발전 방향
- [ ] **카나리 배포 시연 기능**: `x-canary: true` 헤더 테스트
- [ ] **장애 주입 및 복구 시나리오**: Fault Injection 실습
- [ ] **분산 추적 통합**: Jaeger/Zipkin 연동
- [ ] **메트릭 대시보드**: Prometheus + Grafana 통합
- [ ] **보안 정책 실습**: mTLS, AuthorizationPolicy 시연
- [ ] **동적 트래픽 제어**: 실시간 VirtualService 가중치 조정 UI

## 기술 스택
- **컨테이너**: Docker, Kubernetes
- **서비스 메시**: Istio (VirtualService, DestinationRule)
- **프론트엔드**: HTML5, CSS3, Vanilla JavaScript
- **백엔드**: Go (API Gateway), Node.js (Services)
- **클라우드**: NaverCloud, NHN Cloud
- **모니터링**: Kubernetes API 기반 실시간 상태 확인

## 교육적 가치 및 활용 방안

### 🎓 **서비스 메시 교육 시나리오**
1. **Level 1 - 기본 이해**: VirtualService 트래픽 분산 관찰
2. **Level 2 - 실전 적용**: 실제 Istio 라우팅 결과 분석
3. **Level 3 - 고급 활용**: 카나리 배포, 장애 주입 실습
4. **Level 4 - 운영 관리**: 메트릭 모니터링, 보안 정책 적용

### 🔍 **주요 학습 포인트**
- **멀티클라우드 아키텍처**: 실제 클라우드 간 서비스 통신
- **서비스 메시 트래픽 관리**: VirtualService/DestinationRule 실습
- **관찰 가능성**: 실시간 라우팅 추적 및 시각화
- **마이크로서비스 패턴**: API Gateway, 서비스 분산, 데이터 관리

### 🚀 **실무 적용성**
- 실제 프로덕션 환경에서 사용 가능한 모니터링 패턴
- Istio 서비스 메시의 실제 동작 원리 이해
- 멀티클라우드 환경에서의 서비스 운영 노하우

## 참고사항
- 모든 설정 파일은 `/k8s/` 디렉토리에 위치
- 배포 스크립트: `./deploy-all.sh`
- UI는 ConfigMap을 통해 API Gateway에서 서빙
- **트래픽 추적**: 실제 Istio 라우팅 결과 기반 (시뮬레이션 아님)
- **교육 목적**: 서비스 메시 교육용 데모 애플리케이션
- **실제 검증**: 브라우저 개발자 콘솔에서 라우팅 로그 확인 가능