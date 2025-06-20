# Theater MSA - 운영 배포 진행 상황

## 📋 프로젝트 개요

**Theater MSA**는 교육용 마이크로서비스 아키텍처 애플리케이션으로, NaverCloud Platform(ctx1)과 NHN Cloud NKS(ctx2)의 멀티클라우드 Kubernetes 환경에서 Istio 서비스메시를 통해 운영되고 있습니다.

## 🏗️ 아키텍처 구성

### 마이크로서비스 분산 배치
```
┌─────────────────────────────────────────────────────────────┐
│                Istio 멀티클라우드 서비스메시                  │
├─────────────────────────────────────────────────────────────┤
│  NaverCloud Platform (ctx1)  │    NHN Cloud NKS (ctx2)      │
│  ┌─────────────────────────┐  │    ┌─────────────────────┐   │
│  │   User Service          │  │    │   Movie Service     │   │
│  │   API Gateway           │  │    │   Booking Service   │   │
│  │   Redis                 │  │    │   Redis             │   │
│  └─────────────────────────┘  │    └─────────────────────┘   │
│           │                   │              │               │
│    ┌─────────────┐            │       ┌─────────────┐        │
│    │ IngressGW   │◄───────────┼──────►│ IngressGW   │        │
│    └─────────────┘            │       └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
```

### 서비스 포트 구성
- **API Gateway**: 8080 (프론트엔드, UI 제공)
- **User Service**: 8081 (사용자 관리)
- **Movie Service**: 8082 (영화 정보 관리)
- **Booking Service**: 8083 (예약 관리)
- **Redis**: 6379 (데이터 저장소)

## 🚀 배포 현황

### ✅ 성공적으로 배포된 구성요소

#### 1. 기본 인프라
- **Kubernetes 클러스터**: ctx1(NaverCloud), ctx2(NHN Cloud) 정상 운영
- **네임스페이스**: `theater-msa` 생성 완료
- **Istio 서비스메시**: 양쪽 클러스터에 사전 설치된 Istio 활용
- **Harbor Registry**: `harbor.27.96.156.180.nip.io/theater-msa` 공개 저장소 구성

#### 2. 마이크로서비스 배포
- ✅ **User Service** (ctx1): 2개 복제본, 정상 운영
- ✅ **API Gateway** (ctx1): Kubernetes API 연동, UI ConfigMap 통합
- ✅ **Movie Service** (ctx2): 2개 복제본, 정상 운영
- ✅ **Booking Service** (ctx2): 2개 복제본, 정상 운영
- ✅ **Redis** (양쪽): 데이터 저장소 정상 운영

#### 3. 네트워킹 구성
- ✅ **cp-gateway**: 기존 Istio Gateway 활용 (HTTPS 강제)
- ✅ **VirtualService**: 도메인 기반 라우팅 설정
- ✅ **서비스 디스커버리**: Kubernetes 내부 DNS 정상 작동

#### 4. 애플리케이션 기능
- ✅ **UI 인터페이스**: 3열 반응형 레이아웃 (사용자/영화/예약)
- ✅ **실시간 배포 상태**: Kubernetes API를 통한 Pod/Node 정보 수집
- ✅ **데이터 초기화**: 자동 샘플 데이터 생성
- ✅ **예약 기능**: 좌석 선택 및 예약 처리

### 🔄 현재 진행 중인 이슈

#### 1. 멀티클러스터 통신 문제
**상태**: 🔴 해결 필요
**증상**: 
- User Service (ctx1): ✅ 정상 작동 (`https://theater.27.96.156.180.nip.io/users/`)
- Movie Service (ctx2): ❌ 503 Service Unavailable
- Booking Service (ctx2): ❌ 503 Service Unavailable

**진단 결과**:
```bash
# Pod 간 직접 통신 (성공)
kubectl exec api-gateway -- wget http://movie-service:8082/  # ✅ 정상 응답 []
kubectl exec api-gateway -- wget http://booking-service:8083/  # ✅ 정상 응답 []

# VirtualService를 통한 외부 접근 (실패)
curl https://theater.27.96.156.180.nip.io/movies/  # ❌ "no healthy upstream"
```

**근본 원인**:
- Istio가 멀티클러스터 서비스를 외부 인그레스(133.186.216.73:15443)를 통해 라우팅 시도
- EASTWESTGATEWAY 부재로 인한 라우팅 경로 문제
- VirtualService와 실제 멀티클러스터 서비스 디스커버리 간 불일치

### 📊 서비스 상태 요약

| 컴포넌트 | 클러스터 | Pod 상태 | 내부 통신 | 외부 접근 | 비고 |
|---------|---------|----------|-----------|-----------|------|
| User Service | ctx1 | ✅ Running(2/2) | ✅ 정상 | ✅ 정상 | 완전 정상 |
| API Gateway | ctx1 | ✅ Running(2/2) | ✅ 정상 | ✅ 정상 | UI 및 라우팅 정상 |
| Movie Service | ctx2 | ✅ Running(2/2) | ✅ 정상 | ❌ 503 오류 | 멀티클러스터 이슈 |
| Booking Service | ctx2 | ✅ Running(2/2) | ✅ 정상 | ❌ 503 오류 | 멀티클러스터 이슈 |
| Redis | ctx1/ctx2 | ✅ Running(2/2) | ✅ 정상 | N/A | 정상 운영 |

