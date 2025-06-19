#!/bin/bash

# 서비스 재배포 스크립트 (cp-gateway 위치 고려)

echo "=== Theater MSA 서비스 재배포 ==="
echo "cp-gateway가 ctx1에 위치하므로 API Gateway도 ctx1로 이동"
echo ""

echo ">>> 1. 기존 배포 정리"
echo "ctx2에서 API Gateway 제거 중..."
kubectl config use-context ctx2
kubectl delete deployment api-gateway -n theater-msa --ignore-not-found=true
kubectl delete pod -l app=api-gateway -n theater-msa --ignore-not-found=true

echo ""
echo "ctx1에서 Booking Service 제거 중..."
kubectl config use-context ctx1
kubectl delete deployment booking-service -n theater-msa --ignore-not-found=true
kubectl delete pod -l app=booking-service -n theater-msa --ignore-not-found=true

echo ""
echo ">>> 2. VirtualService ctx2에서 제거"
kubectl config use-context ctx2
kubectl delete virtualservice theater-msa -n istio-system --ignore-not-found=true

echo ""
echo ">>> 3. 새로운 배포"
echo ""
echo "ctx1 클러스터 배포 (User Service + API Gateway):"
kubectl config use-context ctx1
kubectl apply -f namespace.yaml
kubectl apply -f redis.yaml
kubectl apply -f user-service.yaml
kubectl apply -f api-gateway.yaml
kubectl apply -f istio-virtualservice.yaml
kubectl apply -f istio-destinationrule.yaml

echo ""
echo "ctx2 클러스터 배포 (Movie Service + Booking Service):"
kubectl config use-context ctx2
kubectl apply -f namespace.yaml
kubectl apply -f redis.yaml
kubectl apply -f movie-service.yaml
kubectl apply -f booking-service.yaml
kubectl apply -f istio-destinationrule.yaml

echo ""
echo ">>> 4. 배포 상태 확인"
echo ""
echo "ctx1 클러스터:"
kubectl config use-context ctx1
kubectl get pods -n theater-msa
echo ""
echo "VirtualService 확인:"
kubectl get virtualservice -n istio-system theater-msa

echo ""
echo "ctx2 클러스터:"
kubectl config use-context ctx2
kubectl get pods -n theater-msa

echo ""
echo ">>> 5. 접속 테스트"
echo "약 30초 후 다음 URL로 접속 테스트:"
echo "https://theater.27.96.156.180.nip.io"
echo ""
echo "또는 직접 포트포워딩 테스트:"
echo "kubectl port-forward svc/api-gateway 8080:8080 -n theater-msa --context=ctx1"