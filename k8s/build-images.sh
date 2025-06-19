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

echo "=== Harbor Registry 이미지 빌드 시작 ==="
echo "Registry: ${HARBOR_REGISTRY}/${PROJECT_NAME}"
echo "Domain: ${DOMAIN}"
echo ""

# 상위 디렉토리로 이동 (소스 코드가 있는 위치)
cd ..

# 서비스별 이미지 빌드 및 푸시
SERVICES=("api-gateway" "user-service" "movie-service" "booking-service")

for SERVICE in "${SERVICES[@]}"; do
    echo ">>> 빌드 중: ${SERVICE}"
    
    # 이미지 태그 설정
    IMAGE_TAG="${HARBOR_REGISTRY}/${PROJECT_NAME}/${SERVICE}:latest"
    
    # Podman 빌드
    if [ -d "${SERVICE}" ]; then
        echo "  - 빌드: ${IMAGE_TAG}"
        podman build -t ${IMAGE_TAG} ./${SERVICE}/
        
        echo "  - 푸시: ${IMAGE_TAG}"
        podman push ${IMAGE_TAG}
        
        echo "  ✓ 완료: ${SERVICE}"
    else
        echo "  ⚠ 경고: ${SERVICE} 디렉토리를 찾을 수 없습니다"
    fi
    echo ""
done

echo "=== 빌드 완료 ==="
echo "다음 명령어로 이미지 태그를 업데이트하세요:"
echo ""
echo "cd k8s/"
for SERVICE in "${SERVICES[@]}"; do
    echo "kubectl set image deployment/${SERVICE} ${SERVICE}=${HARBOR_REGISTRY}/${PROJECT_NAME}/${SERVICE}:latest -n theater-msa"
done
echo ""
echo "또는 YAML 파일에서 이미지 태그를 직접 수정하세요:"
for SERVICE in "${SERVICES[@]}"; do
    echo "  image: ${HARBOR_REGISTRY}/${PROJECT_NAME}/${SERVICE}:latest"
done