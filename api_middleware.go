package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ============================================================================
// CORS MIDDLEWARE
// ============================================================================

// corsMiddleware handles Cross-Origin Resource Sharing (CORS) based on server.yml configuration
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !Globals.CORSEnabled {
			next.ServeHTTP(w, r)
			return
		}

		// Log the request
		var scheme string = "R"
		if Globals.EnableHTTPS {
			scheme = "HTTP r"
			if r.TLS != nil {
				scheme = "HTTPS r"
			}
		}
		mlog(5, "§bcorsMiddleware(): §f%sequest from §9%s§f to §9%s§f with method §9%s", scheme, r.RemoteAddr, r.URL.Path, r.Method)

		// Set CORS headers based on configuration
		origin := r.Header.Get("Origin")

		// Handle allowed origins
		if len(Globals.CORSAllowedOrigins) > 0 {
			if contains(Globals.CORSAllowedOrigins, "*") {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if contains(Globals.CORSAllowedOrigins, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			}
		}

		// Set allowed methods
		if len(Globals.CORSAllowedMethods) > 0 {
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(Globals.CORSAllowedMethods, ", "))
		}

		// Set allowed headers
		if len(Globals.CORSAllowedHeaders) > 0 {
			if contains(Globals.CORSAllowedHeaders, "*") {
				w.Header().Set("Access-Control-Allow-Headers", "*")
			} else {
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(Globals.CORSAllowedHeaders, ", "))
			}
		}

		// Set exposed headers
		if len(Globals.CORSExposedHeaders) > 0 {
			w.Header().Set("Access-Control-Expose-Headers", strings.Join(Globals.CORSExposedHeaders, ", "))
		}

		// Set credentials flag
		if Globals.CORSAllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Set max age for preflight caching
		if Globals.CORSMaxAge > 0 {
			w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", Globals.CORSMaxAge))
		}

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Set Content-Type for non-OPTIONS requests
		if r.Method != "OPTIONS" {
			w.Header().Set("Content-Type", "application/json")
		}

		next.ServeHTTP(w, r)
	})
}

// contains checks if a string slice contains a specific string
func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

// ============================================================================
// MAX REQUEST SIZE MIDDLEWARE
// ============================================================================

// maxRequestSizeMiddleware limits the maximum size of request bodies based on server.yml
func maxRequestSizeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use configurable max request size from server.yml (in KB)
		maxSizeBytes := int64(Globals.MaxRequestSizeKB * 1024)
		r.Body = http.MaxBytesReader(w, r.Body, maxSizeBytes)

		if err := r.ParseForm(); err != nil {
			mlog(3, "§bmaxRequestSizeMiddleware(): §eRequest too large from §9%s§e (max: §9%d KB§e)", r.RemoteAddr, Globals.MaxRequestSizeKB)
			http.Error(w, "Request too large", http.StatusRequestEntityTooLarge)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ============================================================================
// RATE LIMITER DATA STRUCTURES
// ============================================================================

// RateLimiterState tracks rate limit state for the entire system
type RateLimiterState struct {
	mu sync.RWMutex

	// Per-group totals: map[groupName][timeScale]counter
	groupCounters map[string]map[string]*RateLimitCounter

	// Per-IP within group: map[groupName][ip][timeScale]counter
	ipCounters map[string]map[string]map[string]*RateLimitCounter
}

// RateLimitCounter tracks requests in a time window
type RateLimitCounter struct {
	count       int
	windowStart time.Time
	windowSize  time.Duration
}

// Global rate limiter state
var rateLimiter = &RateLimiterState{
	groupCounters: make(map[string]map[string]*RateLimitCounter),
	ipCounters:    make(map[string]map[string]map[string]*RateLimitCounter),
}

// ============================================================================
// RATE LIMITER CORE LOGIC
// ============================================================================

// GetGroupFromToken extracts the group name from a bearer token
func GetGroupFromToken(authHeader string) string {
	if authHeader == "" {
		return "default"
	}

	// Extract token from "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "default"
	}

	token := strings.TrimSpace(parts[1])

	// Find which group this token belongs to
	for groupName, tokens := range Globals.GroupTokens {
		for _, groupToken := range tokens {
			if token == groupToken {
				return groupName
			}
		}
	}

	return "default"
}

