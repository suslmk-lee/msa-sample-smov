# 극장 예매 MSA 애플리케이션 개발 계획

## Notes
- golang 기반의 MSA(마이크로서비스 아키텍처) 극장 예매 시스템을 개발
- UI(프론트엔드)와 여러 백엔드 마이크로서비스로 구성
- 단계별(기반 구축, 예매 기능, 고급 기능/배포)로 개발 진행
- DB는 사용하지 않음, 데이터 저장이 필요하면 Redis 사용
- Docker Compose는 사용하지 않음, Kubernetes에서 테스트/운영 예정
- 서비스별 go.mod 초기화 및 기본 main.go 작성 완료
- API Gateway, User Service, Movie Service 정상 라우팅 확인
- Redis 설치 및 서비스 실행 완료
- User Service의 Redis 연동 및 사용자 API 정상 동작 확인
- Movie Service의 Redis 연동 및 영화 API 정상 동작 확인
- UI 기본 파일(index.html, script.js, style.css) 생성 및 영화 목록 연동 확인
- Booking Service 디렉토리/Go 모듈 생성 및 API Gateway 라우팅 완료
- Booking Service: 예매 생성/조회 API 및 Redis 연동 완료
- 결제 서비스 개발은 교육용 샘플 범위로 진행하지 않음(중단)
- UI에서 예매 내역 조회 기능 구현 완료
- 서비스 코드 리팩토링(핸들러/모델/스토어 분리) 완료
- 각 서비스 및 Redis, 프론트엔드의 컨테이너화(Docker) 완료

## Task List
- [ ] 1단계: 기반 시스템 구축
  - [x] 프로젝트 구조 설정 (모노레포, 디렉토리 구성)
  - [x] API 게이트웨이 구축
  - [x] 사용자 서비스 개발
  - [x] 영화 서비스 개발
  - [x] 기본 UI 개발
  - [x] 예매 서비스 개발 (디렉토리/모듈/기본 서버 및 게이트웨이 라우팅)
  - [x] 예매 서비스 API 개발
- [ ] 2단계: 핵심 예매 기능 구현
  - [x] UI 예매 기능 확장(사용자 생성, 영화 선택, 예매 생성 등)
  - [x] UI: 예매 내역 조회 기능 추가
  - [x] 서비스 코드 리팩토링(핸들러/모델/스토어 분리)
- [ ] 3단계: 고급 기능 및 배포
  - [x] 컨테이너화(Docker)
  - [ ] 오케스트레이션(Kubernetes)
  - [ ] CI/CD 파이프라인 구축

## Current Goal
Kubernetes 오케스트레이션 환경 구성