## 🛠️ 기술 스택

### 인프라
- **Container Runtime**: Podman (Harbor 이미지 빌드)
- **Orchestration**: Kubernetes 1.26+
- **Service Mesh**: Istio 1.26.0
- **Registry**: Harbor (공개 저장소)
- **SSL/TLS**: 자체 서명 인증서

### 애플리케이션
- **Backend**: Go 1.24
- **Frontend**: Vanilla JavaScript, CSS Grid
- **Database**: Redis
- **Build**: Multi-stage Docker builds

### 네트워킹
- **Ingress**: Istio IngressGateway
- **Load Balancer**: 클라우드 제공자 LB
- **DNS**: `theater.27.96.156.180.nip.io`
- **Protocol**: HTTPS (HTTP → HTTPS 리다이렉트)

## 🔧 배포 도구 및 스크립트

### 자동화 스크립트
- **`build-images.sh`**: 타임스탬프 기반 이미지 빌드 및 푸시
- **`update-deployment-images.sh`**: Deployment 이미지 태그 일괄 업데이트
- **`kustomization.yaml`**: 통합 배포 설정

### 설정 파일
- **ConfigMap**: UI 파일 통합 관리 (`ui-configmap.yaml`)
- **RBAC**: Kubernetes API 접근 권한 (`rbac.yaml`)
- **VirtualService**: Istio 라우팅 규칙
- **Affinity**: 클러스터별 서비스 배치 설정

## 📈 모니터링 및 관측성

### 현재 구현된 기능
- ✅ **실시간 Pod 상태**: Kubernetes API를 통한 Pod/Node 정보 수집
- ✅ **서비스 헬스체크**: HTTP 엔드포인트 응답 확인
- ✅ **배포 상태 대시보드**: 웹 UI에서 시각적 확인
- ✅ **로그 수집**: kubectl logs를 통한 실시간 로그 확인

### Istio 관측성 도구
- **Kiali**: 서비스 토폴로지 시각화 (ctx1 포트 32365)
- **Prometheus**: 메트릭 수집 (양쪽 클러스터)
- **Envoy 프록시**: 분산 추적 및 메트릭

## 🚨 알려진 문제 및 제한사항

### 1. 멀티클러스터 통신 이슈
- **문제**: VirtualService를 통한 ctx2 서비스 접근 실패
- **임시 해결책**: API Gateway를 통한 내부 프록시 (직접 통신은 정상)
- **영향도**: 높음 (웹 애플리케이션 기능 제한)

### 2. EASTWESTGATEWAY 부재
- **문제**: 표준 Istio 멀티클러스터 게이트웨이 미설치
- **현재 상태**: 인그레스 게이트웨이를 통한 우회 라우팅
- **영향도**: 중간 (성능 및 안정성 영향)

### 3. 이미지 캐시 문제
- **문제**: `latest` 태그 사용 시 이미지 업데이트 미반영
- **해결책**: 타임스탬프 기반 태그 시스템 도입
- **영향도**: 낮음 (배포 프로세스 개선됨)

## 📋 향후 작업 계획

### 단기 (1-2일)
1. **멀티클러스터 통신 복구**
   - EASTWESTGATEWAY 설치 또는 설정 점검
   - VirtualService 라우팅 규칙 최적화
   - 네트워크 정책 및 방화벽 설정 확인

2. **임시 해결책 적용** (필요시)
   - API Gateway 중심 라우팅으로 전환
   - 모든 요청을 ctx1으로 라우팅 후 내부 프록시

### 중기 (1주)
1. **모니터링 강화**
   - Grafana 대시보드 구성
   - 알림 시스템 설정
   - 성능 메트릭 수집

2. **보안 강화**
   - mTLS 정책 적용
   - 네트워크 정책 설정
   - 시크릿 관리 개선

### 장기 (1개월)
1. **자동화 개선**
   - CI/CD 파이프라인 구축
   - GitOps 워크플로우 도입
   - 자동 스케일링 설정

2. **고가용성 확보**
   - 다중 영역 배포
   - 백업 및 복구 절차
   - 재해 복구 계획

## 📞 연락처 및 리소스

### 접근 URL
- **프로덕션**: https://theater.27.96.156.180.nip.io
- **Harbor Registry**: https://harbor.27.96.156.180.nip.io
- **Kiali**: http://27.96.156.180.nip.io:32365

### 클러스터 정보
- **ctx1 (NaverCloud)**: `kubectl config use-context ctx1`
- **ctx2 (NHN Cloud)**: `kubectl config use-context ctx2`
- **네임스페이스**: `theater-msa`

### 주요 명령어
```bash
# 서비스 상태 확인
kubectl get pods -n theater-msa -o wide

# VirtualService 확인
kubectl get vs theater-msa -n istio-system

# 로그 확인
kubectl logs -n theater-msa -l app=api-gateway -f

# 멀티클러스터 통신 테스트
kubectl exec -n theater-msa deployment/api-gateway -- wget -qO- http://movie-service:8082/
```

---

**최종 업데이트**: 2025-06-19 21:30 KST  
**문서 버전**: v1.0  
**상태**: 부분 운영 (User Service 정상, Movie/Booking Service 이슈 해결 중)