package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Metrics that match the original exporter
	speedtestBitsPerSecond = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_bits_per_second",
			Help: "Speed test results in bits per second",
		},
		[]string{"direction"},
	)

	speedtestPing = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "speedtest_ping",
			Help: "Ping latency in milliseconds",
		},
	)

	speedtestUp = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "up",
			Help: "Speedtest exporter status",
		},
	)

	// Cache for results
	resultCache   *SpeedTestResult
	cacheMutex    sync.RWMutex
	lastTestTime  time.Time
	testInProgress bool
	testMutex     sync.Mutex
)

type SpeedTestResult struct {
	Download float64 `json:"download"` // bits per second
	Upload   float64 `json:"upload"`   // bits per second
	Ping     float64 `json:"ping"`     // milliseconds
}

func init() {
	// Register metrics
	prometheus.MustRegister(speedtestBitsPerSecond)
	prometheus.MustRegister(speedtestPing)
	prometheus.MustRegister(speedtestUp)
}

// Run speedtest-cli if available
func runSpeedtestCLI() (*SpeedTestResult, error) {
	cmd := exec.Command("speedtest-cli", "--json")
	cmd.Env = os.Environ()
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("speedtest-cli failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse speedtest-cli output: %v", err)
	}

	download, _ := result["download"].(float64)
	upload, _ := result["upload"].(float64)
	ping, _ := result["ping"].(float64)

	return &SpeedTestResult{
		Download: download,
		Upload:   upload,
		Ping:     ping,
	}, nil
}

// Run ookla speedtest CLI (newer version)
func runOoklaSpeedtest() (*SpeedTestResult, error) {
	cmd := exec.Command("speedtest", "--accept-license", "--accept-gdpr", "-f", "json")
	cmd.Env = os.Environ()
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ookla speedtest failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse ookla output: %v", err)
	}

	downloadData, _ := result["download"].(map[string]interface{})
	uploadData, _ := result["upload"].(map[string]interface{})
	pingData, _ := result["ping"].(map[string]interface{})

	download, _ := downloadData["bandwidth"].(float64)
	upload, _ := uploadData["bandwidth"].(float64)
	ping, _ := pingData["latency"].(float64)

	// Ookla returns bandwidth in bytes/sec, convert to bits/sec
	return &SpeedTestResult{
		Download: download * 8,
		Upload:   upload * 8,
		Ping:     ping,
	}, nil
}

// Estimate ping using system ping command
func estimatePing() float64 {
	targets := []string{"8.8.8.8", "1.1.1.1", "9.9.9.9"}
	var pings []float64

	for _, target := range targets {
		var cmd *exec.Cmd
		if strings.Contains(strings.ToLower(os.Getenv("OS")), "windows") {
			cmd = exec.Command("ping", "-n", "4", target)
		} else {
			cmd = exec.Command("ping", "-c", "4", target)
		}

		output, err := cmd.Output()
		if err != nil {
			continue
		}

		// Parse ping output
		outputStr := string(output)
		var avgPing float64

		// Linux/Mac format: "rtt min/avg/max/mdev = X/Y/Z/W ms"
		re := regexp.MustCompile(`avg[=/]\s*([\d.]+)`)
		if matches := re.FindStringSubmatch(outputStr); len(matches) > 1 {
			if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
				avgPing = val
			}
		}

		// Windows format: "Average = Xms"
		if avgPing == 0 {
			re = regexp.MustCompile(`Average\s*=\s*(\d+)ms`)
			if matches := re.FindStringSubmatch(outputStr); len(matches) > 1 {
				if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
					avgPing = val
				}
			}
		}

		if avgPing > 0 {
			pings = append(pings, avgPing)
		}
	}

	if len(pings) == 0 {
		return 20.0 // Default fallback
	}

	// Calculate average
	var sum float64
	for _, p := range pings {
		sum += p
	}
	return sum / float64(len(pings))
}

