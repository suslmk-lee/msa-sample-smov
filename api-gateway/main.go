package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"log"
	"math/big"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	istioclient "istio.io/client-go/pkg/clientset/versioned"
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

// TrafficWeight represents service traffic distribution
type TrafficWeight struct {
	UserServiceCtx1Weight    int
	UserServiceCtx2Weight    int
	MovieServiceCtx1Weight   int
	MovieServiceCtx2Weight   int
	BookingServiceCtx1Weight int
	BookingServiceCtx2Weight int
}

// TrafficHistory represents recent traffic routing decisions
type TrafficHistory struct {
	UserServiceHistory    []string `json:"userServiceHistory"`
	MovieServiceHistory   []string `json:"movieServiceHistory"`
	BookingServiceHistory []string `json:"bookingServiceHistory"`
}

var kubernetesClient *kubernetes.Clientset
var istioClient *istioclient.Clientset
var trafficWeights TrafficWeight
var trafficHistory TrafficHistory
var maxHistorySize = 10

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

	// Istio 클라이언트 초기화
	istioClient, err = istioclient.NewForConfig(config)
	if err != nil {
		log.Printf("Failed to create Istio client: %v", err)
		return
	}

	// 트래픽 가중치 초기화 (환경변수 또는 기본값)
	trafficWeights = TrafficWeight{
		UserServiceCtx1Weight:    getEnvInt("USER_SERVICE_CTX1_WEIGHT", 70),
		UserServiceCtx2Weight:    getEnvInt("USER_SERVICE_CTX2_WEIGHT", 30),
		MovieServiceCtx1Weight:   getEnvInt("MOVIE_SERVICE_CTX1_WEIGHT", 30),
		MovieServiceCtx2Weight:   getEnvInt("MOVIE_SERVICE_CTX2_WEIGHT", 70),
		BookingServiceCtx1Weight: getEnvInt("BOOKING_SERVICE_CTX1_WEIGHT", 50),
		BookingServiceCtx2Weight: getEnvInt("BOOKING_SERVICE_CTX2_WEIGHT", 50),
	}

	log.Printf("Traffic weights initialized: %+v", trafficWeights)
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getVirtualServiceWeights reads actual weights from VirtualService resources
func getVirtualServiceWeights() TrafficWeight {
	weights := TrafficWeight{
		UserServiceCtx1Weight:    70, // 기본값
		UserServiceCtx2Weight:    30,
		MovieServiceCtx1Weight:   30,
		MovieServiceCtx2Weight:   70,
		BookingServiceCtx1Weight: 50,
		BookingServiceCtx2Weight: 50,
	}

	if istioClient == nil {
		log.Printf("Istio client not available, using default weights")
		return weights
	}

	// User Service VirtualService 조회
	if userVS, err := istioClient.NetworkingV1().VirtualServices("theater-msa").Get(context.TODO(), "user-service-vs", metav1.GetOptions{}); err == nil {
		if len(userVS.Spec.Http) > 1 && len(userVS.Spec.Http[1].Route) >= 2 {
			// 카나리가 아닌 일반 라우팅 규칙에서 가중치 추출
			for _, route := range userVS.Spec.Http[1].Route {
				if route.Destination.Subset == "ctx1" {
					weights.UserServiceCtx1Weight = int(route.Weight)
				} else if route.Destination.Subset == "ctx2" {
					weights.UserServiceCtx2Weight = int(route.Weight)
				}
			}
			log.Printf("User service weights from VirtualService: ctx1=%d, ctx2=%d", weights.UserServiceCtx1Weight, weights.UserServiceCtx2Weight)
		}
	} else {
		log.Printf("Failed to get user-service-vs: %v", err)
	}

	// Movie Service VirtualService 조회
	if movieVS, err := istioClient.NetworkingV1().VirtualServices("theater-msa").Get(context.TODO(), "movie-service-vs", metav1.GetOptions{}); err == nil {
		if len(movieVS.Spec.Http) > 1 && len(movieVS.Spec.Http[1].Route) >= 2 {
			for _, route := range movieVS.Spec.Http[1].Route {
				if route.Destination.Subset == "ctx1" {
					weights.MovieServiceCtx1Weight = int(route.Weight)
				} else if route.Destination.Subset == "ctx2" {
					weights.MovieServiceCtx2Weight = int(route.Weight)
				}
			}
			log.Printf("Movie service weights from VirtualService: ctx1=%d, ctx2=%d", weights.MovieServiceCtx1Weight, weights.MovieServiceCtx2Weight)
		}
	} else {
		log.Printf("Failed to get movie-service-vs: %v", err)
	}

	// Booking Service VirtualService 조회
	if bookingVS, err := istioClient.NetworkingV1().VirtualServices("theater-msa").Get(context.TODO(), "booking-service-vs", metav1.GetOptions{}); err == nil {
		if len(bookingVS.Spec.Http) > 1 && len(bookingVS.Spec.Http[1].Route) >= 2 {
			for _, route := range bookingVS.Spec.Http[1].Route {
				if route.Destination.Subset == "ctx1" {
					weights.BookingServiceCtx1Weight = int(route.Weight)
				} else if route.Destination.Subset == "ctx2" {
					weights.BookingServiceCtx2Weight = int(route.Weight)
				}
			}
			log.Printf("Booking service weights from VirtualService: ctx1=%d, ctx2=%d", weights.BookingServiceCtx1Weight, weights.BookingServiceCtx2Weight)
		}
	} else {
		log.Printf("Failed to get booking-service-vs: %v", err)
	}

	return weights
}

