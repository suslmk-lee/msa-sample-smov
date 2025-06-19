#!/bin/bash

# Harbor Registry 이미지 태그 업데이트 스크립트
# 사용법: ./update-images.sh [DOMAIN]
# 예시: ./update-images.sh 27.96.156.180.nip.io

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

echo "=== YAML 파일 이미지 태그 업데이트 ==="
echo "Registry: ${HARBOR_REGISTRY}/${PROJECT_NAME}"
echo ""

# 백업 생성
echo ">>> 백업 생성 중..."
cp api-gateway.yaml api-gateway.yaml.bak
cp user-service.yaml user-service.yaml.bak
cp movie-service.yaml movie-service.yaml.bak
cp booking-service.yaml booking-service.yaml.bak

# 이미지 태그 업데이트
echo ">>> 이미지 태그 업데이트 중..."

# API Gateway
sed -i.tmp "s|image: api-gateway:latest|image: ${HARBOR_REGISTRY}/${PROJECT_NAME}/api-gateway:latest|g" api-gateway.yaml
echo "  ✓ api-gateway.yaml 업데이트 완료"

# User Service
sed -i.tmp "s|image: user-service:latest|image: ${HARBOR_REGISTRY}/${PROJECT_NAME}/user-service:latest|g" user-service.yaml
echo "  ✓ user-service.yaml 업데이트 완료"

# Movie Service
sed -i.tmp "s|image: movie-service:latest|image: ${HARBOR_REGISTRY}/${PROJECT_NAME}/movie-service:latest|g" movie-service.yaml
echo "  ✓ movie-service.yaml 업데이트 완료"

# Booking Service
sed -i.tmp "s|image: booking-service:latest|image: ${HARBOR_REGISTRY}/${PROJECT_NAME}/booking-service:latest|g" booking-service.yaml
echo "  ✓ booking-service.yaml 업데이트 완료"

# 임시 파일 정리
rm -f *.tmp

echo ""
echo "=== 업데이트 완료 ==="
echo "백업 파일:"
echo "  - api-gateway.yaml.bak"
echo "  - user-service.yaml.bak" 
echo "  - movie-service.yaml.bak"
echo "  - booking-service.yaml.bak"
echo ""
echo "복원하려면: mv *.bak을 원본 파일명으로 복사하세요"
echo ""
echo "배포: kubectl apply -k ."