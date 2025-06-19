#!/bin/bash

# Harbor Registry 이미지 빌드 스크립트
# 사용법: ./build-images.sh [DOMAIN]
# 예시: ./build-images.sh 27.96.156.180.nip.io

set -e

# 도메인 설정
if [ -n "$1" ]; then
    DOMAIN="$1"
elif [ -n "$DOMAIN" ]; then
    echo "환경변수 DOMAIN 사용: $DOMAIN"
else
    read -p "Harbor 도메인을 입력하세요 (예: 27.96.156.180.nip.io): " DOMAIN
fi

if [ -z "$DOMAIN" ]; then
    echo "도메인이 필요합니다."
    exit 1
fi

HARBOR_REGISTRY="harbor.${DOMAIN}"
PROJECT_NAME="theater-msa"

# 현재 시간을 태그로 사용 (이미지 캐시 문제 해결)
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
IMAGE_TAG="v${TIMESTAMP}"

echo "=== Harbor Registry 이미지 빌드 시작 ==="
echo "Registry: ${HARBOR_REGISTRY}/${PROJECT_NAME}"
echo "Domain: ${DOMAIN}"
echo "Image Tag: ${IMAGE_TAG}"
echo ""

# 상위 디렉토리로 이동 (소스 코드가 있는 위치)
cd ..

# 서비스별 이미지 빌드 및 푸시
SERVICES=("api-gateway" "user-service" "movie-service" "booking-service")

for SERVICE in "${SERVICES[@]}"; do
    echo ">>> 빌드 중: ${SERVICE}"
    
    # 이미지 태그 설정 (타임스탬프와 latest 모두 생성)
    IMAGE_TAG_TIMESTAMP="${HARBOR_REGISTRY}/${PROJECT_NAME}/${SERVICE}:${IMAGE_TAG}"
    IMAGE_TAG_LATEST="${HARBOR_REGISTRY}/${PROJECT_NAME}/${SERVICE}:latest"
    
    # 서비스 디렉토리 경로 설정
    if [ "${SERVICE}" = "api-gateway" ]; then
        SERVICE_DIR="./${SERVICE}"
    else
        SERVICE_DIR="./services/${SERVICE}"
    fi
    
    # 컨테이너 런타임 자동 감지 및 빌드
    if [ -d "${SERVICE_DIR}" ]; then
        echo "  - 빌드: ${IMAGE_TAG_TIMESTAMP} and ${IMAGE_TAG_LATEST}"
        
        # 컨테이너 런타임 확인 (docker 또는 podman)
        if command -v docker >/dev/null 2>&1; then
            docker build -t ${IMAGE_TAG_TIMESTAMP} -t ${IMAGE_TAG_LATEST} ${SERVICE_DIR}/
            echo "  - 푸시: ${IMAGE_TAG_TIMESTAMP}"
            docker push ${IMAGE_TAG_TIMESTAMP}
            echo "  - 푸시: ${IMAGE_TAG_LATEST}"
            docker push ${IMAGE_TAG_LATEST}
        elif command -v podman >/dev/null 2>&1; then
            # Podman에서 docker.io 레지스트리 명시적 사용
            podman build --format docker -t ${IMAGE_TAG_TIMESTAMP} -t ${IMAGE_TAG_LATEST} ${SERVICE_DIR}/
            echo "  - 푸시: ${IMAGE_TAG_TIMESTAMP}"
            podman push ${IMAGE_TAG_TIMESTAMP}
            echo "  - 푸시: ${IMAGE_TAG_LATEST}"
            podman push ${IMAGE_TAG_LATEST}
        else
            echo "  ❌ 오류: Docker 또는 Podman이 설치되지 않았습니다"
            exit 1
        fi
        
        echo "  ✓ 완료: ${SERVICE}"
    else
        echo "  ⚠ 경고: ${SERVICE_DIR} 디렉토리를 찾을 수 없습니다"
    fi
    echo ""
done

echo "=== 빌드 완료 ==="
echo ""
echo "📝 생성된 이미지 태그: ${IMAGE_TAG}"
echo ""
echo "📝 배포 업데이트를 위해 다음 명령어를 실행하세요:"
echo "kubectl config use-context ctx1"
echo "kubectl patch deployment user-service -n theater-msa -p '{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"user-service\",\"image\":\"harbor.${DOMAIN}/theater-msa/user-service:${IMAGE_TAG}\"}]}}}}'"
echo "kubectl patch deployment api-gateway -n theater-msa -p '{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"api-gateway\",\"image\":\"harbor.${DOMAIN}/theater-msa/api-gateway:${IMAGE_TAG}\"}]}}}}'"
echo ""
echo "kubectl config use-context ctx2"
echo "kubectl patch deployment movie-service -n theater-msa -p '{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"movie-service\",\"image\":\"harbor.${DOMAIN}/theater-msa/movie-service:${IMAGE_TAG}\"}]}}}}'"
echo "kubectl patch deployment booking-service -n theater-msa -p '{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"booking-service\",\"image\":\"harbor.${DOMAIN}/theater-msa/booking-service:${IMAGE_TAG}\"}]}}}}'"