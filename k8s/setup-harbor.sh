#!/bin/bash

# Harbor Registry 설정 스크립트
# 사용법: ./setup-harbor.sh [DOMAIN]
# 예시: ./setup-harbor.sh 27.96.156.180.nip.io

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

echo "=== Harbor Registry 설정 ==="
echo "Domain: ${DOMAIN}"
echo "Registry: harbor.${DOMAIN}"
echo ""

# 백업 생성
echo ">>> 백업 생성 중..."
cp kustomization.yaml kustomization.yaml.bak

# kustomization.yaml 도메인 치환
echo ">>> kustomization.yaml 업데이트 중..."
sed -i.tmp "s/{{DOMAIN}}/$DOMAIN/g" kustomization.yaml
rm -f kustomization.yaml.tmp

echo "  ✓ kustomization.yaml 업데이트 완료"
echo ""

echo "=== Harbor Registry 로그인 ==="
echo "다음 명령어로 Harbor에 로그인하세요:"
echo "docker login harbor.${DOMAIN}"
echo ""

echo "=== 이미지 빌드 및 배포 ==="
echo "1. 이미지 빌드:"
echo "   ./build-images.sh ${DOMAIN}"
echo ""
echo "2. 배포:"
echo "   kubectl apply -k ."
echo ""

echo "=== 복원하려면 ==="
echo "mv kustomization.yaml.bak kustomization.yaml"