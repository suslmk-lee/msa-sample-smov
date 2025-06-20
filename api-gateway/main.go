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
		var selectedService string
		forceCluster := r.Header.Get("X-Force-Cluster")
		
		if forceCluster == "ctx1" {
			selectedService = "http://user-service:8081"
			addToHistory("user", "ctx1")
			log.Printf("Forced routing to user-service CTX1: %s", selectedService)
		} else if forceCluster == "ctx2" {
			selectedService = "http://user-service-ctx2:8081"
			addToHistory("user", "ctx2")
			log.Printf("Forced routing to user-service CTX2: %s", selectedService)
		} else {
			selectedService = weightedServiceSelect(
				"user",
				trafficWeights.UserServiceCtx1Weight,
				trafficWeights.UserServiceCtx2Weight,
				"http://user-service:8081",
				"http://user-service-ctx2:8081",
			)
			log.Printf("Weighted routing to user-service: %s", selectedService)
		}
		
		proxy := newReverseProxy(selectedService)
		proxy.ServeHTTP(w, r)
		return
	}
	
	if strings.HasPrefix(r.URL.Path, "/movies/") {
		selectedService := weightedServiceSelect(
			"movie",
			trafficWeights.MovieServiceCtx1Weight,
			trafficWeights.MovieServiceCtx2Weight,
			"http://movie-service-ctx1:8082",
			"http://movie-service:8082",
		)
		log.Printf("Routing to movie-service: %s", selectedService)
		proxy := newReverseProxy(selectedService)
		proxy.ServeHTTP(w, r)
		return
	}
	
	if strings.HasPrefix(r.URL.Path, "/bookings/") {
		selectedService := weightedServiceSelect(
			"booking",
			trafficWeights.BookingServiceCtx1Weight,
			trafficWeights.BookingServiceCtx2Weight,
			"http://booking-service-ctx1:8083",
			"http://booking-service:8083",
		)
		log.Printf("Routing to booking-service: %s", selectedService)
		proxy := newReverseProxy(selectedService)
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

	// Static files fallback
	http.FileServer(http.Dir("ui")).ServeHTTP(w, r)
}

// getTrafficWeights returns current traffic weight configuration
func getTrafficWeights(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trafficWeights)
}

// getTrafficHistory returns recent traffic routing history
func getTrafficHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trafficHistory)
}

// getDeploymentStatus returns deployment status information for Kubernetes
func getDeploymentStatus(w http.ResponseWriter, r *http.Request) {
	if kubernetesClient == nil {
		http.Error(w, "Kubernetes client not available", http.StatusServiceUnavailable)
		return
	}

	// Get pods in theater-msa namespace
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
			// Determine cluster based on node or other criteria
			cluster = "ctx1" // Default to ctx1, adjust logic as needed
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

func main() {
	log.Println("Starting API Gateway with weighted traffic distribution...")
	
	http.HandleFunc("/", customHandler)
	
	log.Println("API Gateway is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}