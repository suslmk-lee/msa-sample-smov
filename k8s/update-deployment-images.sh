#!/bin/bash

# Deployment YAML 파일의 이미지 태그를 Harbor Registry로 일괄 변경하는 스크립트
# 사용법: ./update-deployment-images.sh [DOMAIN]
# 예시: ./update-deployment-images.sh 27.96.156.180.nip.io

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

echo "=== Deployment YAML 이미지 태그 업데이트 ==="
echo "Registry: ${HARBOR_REGISTRY}/${PROJECT_NAME}"
echo "Domain: ${DOMAIN}"
echo ""

# 백업 생성
echo ">>> 백업 생성 중..."
for SERVICE in api-gateway user-service movie-service booking-service; do
    if [ -f "${SERVICE}.yaml" ]; then
        cp "${SERVICE}.yaml" "${SERVICE}.yaml.bak"
        echo "  ✓ ${SERVICE}.yaml 백업 완료"
    fi
done

echo ""
echo ">>> 이미지 태그 업데이트 중..."

# API Gateway 이미지 업데이트
if [ -f "api-gateway.yaml" ]; then
    sed -i.tmp "s|image: api-gateway:latest|image: ${HARBOR_REGISTRY}/${PROJECT_NAME}/api-gateway:latest|g" api-gateway.yaml
    sed -i.tmp "s|image: api-gateway:.*|image: ${HARBOR_REGISTRY}/${PROJECT_NAME}/api-gateway:latest|g" api-gateway.yaml
    echo "  ✓ api-gateway.yaml 업데이트 완료"
fi

# User Service 이미지 업데이트
if [ -f "user-service.yaml" ]; then
    sed -i.tmp "s|image: user-service:latest|image: ${HARBOR_REGISTRY}/${PROJECT_NAME}/user-service:latest|g" user-service.yaml
    sed -i.tmp "s|image: user-service:.*|image: ${HARBOR_REGISTRY}/${PROJECT_NAME}/user-service:latest|g" user-service.yaml
    echo "  ✓ user-service.yaml 업데이트 완료"
fi

# Movie Service 이미지 업데이트
if [ -f "movie-service.yaml" ]; then
    sed -i.tmp "s|image: movie-service:latest|image: ${HARBOR_REGISTRY}/${PROJECT_NAME}/movie-service:latest|g" movie-service.yaml
    sed -i.tmp "s|image: movie-service:.*|image: ${HARBOR_REGISTRY}/${PROJECT_NAME}/movie-service:latest|g" movie-service.yaml
    echo "  ✓ movie-service.yaml 업데이트 완료"
fi

# Booking Service 이미지 업데이트
if [ -f "booking-service.yaml" ]; then
    sed -i.tmp "s|image: booking-service:latest|image: ${HARBOR_REGISTRY}/${PROJECT_NAME}/booking-service:latest|g" booking-service.yaml
    sed -i.tmp "s|image: booking-service:.*|image: ${HARBOR_REGISTRY}/${PROJECT_NAME}/booking-service:latest|g" booking-service.yaml
    echo "  ✓ booking-service.yaml 업데이트 완료"
fi

# 임시 파일 정리
rm -f *.tmp

echo ""
echo "=== 업데이트 완료 ==="
echo "변경된 이미지 태그:"
for SERVICE in api-gateway user-service movie-service booking-service; do
    echo "  - ${SERVICE}: ${HARBOR_REGISTRY}/${PROJECT_NAME}/${SERVICE}:latest"
done

echo ""
echo "백업 파일:"
for SERVICE in api-gateway user-service movie-service booking-service; do
    if [ -f "${SERVICE}.yaml.bak" ]; then
        echo "  - ${SERVICE}.yaml.bak"
    fi
done

echo ""
echo "복원하려면:"
echo "for f in *.yaml.bak; do mv \"\$f\" \"\${f%.bak}\"; done"
echo ""
echo "배포하려면:"
echo "kubectl apply -k ."
echo ""
echo "개별 배포:"
for SERVICE in api-gateway user-service movie-service booking-service; do
    echo "kubectl apply -f ${SERVICE}.yaml"
done