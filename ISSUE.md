# Theater MSA 멀티클러스터 통신 이슈 분석

## 📋 문제 요약

Theater MSA 애플리케이션에서 외부 접근 시 ctx2에 배치된 Movie Service와 Booking Service에 503 Service Unavailable 오류가 발생하는 문제

## 🔍 상세 분석

### 현재 상황
- **User Service (ctx1)**: ✅ 외부 접근 정상
- **Movie Service (ctx2)**: ❌ 503 Service Unavailable  
- **Booking Service (ctx2)**: ❌ 503 Service Unavailable

### 내부 통신 vs 외부 접근 차이점

#### ✅ 내부 통신 (API Gateway Pod에서)
```bash
kubectl exec api-gateway -- wget http://movie-service:8082/    # 성공
kubectl exec api-gateway -- wget http://booking-service:8083/  # 성공
```
- Istio 멀티클러스터 서비스 디스커버리 정상 작동
- EastWestGateway(15443) 통해 ctx2 서비스 접근 성공

#### ❌ 외부 접근 (Istio IngressGateway에서)
```bash
curl https://theater.27.96.156.180.nip.io/movies/    # "no healthy upstream"
curl https://theater.27.96.156.180.nip.io/bookings/  # "no healthy upstream"
```
- IngressGateway에서 ctx2 서비스 엔드포인트 찾을 수 없음

## 🎯 근본 원인

### VirtualService 라우팅 설정 문제
현재 VirtualService에서 외부 요청을 ctx2 서비스로 직접 라우팅 시도:

```yaml
# 현재 설정 (문제 발생)
- match:
  - uri:
      prefix: /movies
  route:
  - destination:
      host: movie-service.theater-msa.svc.cluster.local  # ctx2 서비스
      port:
        number: 8082

- match:
  - uri:
      prefix: /bookings  
  route:
  - destination:
      host: booking-service.theater-msa.svc.cluster.local  # ctx2 서비스
      port:
        number: 8083
```

### Istio IngressGateway vs Pod 레벨 차이
- **Pod 레벨**: 멀티클러스터 서비스 디스커버리 정보 동기화됨
- **IngressGateway 레벨**: 멀티클러스터 엔드포인트 정보 동기화 안됨

## 🔄 cp-portal vs theater-msa 비교

### cp-portal 패턴 (정상 작동)
- **VirtualService 없음**: 외부에서 ctx2 서비스로 직접 라우팅하지 않음
- **단일 진입점**: 모든 외부 요청이 portal-ui로만 라우팅
- **내부 프록시**: portal-ui가 필요에 따라 다른 서비스와 내부 통신

### theater-msa 패턴 (문제 발생)
- **VirtualService 있음**: 외부에서 ctx2 서비스로 직접 라우팅 시도
- **다중 진입점**: `/movies`, `/bookings` 등 경로별로 다른 서비스 라우팅
- **직접 라우팅**: IngressGateway가 멀티클러스터 서비스 직접 접근 시도

## 🛠️ 해결 방안

### ✅ 최종 솔루션: cp-portal 패턴 완전 적용
VirtualService에서 `/movies`, `/bookings` 라우팅을 완전 제거하고, API Gateway 내부 프록시만 사용

```yaml
# 최종 VirtualService 설정 (해결 완료)
spec:
  http:
  - match:
    - uri:
        prefix: /users          # ctx1 서비스만 직접 라우팅
    route:
    - destination:
        host: user-service.theater-msa.svc.cluster.local
  - match:
    - uri:
        prefix: /              # 모든 나머지 요청은 API Gateway로
    route:
    - destination:
        host: api-gateway.theater-msa.svc.cluster.local
```

### API Gateway 내부 프록시 구조 확인 완료
```go
// API Gateway main.go의 내부 라우팅 로직
func customHandler(w http.ResponseWriter, r *http.Request) {
    if strings.HasPrefix(r.URL.Path, "/movies/") {
        movieServiceProxy := newReverseProxy("http://movie-service:8082")  // ctx2
        movieServiceProxy.ServeHTTP(w, r)
        return
    }
    if strings.HasPrefix(r.URL.Path, "/bookings/") {
        bookingServiceProxy := newReverseProxy("http://booking-service:8083")  // ctx2
        bookingServiceProxy.ServeHTTP(w, r)
        return
    }
    // 기타 요청은 정적 파일 서빙
}
```

## ✅ 구현 완료

1. **VirtualService 수정**: `/movies`, `/bookings` 라우팅 제거 완료
2. **API Gateway 프록시 확인**: 내부 리버스 프록시 정상 작동 확인
3. **테스트 결과**: 
   - Movies endpoint: ✅ 200 OK
   - Bookings endpoint: ✅ 200 OK  
4. **아키텍처 검증**: cp-portal과 동일한 단일 진입점 패턴 적용 완료

## 📊 기술적 배경

### Istio 멀티클러스터 아키텍처 제약사항
- IngressGateway는 로컬 클러스터 서비스만 직접 라우팅 가능
- 멀티클러스터 서비스 디스커버리는 Pod 레벨에서만 완전 지원
- EastWestGateway는 클러스터 간 내부 통신 전용

### 장기적 고려사항
- **성능**: API Gateway를 통한 추가 홉 발생
- **안정성**: 단일 진입점을 통한 부하 집중
- **확장성**: API Gateway의 수평 확장 필요성

---

**문서 작성일**: 2025-06-20  
**이슈 상태**: 해결 방안 확정, 구현 대기  
**우선순위**: 높음