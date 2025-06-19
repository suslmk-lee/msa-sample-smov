package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// DeploymentInfo represents service deployment information for Kubernetes
type DeploymentInfo struct {
	Service     string `json:"service"`
	Cluster     string `json:"cluster"`
	Namespace   string `json:"namespace"`
	PodName     string `json:"podName"`
	NodeName    string `json:"nodeName"`
	Status      string `json:"status"`
	Port        string `json:"port"`
	Icon        string `json:"icon"`
	LastChecked string `json:"lastChecked"`
}

var kubernetesClient *kubernetes.Clientset

func init() {
	// Kubernetes 클라이언트 초기화
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("Failed to get in-cluster config: %v", err)
		return
	}

	kubernetesClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("Failed to create Kubernetes client: %v", err)
		return
	}
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
	deploymentInfo := checkKubernetesDeploymentStatus()
	
	if err := json.NewEncoder(w).Encode(deploymentInfo); err != nil {
		log.Printf("Failed to encode deployment status: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// checkKubernetesDeploymentStatus checks the actual status of services in Kubernetes
func checkKubernetesDeploymentStatus() []DeploymentInfo {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	
	services := []struct {
		name      string
		url       string
		port      string
		icon      string
		labelApp  string
	}{
		{"API Gateway", "http://api-gateway:8080", "8080", "🌐", "api-gateway"},
		{"User Service", "http://user-service:8081", "8081", "👥", "user-service"},
		{"Movie Service", "http://movie-service:8082", "8082", "🎭", "movie-service"},
		{"Booking Service", "http://booking-service:8083", "8083", "📋", "booking-service"},
		{"Redis Cache", "redis:6379", "6379", "💾", "redis"},
	}
	
	var deploymentInfo []DeploymentInfo
	namespace := getNamespace()
	clusterName := getClusterName()
	
	for _, service := range services {
		status := "운영중"
		podName := ""
		nodeName := ""
		
		// Kubernetes에서 Pod 정보 수집
		if kubernetesClient != nil {
			if podInfo := getKubernetesPodInfo(namespace, service.labelApp); podInfo != nil {
				podName = podInfo.PodName
				nodeName = podInfo.NodeName
				if podInfo.Status != "Running" {
					status = "오류"
				}
			} else {
				status = "오류"
			}
		}
		
		// 헬스체크
		if status == "운영중" && !isServiceHealthy(service.url, service.name) {
			status = "오류"
		}
		
		deploymentInfo = append(deploymentInfo, DeploymentInfo{
			Service:     service.name,
			Cluster:     clusterName,
			Namespace:   namespace,
			PodName:     podName,
			NodeName:    nodeName,
			Status:      status,
			Port:        service.port,
			Icon:        service.icon,
			LastChecked: currentTime,
		})
	}
	
	return deploymentInfo
}

type PodInfo struct {
	PodName  string
	NodeName string
	Status   string
}

// getKubernetesPodInfo retrieves pod information from Kubernetes API
func getKubernetesPodInfo(namespace, labelApp string) *PodInfo {
	if kubernetesClient == nil {
		return nil
	}
	
	pods, err := kubernetesClient.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "app=" + labelApp,
	})
	
	if err != nil {
		log.Printf("Failed to get pods for %s: %v", labelApp, err)
		return nil
	}
	
	if len(pods.Items) == 0 {
		return nil
	}
	
	pod := pods.Items[0] // 첫 번째 Pod 사용
	return &PodInfo{
		PodName:  pod.Name,
		NodeName: pod.Spec.NodeName,
		Status:   string(pod.Status.Phase),
	}
}

// getNamespace returns the current namespace
func getNamespace() string {
	if ns := os.Getenv("POD_NAMESPACE"); ns != "" {
		return ns
	}
	return "theater-msa"
}

// getClusterName returns the cluster name from environment or node labels
func getClusterName() string {
	if cluster := os.Getenv("CLUSTER_NAME"); cluster != "" {
		return cluster
	}
	
	// 현재 Pod의 노드에서 클러스터 정보 추출
	if kubernetesClient != nil {
		if nodeName := os.Getenv("NODE_NAME"); nodeName != "" {
			node, err := kubernetesClient.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
			if err == nil {
				if clusterName, exists := node.Labels["cluster-name"]; exists {
					return clusterName
				}
			}
		}
	}
	
	return "Unknown Cluster"
}

// isServiceHealthy performs a simple health check
func isServiceHealthy(serviceURL, serviceName string) bool {
	// Redis는 다른 프로토콜이므로 스킵
	if serviceName == "Redis Cache" {
		return true
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