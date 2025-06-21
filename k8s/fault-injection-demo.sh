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
    echo "  circuit         Circuit Breaker 전용 테스트 (연속 오류 발생)"
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

# Movie Service 3초 지연 장애 주입
inject_delay_fault() {
    log "Movie Service에 3초 지연 장애를 주입합니다..."
    
    # 기존 VirtualService 백업
    k get vs movie-service-vs -n theater-msa -o yaml > /tmp/movie-service-vs-backup.yaml
    
    # 3초 지연 적용하는 VirtualService 적용
    k apply -f - <<EOF
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: movie-service-vs
  namespace: theater-msa
spec:
  hosts:
  - movie-service
  http:
  - match:
    - headers:
        x-canary:
          exact: "true"
    route:
    - destination:
        host: movie-service
        subset: ctx1
      weight: 100
  - fault:
      delay:
        percentage:
          value: 70.0
        fixedDelay: 3s
    route:
    - destination:
        host: movie-service
        subset: ctx1
      weight: 30
    - destination:
        host: movie-service
        subset: ctx2
      weight: 70
EOF
    
    log "Movie Service에 3초 지연 장애가 주입되었습니다!"
    warn "요청의 70%에 3초 지연이 적용됩니다."
    info "웹 UI에서 영화 목록 로딩을 여러 번 시도해보세요:"
    echo "  - 30% 확률: 빠른 응답"
    echo "  - 70% 확률: 3초 지연"
    echo "  - URL: $APP_URL"
}

# User Service HTTP 오류 장애 주입 (Circuit Breaker 트리거용)
inject_error_fault() {
    log "User Service CTX2에 100% HTTP 500 오류를 주입합니다 (Circuit Breaker 테스트)..."
    
    # 기존 VirtualService 백업
    k get vs user-service-vs -n theater-msa -o yaml > /tmp/user-service-vs-backup.yaml
    
    # CTX2에 집중된 트래픽으로 Circuit Breaker 트리거
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
  - match:
    - headers:
        x-canary:
          exact: "true"
    route:
    - destination:
        host: user-service
        subset: ctx1
      weight: 100
  - fault:
      abort:
        percentage:
          value: 100.0
        httpStatus: 500
    route:
    - destination:
        host: user-service
        subset: ctx2
      weight: 100
EOF
    
    log "User Service CTX2에 100% HTTP 500 오류가 주입되었습니다!"
    warn "모든 User Service 요청이 CTX2로 라우팅되어 100% 실패합니다."
    info "약 3-5회 요청 후 Circuit Breaker가 CTX2를 격리합니다."
    echo ""
    echo "테스트 방법:"
    echo "1. 웹 UI에서 User 섹션 새로고침을 연속으로 클릭"
    echo "2. 처음 3-5회는 모두 오류 발생"
    echo "3. Circuit Breaker 작동 후 30초간 모든 요청이 CTX1으로 우회"
    echo "4. URL: $APP_URL"
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
    log "모든 장애를 복구하고 원본 VirtualService를 복원합니다..."
    
    # 원본 VirtualService 복원
    info "원본 VirtualService 복원 중..."
    k apply -f istio-virtualservices.yaml
    
    # API Gateway 환경변수 정리 (있을 경우)
    k set env deployment/api-gateway -n theater-msa DELAY_INJECTION_MODE- 2>/dev/null || true
    
    # 백업 파일 정리
    rm -f /tmp/*-vs-backup.yaml
    
    log "모든 장애가 복구되었습니다!"
    info "서비스가 정상 트래픽 분산으로 복원되었습니다."
    echo "  - User Service: 70% CTX1, 30% CTX2"
    echo "  - Movie Service: 30% CTX1, 70% CTX2"
    echo "  - Booking Service: 50% CTX1, 50% CTX2"
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

# Circuit Breaker 전용 테스트 (확실한 트리거)
test_circuit_breaker() {
    log "Circuit Breaker 테스트를 위한 전용 시나리오를 실행합니다..."
    
    # 먼저 Circuit Breaker 설정 적용
    info "개선된 Circuit Breaker 설정 적용 중..."
    k apply -f istio-circuit-breaker.yaml
    
    # 잠시 대기 (설정 적용)
    sleep 5
    
    # User Service에 30% 오류율 적용으로 Circuit Breaker 테스트
    log "User Service에 30% 오류율을 적용하여 Circuit Breaker 테스트..."
    
    k get vs user-service-vs -n theater-msa -o yaml > /tmp/user-service-vs-backup.yaml
    
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
  - match:
    - headers:
        x-canary:
          exact: "true"
    route:
    - destination:
        host: user-service
        subset: ctx1
      weight: 100
  - fault:
      abort:
        percentage:
          value: 30.0
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
    
    log "Circuit Breaker 테스트 환경이 준비되었습니다!"
    echo ""
    warn "Circuit Breaker 동작 확인 방법:"
    echo "1. 웹 UI에서 User 섹션을 연속으로 10-20회 새로고침"
    echo "2. 처음에는 70% 성공, 30% 오류가 랜덤하게 발생"
    echo "3. 연속 2회 오류 발생시 Circuit Breaker 작동"
    echo "4. Circuit Breaker 작동 후: 30초간 모든 요청 성공 (오류 인스턴스 격리)"
    echo "5. 30초 후: 다시 테스트 시작 (복구 시도)"
    echo ""
    info "자동화된 테스트를 실행하시겠습니까? (y/n)"
    read -r response
    
    if [[ "$response" =~ ^[Yy]$ ]]; then
        info "자동화된 Circuit Breaker 테스트를 시작합니다..."
        
        echo ""
        echo "=== Circuit Breaker 트리거 테스트 (30% 오류율) ==="
        success_count=0
        error_count=0
        for i in {1..20}; do
            echo -n "요청 $i: "
            if response=$(curl -k -s -w "HTTP_%{http_code}_%{time_total}s" "$APP_URL/users/" 2>/dev/null); then
                if [[ "$response" == *"HTTP_200"* ]]; then
                    echo "$response ✓ (성공: $((++success_count)))"
                else
                    echo "$response ✗ (오류: $((++error_count)))"
                fi
            else
                echo "요청 실패 ✗ (오류: $((++error_count)))"
            fi
            sleep 1
        done
        
        echo ""
        echo "결과: 성공 $success_count회, 오류 $error_count회"
        
        echo ""
        echo "=== Circuit Breaker 복구 확인 (30초 후) ==="
        info "30초 대기 중... (Circuit Breaker 복구 시간)"
        sleep 30
        
        for i in {1..5}; do
            echo -n "복구 테스트 $i: "
            if response=$(curl -k -s -w "HTTP_%{http_code}_%{time_total}s" "$APP_URL/users/" 2>/dev/null); then
                if [[ "$response" == *"HTTP_200"* ]]; then
                    echo "$response ✓"
                else
                    echo "$response ✗"
                fi
            else
                echo "요청 실패 ✗"
            fi
            sleep 2
        done
    fi
    
    echo ""
    info "Circuit Breaker 테스트 완료!"
    echo "URL: $APP_URL"
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
    circuit)
        test_circuit_breaker
        ;;
    help|*)
        usage
        ;;
esac