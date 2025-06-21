#!/bin/bash

# 장애 주입 교육 시연 스크립트
# K-PaaS 영화관 MSA 샘플 - Fault Injection Demo

set -e

# 색상 설정
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 로그 함수
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
}

info() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] INFO: $1${NC}"
}

# 도메인 설정
DOMAIN=${DOMAIN:-"27.96.156.180.nip.io"}
APP_URL="https://theater.$DOMAIN"

# 사용법
usage() {
    echo "사용법: $0 [COMMAND] [OPTIONS]"
    echo ""
    echo "COMMANDS:"
    echo "  setup           Circuit Breaker 및 Fault Injection 설정 배포"
    echo "  delay           Movie Service에 5초 지연 장애 주입"
    echo "  error           User Service에 50% HTTP 500 오류 주입"
    echo "  block           Booking Service CTX2 클러스터 차단"
    echo "  recover         모든 장애 복구 (원본 VirtualService 복원)"
    echo "  status          현재 장애 주입 상태 확인"
    echo "  test            장애 주입 테스트 (curl 요청)"
    echo "  cleanup         모든 Fault Injection 설정 제거"
    echo ""
    echo "OPTIONS:"
    echo "  --context CTX   kubectl context 지정 (기본값: ctx1)"
    echo "  --help          이 도움말 표시"
    echo ""
    echo "예시:"
    echo "  $0 setup"
    echo "  $0 delay"
    echo "  $0 test"
    echo "  $0 recover"
}

# kubectl context 설정
KUBECTL_CONTEXT="ctx1"

# 파라미터 파싱
while [[ $# -gt 0 ]]; do
    case $1 in
        --context)
            KUBECTL_CONTEXT="$2"
            shift 2
            ;;
        --help)
            usage
            exit 0
            ;;
        *)
            COMMAND="$1"
            shift
            ;;
    esac
done

# kubectl 명령어 래퍼
k() {
    kubectl --context="$KUBECTL_CONTEXT" "$@"
}

# Circuit Breaker 및 Fault Injection 설정 배포
setup_fault_injection() {
    log "Circuit Breaker 및 Fault Injection 설정을 배포합니다..."
    
    # Circuit Breaker 배포
    info "Circuit Breaker DestinationRule 배포 중..."
    k apply -f istio-circuit-breaker.yaml
    
    log "Fault Injection 설정이 준비되었습니다."
    info "사용 가능한 장애 시나리오:"
    echo "  - delay: Movie Service 지연 장애"
    echo "  - error: User Service HTTP 오류 장애"  
    echo "  - block: Booking Service 클러스터 차단"
}

# Movie Service CTX2 라우팅시에만 지연 주입
inject_delay_fault() {
    log "Movie Service CTX2 라우팅시에만 5초 지연 장애를 주입합니다..."
    
    # API Gateway에 DELAY_INJECTION_MODE 환경변수 설정
    k patch deployment api-gateway -n theater-msa --type='json' -p='[
      {
        "op": "add",
        "path": "/spec/template/spec/containers/0/env/-",
        "value": {
          "name": "DELAY_INJECTION_MODE",
          "value": "true"
        }
      }
    ]'
    
    # Pod 재시작 대기
    info "API Gateway Pod 재시작 중..."
    k rollout status deployment/api-gateway -n theater-msa --timeout=60s
    
    log "CTX2 라우팅 지연 장애가 주입되었습니다!"
    warn "CTX2(빨간색 신호등)로 라우팅되는 Movie 요청에만 5초 지연이 적용됩니다."
    info "웹 UI에서 영화 목록 로딩을 여러 번 시도해보세요:"
    echo "  - CTX1(녹색) 라우팅: 빠른 응답 (30% 확률)"
    echo "  - CTX2(빨간색) 라우팅: 5초 지연 (70% 확률)"
    echo "  - 신호등과 지연이 정확히 일치합니다!"
    echo "  - URL: $APP_URL"
}

# User Service HTTP 오류 장애 주입
inject_error_fault() {
    log "User Service에 50% HTTP 500 오류를 주입합니다..."
    
    # 기존 VirtualService 백업
    k get vs user-service-vs -n theater-msa -o yaml > /tmp/user-service-vs-backup.yaml
    
    # Fault Injection VirtualService 적용
    k apply -f - <<EOF
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: user-service-vs
  namespace: theater-msa
spec:
  hosts:
  - user-service
  http:
  - fault:
      abort:
        percentage:
          value: 50.0
        httpStatus: 500
    route:
    - destination:
        host: user-service
        subset: ctx1
      weight: 70
    - destination:
        host: user-service
        subset: ctx2
      weight: 30
EOF
    
    log "User Service HTTP 오류 장애가 주입되었습니다!"
    warn "User Service 요청의 50%가 HTTP 500 오류를 반환합니다."
    info "웹 UI에서 사용자 목록 로딩을 여러 번 시도해보세요: $APP_URL"
}