// CheckRateLimit checks if a request should be allowed based on rate limits
// Returns: allowed (bool), remainingRequests (int), resetTime (time.Time), limitType (string)
func CheckRateLimit(group, ip string) (bool, int, time.Time, string) {
	if !Globals.EnableRateLimit {
		return true, -1, time.Time{}, ""
	}

	// Get group configuration
	groupConfig, exists := Globals.RateLimitGroups[group]
	if !exists {
		// No limit configured for this group
		return true, -1, time.Time{}, ""
	}

	rateLimiter.mu.Lock()
	defer rateLimiter.mu.Unlock()

	now := time.Now()

	// Check per-second limits first (smallest time scale)
	if groupConfig.PerSecond != nil {
		allowed, remaining, reset, limitType := checkTimeScale(
			group, ip, "second", time.Second,
			groupConfig.PerSecond.GroupTotal,
			groupConfig.PerSecond.PerIP,
			now,
		)
		if !allowed {
			return false, remaining, reset, limitType
		}
	}

	// Check per-minute limits
	if groupConfig.PerMinute != nil {
		allowed, remaining, reset, limitType := checkTimeScale(
			group, ip, "minute", time.Minute,
			groupConfig.PerMinute.GroupTotal,
			groupConfig.PerMinute.PerIP,
			now,
		)
		if !allowed {
			return false, remaining, reset, limitType
		}
	}

	// Check per-hour limits
	if groupConfig.PerHour != nil {
		allowed, remaining, reset, limitType := checkTimeScale(
			group, ip, "hour", time.Hour,
			groupConfig.PerHour.GroupTotal,
			groupConfig.PerHour.PerIP,
			now,
		)
		if !allowed {
			return false, remaining, reset, limitType
		}
	}

	// All checks passed - increment counters
	incrementCounters(group, ip, groupConfig, now)

	// Calculate remaining requests (use the most restrictive limit)
	remaining := calculateRemainingRequests(group, ip, groupConfig, now)

	return true, remaining, time.Time{}, ""
}

// checkTimeScale checks rate limit for a specific time scale
func checkTimeScale(group, ip, scale string, duration time.Duration, groupLimit, ipLimit int, now time.Time) (bool, int, time.Time, string) {
	// Check group total limit
	if groupLimit > 0 {
		counter := getOrCreateCounter(rateLimiter.groupCounters, group, scale, duration, now)
		if counter.count >= groupLimit {
			resetTime := counter.windowStart.Add(duration)
			remaining := 0
			return false, remaining, resetTime, fmt.Sprintf("group_%s", scale)
		}
	}

	// Check per-IP limit
	if ipLimit > 0 {
		counter := getOrCreateIPCounter(rateLimiter.ipCounters, group, ip, scale, duration, now)
		if counter.count >= ipLimit {
			resetTime := counter.windowStart.Add(duration)
			remaining := 0
			return false, remaining, resetTime, fmt.Sprintf("ip_%s", scale)
		}
	}

	return true, -1, time.Time{}, ""
}

// getOrCreateCounter gets or creates a counter for group totals
func getOrCreateCounter(counters map[string]map[string]*RateLimitCounter, group, scale string, duration time.Duration, now time.Time) *RateLimitCounter {
	if counters[group] == nil {
		counters[group] = make(map[string]*RateLimitCounter)
	}

	counter := counters[group][scale]
	if counter == nil || now.Sub(counter.windowStart) >= duration {
		// Create new counter or reset expired counter
		counter = &RateLimitCounter{
			count:       0,
			windowStart: now,
			windowSize:  duration,
		}
		counters[group][scale] = counter
	}

	return counter
}

// getOrCreateIPCounter gets or creates a counter for per-IP limits
func getOrCreateIPCounter(counters map[string]map[string]map[string]*RateLimitCounter, group, ip, scale string, duration time.Duration, now time.Time) *RateLimitCounter {
	if counters[group] == nil {
		counters[group] = make(map[string]map[string]*RateLimitCounter)
	}
	if counters[group][ip] == nil {
		counters[group][ip] = make(map[string]*RateLimitCounter)
	}

	counter := counters[group][ip][scale]
	if counter == nil || now.Sub(counter.windowStart) >= duration {
		counter = &RateLimitCounter{
			count:       0,
			windowStart: now,
			windowSize:  duration,
		}
		counters[group][ip][scale] = counter
	}

	return counter
}

