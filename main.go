package main

import (
	"crypto/md5"
	"encoding/hex"
	"log"

	"encoding/json"

	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/patrickmn/go-cache"

	"webear/executor"
)

type ListenerConfig struct {
	Script string `toml:"script"`
	Secret string `toml:"secret"`
	User   string `toml:"user"`
}

type Config struct {
	Listener map[string]ListenerConfig `toml:"listener"`
}

type WebhookPayload struct {
	Secret        string `json:"secret"`
	IdempotentKey string `json:"idempotent_key"`
	Data          string `json:"data"`
}

var config Config
var memoryCache *cache.Cache

func init() {
	memoryCache = cache.New(time.Hour, time.Hour * 2)
	configPath := "/etc/webear/config.toml"
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		log.Fatalf("Error loading the config file: %v", err)
	}
}

func calculateMD5(input string) string {
	hash := md5.New()
	hash.Write([]byte(input))
	return hex.EncodeToString(hash.Sum(nil))
}

func requestValidator(name string, payload *WebhookPayload, w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return false
	}

	if name == "" {
		http.Error(w, "Missing listener name", http.StatusBadRequest)
		return false
	}

	if err := json.NewDecoder(r.Body).Decode(payload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return false
	}

	if payload.Secret == "" {
		http.Error(w, "Missing secret", http.StatusBadRequest)
		return false
	}

	if payload.IdempotentKey == "" {
		http.Error(w, "Missing idempotent key", http.StatusBadRequest)
		return false
	}

	return true
}

func requestAuthenticator(name string, payload *WebhookPayload, w http.ResponseWriter) bool {
	listenerConfig, listener_exists := config.Listener[name]
	if !listener_exists {
		http.Error(w, "Listener not found", http.StatusNotFound)
		return false
	}

	if payload.Secret != listenerConfig.Secret {
		http.Error(w, "Invalid secret", http.StatusUnauthorized)
		return false
	}

	idempotency_hash := calculateMD5(payload.IdempotentKey)
	if _, found := memoryCache.Get(idempotency_hash); found {
		http.Error(w, "Duplicate request", http.StatusConflict)
		return false
	}

	memoryCache.Set(idempotency_hash, true, cache.DefaultExpiration)

	return true
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[1:] // remove the leading slash
	var payload WebhookPayload

	// call to requestValidator also populates the payload
	if !requestValidator(name, &payload, w, r) {
		return
	}

	if !requestAuthenticator(name, &payload, w) {
		return
	}

	listenerConfig := config.Listener[name]

	scriptPath := listenerConfig.Script
	if !filepath.IsAbs(scriptPath) {
		http.Error(w, "Could not execute the script", http.StatusInternalServerError)
		log.Println("Script path is not absolute [", scriptPath, "]")
		return
	}

	if info, err := os.Stat(scriptPath); err != nil || info.Mode()&0111 == 0 {
		http.Error(w, "Could not execute the script", http.StatusInternalServerError)
		log.Println("Script is not executable [", scriptPath, "]")
		return
	}

	err := executor.ExecuteScript(payload.Data, name, scriptPath, listenerConfig.User)
	if err != nil {
		http.Error(w, "Could not execute the script", http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Printf("Webhook [%s] executed successfully", name)
}

func main() {
	http.HandleFunc("/", webhookHandler)

	log.Println("Starting the server on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error starting the server: %v", err)
	}
}