# Booking Service 클러스터 차단
inject_block_fault() {
    log "Booking Service CTX2 클러스터를 차단합니다..."
    
    # 기존 VirtualService 백업
    k get vs booking-service-vs -n theater-msa -o yaml > /tmp/booking-service-vs-backup.yaml
    
    # CTX1만 사용하도록 VirtualService 수정
    k apply -f - <<EOF
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: booking-service-vs
  namespace: theater-msa
spec:
  hosts:
  - booking-service
  http:
  - route:
    - destination:
        host: booking-service
        subset: ctx1
      weight: 100
EOF
    
    log "Booking Service CTX2 클러스터가 차단되었습니다!"
    warn "모든 Booking Service 트래픽이 CTX1으로만 라우팅됩니다."
    info "웹 UI에서 예약 목록의 신호등이 모두 녹색(CTX1)으로 변하는 것을 확인하세요: $APP_URL"
}

# 모든 장애 복구
recover_all_faults() {
    log "모든 장애를 복구하고 원본 설정을 복원합니다..."
    
    # 원본 내부 VirtualService 복원
    info "내부 VirtualService 복원 중..."
    k apply -f istio-virtualservices.yaml
    
    # API Gateway에서 DELAY_INJECTION_MODE 환경변수 제거
    info "API Gateway 지연 주입 모드 비활성화 중..."
    k patch deployment api-gateway -n theater-msa --type='json' -p='[
      {
        "op": "remove",
        "path": "/spec/template/spec/containers/0/env",
        "value": [
          {
            "name": "DELAY_INJECTION_MODE",
            "value": "true"
          }
        ]
      }
    ]' 2>/dev/null || true
    
    # 환경변수가 있는 경우 직접 제거
    k set env deployment/api-gateway -n theater-msa DELAY_INJECTION_MODE- 2>/dev/null || true
    
    # Pod 재시작 대기
    info "API Gateway Pod 재시작 중..."
    k rollout status deployment/api-gateway -n theater-msa --timeout=60s
    
    # 백업 파일 정리
    rm -f /tmp/*-vs-backup.yaml
    
    log "모든 장애가 복구되었습니다!"
    info "서비스가 정상 트래픽 분산으로 복원되었습니다."
    echo "  - User Service: 70% CTX1, 30% CTX2"
    echo "  - Movie Service: 30% CTX1, 70% CTX2"
    echo "  - Booking Service: 50% CTX1, 50% CTX2"
    echo "  - CTX2 라우팅 지연이 비활성화되었습니다"
}

# 현재 상태 확인
check_status() {
    log "현재 Fault Injection 상태를 확인합니다..."
    
    echo ""
    info "VirtualService 현황:"
    k get vs -n theater-msa
    
    echo ""
    info "DestinationRule 현황:"
    k get dr -n theater-msa
    
    echo ""
    info "Pod 상태:"
    k get pods -n theater-msa -o wide
}

# 장애 주입 테스트
test_fault_injection() {
    log "장애 주입 테스트를 실행합니다..."
    
    echo ""
    info "User Service 테스트 (5번 요청):"
    for i in {1..5}; do
        echo -n "  요청 $i: "
        if curl -k -s -o /dev/null -w "HTTP %{http_code} (%{time_total}s)" "$APP_URL/users/" 2>/dev/null; then
            echo ""
        else
            echo "요청 실패"
        fi
        sleep 1
    done
    
    echo ""
    info "Movie Service 테스트 (5번 요청) - 지연 테스트:"
    for i in {1..5}; do
        echo -n "  요청 $i: "
        if curl -k -s -o /dev/null -w "HTTP %{http_code} (%{time_total}s)" "$APP_URL/movies/" 2>/dev/null; then
            echo ""
        else
            echo "요청 실패"
        fi
        sleep 1
    done
    
    echo ""
    info "Booking Service 테스트 (3번 요청):"
    for i in {1..3}; do
        echo -n "  요청 $i: "
        if curl -k -s -o /dev/null -w "HTTP %{http_code} (%{time_total}s)" "$APP_URL/bookings/" 2>/dev/null; then
            echo ""
        else
            echo "요청 실패"
        fi
        sleep 1
    done
}

# 정리
cleanup_fault_injection() {
    log "모든 Fault Injection 설정을 제거합니다..."
    
    # 원본 VirtualService 복원
    k apply -f istio-virtualservices.yaml
    
    # Circuit Breaker DestinationRule 제거
    k delete -f istio-circuit-breaker.yaml --ignore-not-found=true
    
    # 백업 파일 정리
    rm -f /tmp/*-vs-backup.yaml
    
    log "모든 Fault Injection 설정이 제거되었습니다."
}

# 메인 실행
case "${COMMAND:-help}" in
    setup)
        setup_fault_injection
        ;;
    delay)
        inject_delay_fault
        ;;
    error)
        inject_error_fault
        ;;
    block)
        inject_block_fault
        ;;
    recover)
        recover_all_faults
        ;;
    status)
        check_status
        ;;
    test)
        test_fault_injection
        ;;
    cleanup)
        cleanup_fault_injection
        ;;
    help|*)
        usage
        ;;
esac