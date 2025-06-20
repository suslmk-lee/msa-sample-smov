package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
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

	log.Printf("Kubernetes client initialized - Istio will handle traffic distribution")
}

// newReverseProxy creates a new reverse proxy for the target URL.
func newReverseProxy(target string) *httputil.ReverseProxy {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Failed to parse target URL %s: %v", target, err)
	}
	return httputil.NewSingleHostReverseProxy(targetURL)
}

// customHandler handles routing between API calls and static files - Istio handles load balancing
func customHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request: %s %s", r.Method, r.URL.Path)

	// Simple API routes - Istio VirtualService handles traffic distribution
	if strings.HasPrefix(r.URL.Path, "/users/") {
		log.Printf("Routing to user-service (Istio will handle load balancing)")
		proxy := newReverseProxy("http://user-service:8081")
		proxy.ServeHTTP(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/movies/") {
		log.Printf("Routing to movie-service (Istio will handle load balancing)")
		proxy := newReverseProxy("http://movie-service:8082")
		proxy.ServeHTTP(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/bookings/") {
		log.Printf("Routing to booking-service (Istio will handle load balancing)")
		proxy := newReverseProxy("http://booking-service:8083")
		proxy.ServeHTTP(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/deployment-status") {
		log.Printf("Serving deployment status: %s", r.URL.Path)
		getDeploymentStatus(w, r)
		return
	}

	// Static files fallback
	http.FileServer(http.Dir("ui")).ServeHTTP(w, r)
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