// addToHistory adds a cluster selection to the history for a specific service
func addToHistory(serviceType string, cluster string) {
	switch serviceType {
	case "user":
		trafficHistory.UserServiceHistory = append(trafficHistory.UserServiceHistory, cluster)
		if len(trafficHistory.UserServiceHistory) > maxHistorySize {
			trafficHistory.UserServiceHistory = trafficHistory.UserServiceHistory[1:]
		}
	case "movie":
		trafficHistory.MovieServiceHistory = append(trafficHistory.MovieServiceHistory, cluster)
		if len(trafficHistory.MovieServiceHistory) > maxHistorySize {
			trafficHistory.MovieServiceHistory = trafficHistory.MovieServiceHistory[1:]
		}
	case "booking":
		trafficHistory.BookingServiceHistory = append(trafficHistory.BookingServiceHistory, cluster)
		if len(trafficHistory.BookingServiceHistory) > maxHistorySize {
			trafficHistory.BookingServiceHistory = trafficHistory.BookingServiceHistory[1:]
		}
	}
}

// weightedServiceSelect selects service based on weight and records the decision
func weightedServiceSelect(serviceType string, ctx1Weight, ctx2Weight int, ctx1Service, ctx2Service string) string {
	total := ctx1Weight + ctx2Weight
	if total == 0 {
		addToHistory(serviceType, "ctx1")
		return ctx1Service // fallback
	}

	// Generate random number between 0 and total-1
	randomNum, err := rand.Int(rand.Reader, big.NewInt(int64(total)))
	if err != nil {
		log.Printf("Failed to generate random number, falling back to ctx1: %v", err)
		addToHistory(serviceType, "ctx1")
		return ctx1Service
	}

	if randomNum.Int64() < int64(ctx1Weight) {
		log.Printf("Selected %s (weight: %d/%d)", ctx1Service, ctx1Weight, total)
		addToHistory(serviceType, "ctx1")
		return ctx1Service
	} else {
		log.Printf("Selected %s (weight: %d/%d)", ctx2Service, ctx2Weight, total)
		addToHistory(serviceType, "ctx2")
		return ctx2Service
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

// customHandler handles routing between API calls and static files with weighted distribution
func customHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request: %s %s", r.Method, r.URL.Path)
	
	// API routes with weighted distribution
	if strings.HasPrefix(r.URL.Path, "/users/") {
		log.Printf("Routing to user-service via Istio VirtualService")
		proxy := newReverseProxy("http://user-service:8081")
		proxy.ServeHTTP(w, r)
		return
	}
	
	if strings.HasPrefix(r.URL.Path, "/movies/") {
		log.Printf("Routing to movie-service via Istio VirtualService")
		proxy := newReverseProxy("http://movie-service:8082")
		proxy.ServeHTTP(w, r)
		return
	}
	
	if strings.HasPrefix(r.URL.Path, "/bookings/") {
		log.Printf("Routing to booking-service via Istio VirtualService")
		proxy := newReverseProxy("http://booking-service:8083")
		proxy.ServeHTTP(w, r)
		return
	}
	
	if strings.HasPrefix(r.URL.Path, "/deployment-status") {
		log.Printf("Serving deployment status: %s", r.URL.Path)
		getDeploymentStatus(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/traffic-weights") {
		log.Printf("Serving traffic weights: %s", r.URL.Path)
		getTrafficWeights(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/traffic-history") {
		log.Printf("Serving traffic history: %s", r.URL.Path)
		getTrafficHistory(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/topology") {
		log.Printf("Serving multi-cluster topology: %s", r.URL.Path)
		getMultiClusterTopology(w, r)
		return
	}

	// Static files fallback
	http.FileServer(http.Dir("ui")).ServeHTTP(w, r)
}

// getTrafficWeights returns current traffic weight configuration from VirtualServices
func getTrafficWeights(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// 실시간으로 VirtualService에서 가중치 조회
	currentWeights := getVirtualServiceWeights()
	log.Printf("Returning traffic weights: %+v", currentWeights)
	
	json.NewEncoder(w).Encode(currentWeights)
}

// getTrafficHistory returns recent traffic routing history
func getTrafficHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trafficHistory)
}

// MultiClusterTopology represents the overall cluster topology
type MultiClusterTopology struct {
	Clusters    []ClusterInfo    `json:"clusters"`
	Services    []ServiceInfo    `json:"services"`
	TrafficFlow []TrafficFlowInfo `json:"trafficFlow"`
	LastUpdated string           `json:"lastUpdated"`
}

// ClusterInfo represents cluster information
type ClusterInfo struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Status   string `json:"status"`
	NodeCount int   `json:"nodeCount"`
}

// ServiceInfo represents service deployment across clusters
type ServiceInfo struct {
	Name        string            `json:"name"`
	Icon        string            `json:"icon"`
	Deployments map[string]string `json:"deployments"` // cluster -> pod name
	Port        string            `json:"port"`
}

// TrafficFlowInfo represents traffic flow between services
type TrafficFlowInfo struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Weight      int    `json:"weight"`
	IsActive    bool   `json:"isActive"`
	FlowType    string `json:"flowType"` // "internal", "cross-cluster"
}