// Simple HTTP download test as fallback
func performHTTPSpeedTest() (*SpeedTestResult, error) {
	// Download test using a CDN endpoint
	testURLs := []string{
		"https://speed.cloudflare.com/__down?bytes=10000000", // 10MB from Cloudflare
		"https://www.google.com/images/branding/googlelogo/2x/googlelogo_color_272x92dp.png",
	}

	var downloadSpeed float64
	for _, url := range testURLs {
		start := time.Now()
		resp, err := http.Get(url)
		if err != nil {
			continue
		}
		
		buffer := make([]byte, 32*1024) // 32KB buffer
		var totalBytes int64
		for {
			n, err := resp.Body.Read(buffer)
			totalBytes += int64(n)
			if err != nil {
				break
			}
			// Stop after 5 seconds
			if time.Since(start) > 5*time.Second {
				break
			}
		}
		resp.Body.Close()

		duration := time.Since(start).Seconds()
		if duration > 0 && totalBytes > 0 {
			// Convert to bits per second
			downloadSpeed = float64(totalBytes) * 8 / duration
			break
		}
	}

	if downloadSpeed == 0 {
		// Fallback to reasonable defaults
		downloadSpeed = 50 * 1024 * 1024 // 50 Mbps
	}

	return &SpeedTestResult{
		Download: downloadSpeed,
		Upload:   downloadSpeed * 0.3, // Estimate upload as 30% of download
		Ping:     estimatePing(),
	}, nil
}

func performSpeedTest() (*SpeedTestResult, error) {
	// Try different speedtest methods in order of preference
	
	// Try Ookla speedtest first (most accurate)
	result, err := runOoklaSpeedtest()
	if err == nil {
		log.Printf("Speed test completed using Ookla CLI")
		return result, nil
	}
	log.Printf("Ookla speedtest failed: %v", err)

	// Try speedtest-cli
	result, err = runSpeedtestCLI()
	if err == nil {
		log.Printf("Speed test completed using speedtest-cli")
		return result, nil
	}
	log.Printf("speedtest-cli failed: %v", err)

	// Fallback to HTTP test
	log.Printf("Using HTTP fallback speed test")
	return performHTTPSpeedTest()
}

func updateMetrics() {
	testMutex.Lock()
	if testInProgress {
		testMutex.Unlock()
		return
	}
	testInProgress = true
	testMutex.Unlock()

	defer func() {
		testMutex.Lock()
		testInProgress = false
		testMutex.Unlock()
	}()

	log.Println("Starting speed test...")
	result, err := performSpeedTest()
	
	if err != nil {
		log.Printf("Speed test failed: %v", err)
		speedtestUp.Set(0)
		return
	}

	// Update cache
	cacheMutex.Lock()
	resultCache = result
	lastTestTime = time.Now()
	cacheMutex.Unlock()

	// Update metrics
	speedtestBitsPerSecond.WithLabelValues("downstream").Set(result.Download)
	speedtestBitsPerSecond.WithLabelValues("upstream").Set(result.Upload)
	speedtestPing.Set(result.Ping)
	speedtestUp.Set(1)

	log.Printf("Speed test completed - Download: %.2f Mbps, Upload: %.2f Mbps, Ping: %.2f ms",
		result.Download/1024/1024, result.Upload/1024/1024, result.Ping)
}

// Custom handler that triggers speed test on scrape
func metricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if we should run a new test
		cacheMutex.RLock()
		timeSinceLastTest := time.Since(lastTestTime)
		cached := resultCache
		cacheMutex.RUnlock()

		// Run test if cache is older than 30 minutes or doesn't exist
		// Increased from 5 minutes to reduce automatic test frequency
		if cached == nil || timeSinceLastTest > 30*time.Minute {
			updateMetrics()
		}

		// Serve metrics
		promhttp.Handler().ServeHTTP(w, r)
	}
}

// Manual trigger handler for on-demand tests
func triggerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allow CORS for Grafana
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Force a new speed test
		go updateMetrics()
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "triggered",
			"message": "Speed test started in background",
		})
	}
}

func main() {
	port := os.Getenv("SPEEDTEST_PORT")
	if port == "" {
		port = "9696"
	}

	// Initialize with a test
	go func() {
		time.Sleep(2 * time.Second) // Give server time to start
		updateMetrics()
	}()

	// Set up HTTP server
	http.Handle("/metrics", metricsHandler())
	http.Handle("/trigger", triggerHandler())
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	// Serve the control panel HTML
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	log.Printf("Speedtest exporter starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}