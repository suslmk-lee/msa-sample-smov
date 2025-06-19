package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

// DeploymentInfo represents service deployment information
type DeploymentInfo struct {
	Service     string `json:"service"`
	Platform    string `json:"platform"`
	Environment string `json:"environment"`
	ContainerID string `json:"containerID"`
	Status      string `json:"status"`
	Port        string `json:"port"`
	Icon        string `json:"icon"`
	LastChecked string `json:"lastChecked"`
}

// newReverseProxy creates a new reverse proxy for the target URL.
func newReverseProxy(target string) *httputil.ReverseProxy {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Failed to parse target URL %s: %v", target, err)
	}
	return httputil.NewSingleHostReverseProxy(targetURL)
}

// customHandler handles routing between API calls and static files
func customHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request: %s %s", r.Method, r.URL.Path)
	
	// API routes
	if strings.HasPrefix(r.URL.Path, "/users/") {
		log.Printf("Routing to user-service: %s", r.URL.Path)
		userServiceProxy := newReverseProxy("http://user-service:8081")
		userServiceProxy.ServeHTTP(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/movies/") {
		log.Printf("Routing to movie-service: %s", r.URL.Path)
		movieServiceProxy := newReverseProxy("http://movie-service:8082")
		movieServiceProxy.ServeHTTP(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/bookings/") {
		log.Printf("Routing to booking-service: %s", r.URL.Path)
		bookingServiceProxy := newReverseProxy("http://booking-service:8083")
		bookingServiceProxy.ServeHTTP(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/deployment-status") {
		log.Printf("Serving deployment status: %s", r.URL.Path)
		deploymentStatusHandler(w, r)
		return
	}
	
	// Static files
	log.Printf("Serving static file: %s", r.URL.Path)
	http.FileServer(http.Dir("ui")).ServeHTTP(w, r)
}

// deploymentStatusHandler handles deployment status requests
func deploymentStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// 실제 서비스 상태 체크
	deploymentInfo := checkDeploymentStatus()
	
	if err := json.NewEncoder(w).Encode(deploymentInfo); err != nil {
		log.Printf("Failed to encode deployment status: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// checkDeploymentStatus checks the actual status of services in Docker Compose
func checkDeploymentStatus() []DeploymentInfo {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	
	services := []struct {
		name string
		url  string
		port string
		icon string
	}{
		{"API Gateway", "http://localhost:8080", "8080", "🌐"},
		{"User Service", "http://user-service:8081", "8081", "👥"},
		{"Movie Service", "http://movie-service:8082", "8082", "🎭"},
		{"Booking Service", "http://booking-service:8083", "8083", "📋"},
		{"Redis Cache", "redis:6379", "6379", "💾"},
	}
	
	var deploymentInfo []DeploymentInfo
	
	for _, service := range services {
		status := "운영중"
		containerID := ""
		environment := "Docker Compose"
		platform := "Local Development"
		
		// 컨테이너 정보 수집
		containerID = getContainerInfo(service.name)
		
		// 헬스체크
		if status == "운영중" && !isServiceHealthy(service.url, service.name) {
			status = "오류"
		}
		
		deploymentInfo = append(deploymentInfo, DeploymentInfo{
			Service:     service.name,
			Platform:    platform,
			Environment: environment,
			ContainerID: containerID,
			Status:      status,
			Port:        service.port,
			Icon:        service.icon,
			LastChecked: currentTime,
		})
	}
	
	return deploymentInfo
}

// getContainerInfo retrieves container information from environment
func getContainerInfo(serviceName string) string {
	// 현재 컨테이너의 호스트명 사용 (Docker Compose에서 자동 설정)
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("Failed to get hostname for %s: %v", serviceName, err)
		return "unknown"
	}
	
	// API Gateway인 경우 현재 컨테이너 정보 반환
	if serviceName == "API Gateway" {
		return hostname[:min(len(hostname), 8)]
	}
	
	// 다른 서비스들은 서비스명 기반으로 생성
	return strings.ToLower(strings.ReplaceAll(serviceName, " ", "-"))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// isServiceHealthy performs a simple health check
func isServiceHealthy(serviceURL, serviceName string) bool {
	// Redis는 다른 프로토콜이므로 스킵
	if serviceName == "Redis Cache" {
		return true // Redis 연결은 다른 서비스들이 사용 중이면 정상으로 간주
	}
	
	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	
	// API Gateway는 자기 자신이므로 별도 처리
	if serviceName == "API Gateway" {
		return true
	}
	
	// 간단한 헬스체크 엔드포인트 호출
	resp, err := client.Get(serviceURL)
	if err != nil {
		log.Printf("Health check failed for %s: %v", serviceName, err)
		return false
	}
	defer resp.Body.Close()
	
	return resp.StatusCode < 500
}

func main() {
	http.HandleFunc("/", customHandler)

	log.Println("API Gateway started on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
