#!/bin/bash

# Podman Registry 설정 수정 스크립트
# Ubuntu에서 Podman의 registry 문제 해결

echo "=== Podman Registry 설정 수정 ==="

# registries.conf 파일 경로 확인
REGISTRIES_CONF="/etc/containers/registries.conf"

if [ ! -f "$REGISTRIES_CONF" ]; then
    echo "registries.conf 파일이 없습니다. 생성 중..."
    sudo mkdir -p /etc/containers
fi

# 백업 생성
if [ -f "$REGISTRIES_CONF" ]; then
    echo "기존 설정 백업 중..."
    sudo cp "$REGISTRIES_CONF" "${REGISTRIES_CONF}.backup.$(date +%Y%m%d_%H%M%S)"
fi

# 새로운 registries.conf 생성
echo "새로운 registries.conf 설정 적용 중..."
sudo tee "$REGISTRIES_CONF" > /dev/null <<EOF
# 기본 레지스트리 설정
unqualified-search-registries = ["docker.io"]

[[registry]]
prefix = "docker.io"
location = "docker.io"

[registries.search]
registries = ['docker.io']

[registries.insecure]
registries = []

[registries.block]
registries = []
EOF

echo "✓ registries.conf 설정 완료"

# shortnames.conf 설정
SHORTNAMES_CONF="/etc/containers/registries.conf.d/shortnames.conf"
echo "shortnames.conf 설정 중..."
sudo mkdir -p /etc/containers/registries.conf.d
sudo tee "$SHORTNAMES_CONF" > /dev/null <<EOF
# 공통 이미지 단축명 설정
[aliases]
"alpine" = "docker.io/library/alpine"
"golang" = "docker.io/library/golang"
"golang:1.21-alpine" = "docker.io/library/golang:1.21-alpine"
"redis" = "docker.io/library/redis"
"nginx" = "docker.io/library/nginx"
"ubuntu" = "docker.io/library/ubuntu"
EOF
echo "✓ shortnames.conf 설정 완료"

echo ""
echo "=== 설정 완료 ==="
echo "Podman 재시작 후 이미지 빌드를 다시 시도하세요:"
echo "sudo systemctl restart podman (시스템에 따라 다를 수 있음)"
echo "또는 단순히 새 터미널을 열어서 다시 시도하세요"
echo ""
echo "테스트:"
echo "podman pull docker.io/library/alpine:latest"
echo "podman pull docker.io/library/golang:1.21-alpine"