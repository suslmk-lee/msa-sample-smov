#!/bin/bash

# 멀티클러스터 라우팅 문제 해결 스크립트

echo "=== 멀티클러스터 라우팅 문제 해결 ==="
echo ""

echo ">>> 1. API Gateway 직접 테스트"
for ctx in ctx1 ctx2; do
    echo ""
    echo "--- $ctx 클러스터 ---"
    kubectl config use-context $ctx
    
    # API Gateway 포트포워딩 테스트
    API_POD=$(kubectl get pods -n theater-msa -l app=api-gateway -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
    if [ -n "$API_POD" ]; then
        echo "API Gateway Pod 발견: $API_POD"
        echo "API Gateway 직접 테스트 (포트포워딩):"
        echo "다음 명령어를 별도 터미널에서 실행하세요:"
        echo "kubectl port-forward $API_POD 8080:8080 -n theater-msa --context=$ctx"
        echo "그 후 브라우저에서 http://localhost:8080 접속하여 테스트"
        echo ""
        
        # 내부 API 테스트
        echo "내부 API 테스트:"
        kubectl exec $API_POD -n theater-msa -- wget -qO- --timeout=5 "http://localhost:8080/" | head -5
        echo ""
        
        # 다른 서비스로의 API 호출 테스트
        echo "마이크로서비스 API 호출 테스트:"
        kubectl exec $API_POD -n theater-msa -- wget -qO- --timeout=5 "http://localhost:8080/users/" 2>/dev/null || echo "❌ /users/ 호출 실패"
        kubectl exec $API_POD -n theater-msa -- wget -qO- --timeout=5 "http://localhost:8080/movies/" 2>/dev/null || echo "❌ /movies/ 호출 실패"
        kubectl exec $API_POD -n theater-msa -- wget -qO- --timeout=5 "http://localhost:8080/bookings/" 2>/dev/null || echo "❌ /bookings/ 호출 실패"
    else
        echo "API Gateway Pod를 찾을 수 없습니다"
    fi
done

echo ""
echo ">>> 2. Istio 클러스터 간 연결 확인"
for ctx in ctx1 ctx2; do
    echo ""
    echo "--- $ctx 클러스터 Istio 설정 ---"
    kubectl config use-context $ctx
    
    echo "Istio Pilot discovery:"
    kubectl get pods -n istio-system -l app=istiod
    
    echo ""
    echo "Cross-network gateway:"
    kubectl get svc -n istio-system cross-network-gateway
    
    echo ""
    echo "Istio 멀티클러스터 서비스 discovery:"
    if command -v istioctl >/dev/null 2>&1; then
        istioctl proxy-config service api-gateway.theater-msa -n theater-msa 2>/dev/null || echo "istioctl 명령어 사용 불가"
    else
        echo "istioctl 명령어가 설치되지 않았습니다"
    fi
done

echo ""
echo ">>> 3. API Gateway 로그 상세 확인"
for ctx in ctx1 ctx2; do
    kubectl config use-context $ctx
    API_POD=$(kubectl get pods -n theater-msa -l app=api-gateway -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
    if [ -n "$API_POD" ]; then
        echo ""
        echo "--- $ctx 클러스터 API Gateway 로그 ---"
        echo "애플리케이션 로그:"
        kubectl logs $API_POD -c api-gateway -n theater-msa --tail=20
        echo ""
        echo "Istio Proxy 로그:"
        kubectl logs $API_POD -c istio-proxy -n theater-msa --tail=10
    fi
done

echo ""
echo ">>> 4. VirtualService 라우팅 테스트"
echo "현재 VirtualService 설정에서 API Gateway는 루트 경로(/)로 모든 트래픽을 받습니다."
echo "하지만 API Gateway 내부에서 다른 마이크로서비스 호출 시 멀티클러스터 연결이 필요합니다."

echo ""
echo ">>> 5. 문제 해결 방법"
echo ""
echo "방법 1: API Gateway 내부 서비스 호출 URL 확인"
echo "API Gateway Go 코드에서 다른 서비스 호출 시 사용하는 URL을 확인하세요:"
echo "- user-service.theater-msa.svc.cluster.local:8081"
echo "- movie-service.theater-msa.svc.cluster.local:8082" 
echo "- booking-service.theater-msa.svc.cluster.local:8083"
echo ""

echo "방법 2: 포트포워딩으로 직접 테스트"
echo "kubectl port-forward svc/api-gateway 8080:8080 -n theater-msa --context=ctx2"
echo "브라우저에서 http://localhost:8080 접속"
echo ""

echo "방법 3: ServiceEntry 추가 (멀티클러스터 서비스 등록)"
echo "각 클러스터에 원격 서비스들을 ServiceEntry로 등록할 수 있습니다."
echo ""

echo "=== 즉시 테스트 명령어 ==="
echo "다음 명령어로 API Gateway에 직접 접속하여 테스트하세요:"
echo ""
echo "# 터미널 1에서:"
echo "kubectl port-forward svc/api-gateway 8080:8080 -n theater-msa --context=ctx2"
echo ""
echo "# 터미널 2에서:"
echo "curl http://localhost:8080/"
echo "curl http://localhost:8080/users/"
echo "curl http://localhost:8080/movies/"
echo "curl http://localhost:8080/bookings/"