// getDeploymentStatus returns deployment status information for current cluster only
func getDeploymentStatus(w http.ResponseWriter, r *http.Request) {
	if kubernetesClient == nil {
		http.Error(w, "Kubernetes client not available", http.StatusServiceUnavailable)
		return
	}

	// Get pods in theater-msa namespace (current cluster only)
	pods, err := kubernetesClient.CoreV1().Pods("theater-msa").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		http.Error(w, "Failed to get pods: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var deployments []DeploymentInfo
	for _, pod := range pods.Items {
		status := string(pod.Status.Phase)
		if pod.Status.Phase == "Running" {
			// Check if all containers are ready
			allReady := true
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if !containerStatus.Ready {
					allReady = false
					break
				}
			}
			if !allReady {
				status = "Not Ready"
			}
		}

		// Determine service type and cluster
		serviceName := "unknown"
		cluster := "unknown"
		port := "unknown"
		icon := "❓"

		if strings.Contains(pod.Name, "user-service") {
			serviceName = "User Service"
			port = "8081"
			icon = "👤"
			if strings.Contains(pod.Name, "ctx2") {
				cluster = "ctx2"
			} else {
				cluster = "ctx1"
			}
		} else if strings.Contains(pod.Name, "movie-service") {
			serviceName = "Movie Service"
			port = "8082"
			icon = "🎬"
			if strings.Contains(pod.Name, "ctx1") {
				cluster = "ctx1"
			} else {
				cluster = "ctx2"
			}
		} else if strings.Contains(pod.Name, "booking-service") {
			serviceName = "Booking Service"
			port = "8083"
			icon = "🎟️"
			if strings.Contains(pod.Name, "ctx1") {
				cluster = "ctx1"
			} else {
				cluster = "ctx2"
			}
		} else if strings.Contains(pod.Name, "api-gateway") {
			serviceName = "API Gateway"
			port = "8080"
			icon = "🌐"
			cluster = "ctx1"
		} else if strings.Contains(pod.Name, "redis") {
			serviceName = "Redis"
			port = "6379"
			icon = "💾"
			cluster = "ctx1" // Redis is in CTX1
		}

		deployments = append(deployments, DeploymentInfo{
			Service:     serviceName,
			Cluster:     cluster,
			Namespace:   pod.Namespace,
			PodName:     pod.Name,
			NodeName:    pod.Spec.NodeName,
			Status:      status,
			Port:        port,
			Icon:        icon,
			LastChecked: time.Now().Format("2006-01-02 15:04:05"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deployments)
}

// getMultiClusterTopology returns comprehensive multi-cluster topology with traffic flows
func getMultiClusterTopology(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Get current VirtualService weights
	weights := getVirtualServiceWeights()
	
	// Build topology data
	topology := MultiClusterTopology{
		Clusters: []ClusterInfo{
			{
				Name:     "ctx1",
				Provider: "NaverCloud Platform",
				Status:   "Active",
				NodeCount: 1,
			},
			{
				Name:     "ctx2", 
				Provider: "NHN Cloud NKS",
				Status:   "Active",
				NodeCount: 1,
			},
		},
		Services: []ServiceInfo{
			{
				Name: "api-gateway",
				Icon: "🌐",
				Deployments: map[string]string{
					"ctx1": "api-gateway-xxx",
				},
				Port: "8080",
			},
			{
				Name: "user-service",
				Icon: "👤", 
				Deployments: map[string]string{
					"ctx1": "user-service-ctx1-xxx",
					"ctx2": "user-service-ctx2-xxx",
				},
				Port: "8081",
			},
			{
				Name: "movie-service",
				Icon: "🎬",
				Deployments: map[string]string{
					"ctx1": "movie-service-ctx1-xxx",
					"ctx2": "movie-service-ctx2-xxx", 
				},
				Port: "8082",
			},
			{
				Name: "booking-service",
				Icon: "🎟️",
				Deployments: map[string]string{
					"ctx1": "booking-service-ctx1-xxx",
					"ctx2": "booking-service-ctx2-xxx",
				},
				Port: "8083",
			},
			{
				Name: "redis",
				Icon: "💾",
				Deployments: map[string]string{
					"ctx1": "redis-xxx",
				},
				Port: "6379",
			},
		},
		TrafficFlow: []TrafficFlowInfo{
			// External traffic to API Gateway
			{
				From:     "external",
				To:       "api-gateway-ctx1", 
				Weight:   100,
				IsActive: true,
				FlowType: "external",
			},
			// API Gateway to User Service
			{
				From:     "api-gateway-ctx1",
				To:       "user-service-ctx1",
				Weight:   weights.UserServiceCtx1Weight,
				IsActive: weights.UserServiceCtx1Weight > 0,
				FlowType: "internal",
			},
			{
				From:     "api-gateway-ctx1",
				To:       "user-service-ctx2", 
				Weight:   weights.UserServiceCtx2Weight,
				IsActive: weights.UserServiceCtx2Weight > 0,
				FlowType: "cross-cluster",
			},
			// API Gateway to Movie Service
			{
				From:     "api-gateway-ctx1",
				To:       "movie-service-ctx1",
				Weight:   weights.MovieServiceCtx1Weight,
				IsActive: weights.MovieServiceCtx1Weight > 0,
				FlowType: "internal",
			},
			{
				From:     "api-gateway-ctx1", 
				To:       "movie-service-ctx2",
				Weight:   weights.MovieServiceCtx2Weight,
				IsActive: weights.MovieServiceCtx2Weight > 0,
				FlowType: "cross-cluster",
			},
			// API Gateway to Booking Service
			{
				From:     "api-gateway-ctx1",
				To:       "booking-service-ctx1",
				Weight:   weights.BookingServiceCtx1Weight,
				IsActive: weights.BookingServiceCtx1Weight > 0,
				FlowType: "internal",
			},
			{
				From:     "api-gateway-ctx1",
				To:       "booking-service-ctx2",
				Weight:   weights.BookingServiceCtx2Weight, 
				IsActive: weights.BookingServiceCtx2Weight > 0,
				FlowType: "cross-cluster",
			},
		},
		LastUpdated: time.Now().Format("2006-01-02 15:04:05"),
	}
	
	json.NewEncoder(w).Encode(topology)
}

func main() {
	log.Println("Starting API Gateway with weighted traffic distribution...")
	
	http.HandleFunc("/", customHandler)
	
	log.Println("API Gateway is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}