// incrementCounters increments all relevant counters for a successful request
func incrementCounters(group, ip string, config RateLimitGroup, now time.Time) {
	if config.PerSecond != nil {
		if config.PerSecond.GroupTotal > 0 {
			counter := getOrCreateCounter(rateLimiter.groupCounters, group, "second", time.Second, now)
			counter.count++
		}
		if config.PerSecond.PerIP > 0 {
			counter := getOrCreateIPCounter(rateLimiter.ipCounters, group, ip, "second", time.Second, now)
			counter.count++
		}
	}

	if config.PerMinute != nil {
		if config.PerMinute.GroupTotal > 0 {
			counter := getOrCreateCounter(rateLimiter.groupCounters, group, "minute", time.Minute, now)
			counter.count++
		}
		if config.PerMinute.PerIP > 0 {
			counter := getOrCreateIPCounter(rateLimiter.ipCounters, group, ip, "minute", time.Minute, now)
			counter.count++
		}
	}

	if config.PerHour != nil {
		if config.PerHour.GroupTotal > 0 {
			counter := getOrCreateCounter(rateLimiter.groupCounters, group, "hour", time.Hour, now)
			counter.count++
		}
		if config.PerHour.PerIP > 0 {
			counter := getOrCreateIPCounter(rateLimiter.ipCounters, group, ip, "hour", time.Hour, now)
			counter.count++
		}
	}
}

// calculateRemainingRequests calculates the minimum remaining requests across all limits
func calculateRemainingRequests(group, ip string, config RateLimitGroup, now time.Time) int {
	remaining := -1 // -1 means unlimited

	if config.PerSecond != nil {
		if config.PerSecond.GroupTotal > 0 {
			counter := getOrCreateCounter(rateLimiter.groupCounters, group, "second", time.Second, now)
			r := config.PerSecond.GroupTotal - counter.count
			if remaining == -1 || r < remaining {
				remaining = r
			}
		}
		if config.PerSecond.PerIP > 0 {
			counter := getOrCreateIPCounter(rateLimiter.ipCounters, group, ip, "second", time.Second, now)
			r := config.PerSecond.PerIP - counter.count
			if remaining == -1 || r < remaining {
				remaining = r
			}
		}
	}

	if config.PerMinute != nil {
		if config.PerMinute.GroupTotal > 0 {
			counter := getOrCreateCounter(rateLimiter.groupCounters, group, "minute", time.Minute, now)
			r := config.PerMinute.GroupTotal - counter.count
			if remaining == -1 || r < remaining {
				remaining = r
			}
		}
		if config.PerMinute.PerIP > 0 {
			counter := getOrCreateIPCounter(rateLimiter.ipCounters, group, ip, "minute", time.Minute, now)
			r := config.PerMinute.PerIP - counter.count
			if remaining == -1 || r < remaining {
				remaining = r
			}
		}
	}

	if config.PerHour != nil {
		if config.PerHour.GroupTotal > 0 {
			counter := getOrCreateCounter(rateLimiter.groupCounters, group, "hour", time.Hour, now)
			r := config.PerHour.GroupTotal - counter.count
			if remaining == -1 || r < remaining {
				remaining = r
			}
		}
		if config.PerHour.PerIP > 0 {
			counter := getOrCreateIPCounter(rateLimiter.ipCounters, group, ip, "hour", time.Hour, now)
			r := config.PerHour.PerIP - counter.count
			if remaining == -1 || r < remaining {
				remaining = r
			}
		}
	}

	return remaining
}

// ============================================================================
// RATE LIMITER MIDDLEWARE
// ============================================================================

// rateLimitMiddleware applies rate limiting to requests
func rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !Globals.EnableRateLimit {
			next.ServeHTTP(w, r)
			return
		}

		// Extract client IP (handle X-Forwarded-For, X-Real-IP)
		ip := getClientIP(r)

		// Get group from authorization header
		authHeader := r.Header.Get("Authorization")
		group := GetGroupFromToken(authHeader)

		// Check rate limit
		allowed, remaining, resetTime, limitType := CheckRateLimit(group, ip)

		// Add rate limit headers if enabled
		if Globals.RateLimitHeaders {
			if remaining >= 0 {
				w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			}
			if !resetTime.IsZero() {
				w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))
			}
			w.Header().Set("X-RateLimit-Group", group)
		}

		if !allowed {
			// Rate limit exceeded
			mlog(3, "§brate_limiter: §eRate limit exceeded for §9%s§e (group: §9%s§e, limit: §9%s§e)", ip, group, limitType)

			if !resetTime.IsZero() {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(time.Until(resetTime).Seconds())))
			}

			http.Error(w, "Rate limit exceeded", Globals.RateLimitResponseCode)
			return
		}

		// Request allowed
		next.ServeHTTP(w, r)
	})
}

// getClientIP extracts the real client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (used by proxies)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fallback to RemoteAddr
	ip := r.RemoteAddr
	// Remove port if present
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return ip
}
