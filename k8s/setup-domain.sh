#!/bin/bash

# Theater MSA Domain Setup Script
# 환경별 도메인 설정을 위한 스크립트

set -e

# 색상 코드 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 함수: 로그 출력
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 도메인 입력 받기
get_domain() {
    echo -e "${BLUE}============================================${NC}"
    echo -e "${BLUE}    Theater MSA Domain Configuration${NC}"
    echo -e "${BLUE}============================================${NC}"
    echo
    echo "환경별 도메인 설정 예시:"
    echo "1. NaverCloud + NHN Cloud: 27.96.156.180.nip.io"
    echo "2. Local/Other Cloud: <your-ip>.nip.io"
    echo "3. Custom Domain: example.com"
    echo "4. Local testing: k8s.local"
    echo
    
    read -p "사용할 도메인을 입력하세요 (예: 27.96.156.180.nip.io): " DOMAIN
    
    if [ -z "$DOMAIN" ]; then
        log_error "도메인이 입력되지 않았습니다."
        exit 1
    fi
    
    log_info "설정할 도메인: $DOMAIN"
    echo "최종 접근 URL: http://theater.$DOMAIN"
    echo
    read -p "계속하시겠습니까? (y/N): " confirm
    
    if [[ $confirm != [yY] ]]; then
        log_warning "설정이 취소되었습니다."
        exit 0
    fi
}

# VirtualService 파일 업데이트
update_virtualservice() {
    log_info "VirtualService 파일 업데이트 중..."
    
    # 백업 생성
    if [ ! -f "istio-virtualservice.yaml.backup" ]; then
        cp istio-virtualservice.yaml istio-virtualservice.yaml.backup
        log_info "백업 파일 생성: istio-virtualservice.yaml.backup"
    fi
    
    # 템플릿 치환
    sed -i.tmp "s/{{DOMAIN}}/$DOMAIN/g" istio-virtualservice.yaml
    rm -f istio-virtualservice.yaml.tmp
    
    log_success "VirtualService 업데이트 완료"
    
    # 결과 확인
    if grep -q "theater.$DOMAIN" istio-virtualservice.yaml; then
        log_success "도메인 설정 확인: theater.$DOMAIN"
    else
        log_error "도메인 설정 실패"
        exit 1
    fi
}

# DNS 해결 테스트
test_dns() {
    log_info "DNS 해결 테스트 중..."
    
    if nslookup "theater.$DOMAIN" >/dev/null 2>&1; then
        log_success "DNS 해결 성공: theater.$DOMAIN"
    else
        log_warning "DNS 해결 실패: theater.$DOMAIN"
        log_warning "이는 정상적일 수 있습니다 (클러스터 내부에서만 해결되는 경우)"
    fi
}

# 설정 복원
restore_config() {
    log_info "이전 설정으로 복원 중..."
    
    if [ -f "istio-virtualservice.yaml.backup" ]; then
        cp istio-virtualservice.yaml.backup istio-virtualservice.yaml
        log_success "설정이 복원되었습니다."
    else
        log_error "백업 파일이 없습니다."
        exit 1
    fi
}

# 도움말
show_help() {
    echo "Usage: $0 [OPTIONS]"
    echo
    echo "Options:"
    echo "  -d, --domain DOMAIN    도메인 직접 지정"
    echo "  -r, --restore         이전 설정으로 복원"
    echo "  -h, --help            도움말 표시"
    echo
    echo "Examples:"
    echo "  $0                              # 대화형 모드"
    echo "  $0 -d 27.96.156.180.nip.io     # 도메인 직접 지정"
    echo "  $0 --restore                    # 설정 복원"
}

# 메인 함수
main() {
    # 현재 디렉토리가 k8s 디렉토리인지 확인
    if [ ! -f "istio-virtualservice.yaml" ]; then
        log_error "istio-virtualservice.yaml 파일을 찾을 수 없습니다."
        log_error "k8s 디렉토리에서 실행하세요."
        exit 1
    fi
    
    # 파라미터 처리
    case "${1:-}" in
        -h|--help)
            show_help
            exit 0
            ;;
        -r|--restore)
            restore_config
            exit 0
            ;;
        -d|--domain)
            if [ -z "${2:-}" ]; then
                log_error "도메인을 지정하세요."
                exit 1
            fi
            DOMAIN="$2"
            log_info "지정된 도메인: $DOMAIN"
            ;;
        "")
            get_domain
            ;;
        *)
            log_error "알 수 없는 옵션: $1"
            show_help
            exit 1
            ;;
    esac
    
    # 도메인 설정 실행
    update_virtualservice
    test_dns
    
    echo
    log_success "도메인 설정이 완료되었습니다!"
    echo
    echo "다음 단계:"
    echo "1. kubectl apply -f istio-virtualservice.yaml"
    echo "2. 브라우저에서 접근: http://theater.$DOMAIN"
    echo "3. API 테스트: curl http://theater.$DOMAIN/users/"
    echo
}

# 스크립트 실행
main "$@"