#!/bin/bash

# "no healthy upstream" 문제 진단 스크립트

echo "=== Theater MSA Upstream 진단 ==="
echo "현재 시간: $(date)"
echo ""

echo "=== 1. Pod 상태 확인 ==="
echo ">>> 모든 클러스터에서 Pod 상태 확인"
for ctx in ctx1 ctx2; do
    echo ""
    echo "--- $ctx 클러스터 ---"
    kubectl config use-context $ctx 2>/dev/null
    if [ $? -eq 0 ]; then
        kubectl get pods -n theater-msa -o wide
        echo ""
        echo "Pod 상세 상태:"
        kubectl get pods -n theater-msa -o custom-columns="NAME:.metadata.name,STATUS:.status.phase,READY:.status.containerStatuses[*].ready,RESTARTS:.status.containerStatuses[*].restartCount"
    else
        echo "❌ $ctx 컨텍스트에 접근할 수 없습니다"
    fi
done

echo ""
echo "=== 2. Service 상태 확인 ==="
for ctx in ctx1 ctx2; do
    echo ""
    echo "--- $ctx 클러스터 Services ---"
    kubectl config use-context $ctx 2>/dev/null
    if [ $? -eq 0 ]; then
        kubectl get svc -n theater-msa
        echo ""
        echo "Service Endpoints:"
        kubectl get endpoints -n theater-msa
    fi
done

echo ""
echo "=== 3. VirtualService 설정 확인 ==="
for ctx in ctx1 ctx2; do
    echo ""
    echo "--- $ctx 클러스터 VirtualService ---"
    kubectl config use-context $ctx 2>/dev/null
    if [ $? -eq 0 ]; then
        kubectl get virtualservice -n istio-system theater-msa -o yaml 2>/dev/null || echo "VirtualService not found in $ctx"
    fi
done

echo ""
echo "=== 4. API Gateway 연결 테스트 ==="
for ctx in ctx1 ctx2; do
    echo ""
    echo "--- $ctx 클러스터 API Gateway 테스트 ---"
    kubectl config use-context $ctx 2>/dev/null
    if [ $? -eq 0 ]; then
        # API Gateway Pod 찾기
        API_POD=$(kubectl get pods -n theater-msa -l app=api-gateway -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
        if [ -n "$API_POD" ]; then
            echo "API Gateway Pod: $API_POD"
            echo "Pod 상태:"
            kubectl describe pod $API_POD -n theater-msa | grep -A 5 -B 5 "Status\|Ready\|Conditions"
            echo ""
            echo "컨테이너 로그 (최근 10줄):"
            kubectl logs $API_POD -n theater-msa --tail=10 2>/dev/null || echo "로그를 가져올 수 없습니다"
        else
            echo "API Gateway Pod를 찾을 수 없습니다"
        fi
    fi
done

echo ""
echo "=== 5. 서비스 간 연결 테스트 ==="
for ctx in ctx1 ctx2; do
    echo ""
    echo "--- $ctx 클러스터 내부 연결 테스트 ---"
    kubectl config use-context $ctx 2>/dev/null
    if [ $? -eq 0 ]; then
        # 테스트용 Pod가 있는지 확인
        TEST_POD=$(kubectl get pods -n theater-msa --field-selector=status.phase=Running -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
        if [ -n "$TEST_POD" ]; then
            echo "테스트 Pod: $TEST_POD"
            echo "API Gateway 서비스 연결 테스트:"
            kubectl exec $TEST_POD -n theater-msa -- wget -qO- --timeout=5 http://api-gateway:8080/ 2>/dev/null || echo "❌ API Gateway 연결 실패"
            echo "Redis 연결 테스트:"
            kubectl exec $TEST_POD -n theater-msa -- wget -qO- --timeout=5 http://redis:6379/ 2>/dev/null || echo "❌ Redis 연결 실패 (정상, HTTP가 아님)"
        else
            echo "연결 테스트를 위한 실행 중인 Pod가 없습니다"
        fi
    fi
done

echo ""
echo "=== 6. Istio Proxy 상태 확인 ==="
for ctx in ctx1 ctx2; do
    echo ""
    echo "--- $ctx 클러스터 Istio Proxy ---"
    kubectl config use-context $ctx 2>/dev/null
    if [ $? -eq 0 ]; then
        API_POD=$(kubectl get pods -n theater-msa -l app=api-gateway -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
        if [ -n "$API_POD" ]; then
            echo "API Gateway Pod의 Istio Proxy 상태:"
            kubectl exec $API_POD -c istio-proxy -n theater-msa -- pilot-agent request GET stats/ready 2>/dev/null || echo "Istio Proxy 상태 확인 실패"
        fi
    fi
done

echo ""
echo "=== 7. Gateway 설정 확인 ==="
kubectl config use-context ctx1 2>/dev/null || kubectl config use-context ctx2 2>/dev/null
echo "cp-gateway 설정:"
kubectl get gateway cp-gateway -n istio-system -o yaml 2>/dev/null || echo "cp-gateway를 찾을 수 없습니다"

echo ""
echo "=== 진단 완료 ==="
echo "문제 해결 권장사항:"
echo "1. Pod가 Running 상태가 아니라면: kubectl describe pod <pod-name> -n theater-msa"
echo "2. Service Endpoints가 비어있다면: Pod 라벨과 Service selector 확인"
echo "3. API Gateway 로그 확인: kubectl logs -f <api-gateway-pod> -n theater-msa"
echo "4. Istio sidecar 주입 확인: kubectl get pods -n theater-msa -o jsonpath='{.items[*].spec.containers[*].name}'"