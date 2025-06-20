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

### 배포 현황
- **API Gateway**: CTX1에만 배포 (정상)
- **User/Movie/Booking Services**: CTX1, CTX2 모두 배포 (정상)
- **트래픽 시각화**: 실시간 신호등으로 트래픽 분산 상태 표시
- **설정 연동**: 실제 VirtualService 가중치 값 표시

### 성능 최적화
1. **데이터 초기화 최적화**: 중복 데이터 생성 방지
2. **병렬 요청 처리**: Promise.all로 동시 요청
3. **UI 응답성 개선**: 불필요한 DOM 조작 최소화

### 다음 단계 계획
- [ ] Movie Service, Booking Service 트래픽 시각화 추가
- [ ] 실시간 모니터링 대시보드 구현
- [ ] 클러스터 상태 모니터링 기능 추가
- [ ] 성능 메트릭 수집 및 표시

## 기술 스택
- **컨테이너**: Docker, Kubernetes
- **서비스 메시**: Istio (VirtualService, DestinationRule)
- **프론트엔드**: HTML5, CSS3, Vanilla JavaScript
- **백엔드**: Go (API Gateway), Node.js (Services)
- **클라우드**: NaverCloud, NHN Cloud
- **모니터링**: Kubernetes API 기반 실시간 상태 확인

## 참고사항
- 모든 설정 파일은 `/k8s/` 디렉토리에 위치
- 배포 스크립트: `./deploy-all.sh`
- UI는 ConfigMap을 통해 API Gateway에서 서빙
- 트래픽 분산은 API Gateway에서 가중치 기반 라우팅으로 구현