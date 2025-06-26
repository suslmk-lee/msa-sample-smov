#!/bin/bash

# 노드 라벨 자동 설정 스크립트
# 사용법: ./setup-node-labels.sh

set -e

echo "=== Theater MSA 노드 라벨 설정 ==="

# CTX1 클러스터 노드 라벨 설정
echo ""
echo "1. CTX1 클러스터 노드 라벨 설정 중..."
if kubectl config use-context ctx1 2>/dev/null; then
    kubectl get nodes --no-headers | awk '{print $1}' | while read node; do
        echo "  노드 $node에 cluster-name=ctx1 라벨 설정"
        kubectl label nodes "$node" cluster-name=ctx1 --overwrite
    done
    echo "  ✓ CTX1 완료"
else
    echo "  ⚠ CTX1 컨텍스트를 찾을 수 없습니다"
fi

# CTX2 클러스터 노드 라벨 설정  
echo ""
echo "2. CTX2 클러스터 노드 라벨 설정 중..."
if kubectl config use-context ctx2 2>/dev/null; then
    kubectl get nodes --no-headers | awk '{print $1}' | while read node; do
        echo "  노드 $node에 cluster-name=ctx2 라벨 설정"
        kubectl label nodes "$node" cluster-name=ctx2 --overwrite
    done
    echo "  ✓ CTX2 완료"
else
    echo "  ⚠ CTX2 컨텍스트를 찾을 수 없습니다"
fi

# 결과 확인
echo ""
echo "=== 설정 결과 확인 ==="
for ctx in ctx1 ctx2; do
    if kubectl config use-context $ctx 2>/dev/null; then
        echo ""
        echo "[$ctx] 노드 상태:"
        kubectl get nodes -o custom-columns="NAME:.metadata.name,CLUSTER-NAME:.metadata.labels.cluster-name"
    fi
done

echo ""
echo "노드 라벨 설정 완료!"