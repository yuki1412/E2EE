//This will be the main file for the backend

package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/crypto/hkdf"
)

// Request/Response structures matching the frontend
type KeyExchangeRequest struct {
	MyUserId    string `json:"myUserId"`
	OtherUserId string `json:"otherUserId"`
	MyPublicKey string `json:"myPublicKey"`
}

type KeyExchangeResponse struct {
	TheirPublicKey string `json:"theirPublicKey"`
	SessionId      string `json:"sessionId"`
	Success        bool   `json:"success"`
}

type SendMessageRequest struct {
	SessionId     string `json:"sessionId"`
	EncryptedData string `json:"encryptedData"`
	IV            string `json:"iv"`
	Timestamp     int64  `json:"timestamp"`
}

type Message struct {
	ID               string `json:"id"`
	Sender           string `json:"sender"`
	EncryptedContent string `json:"encryptedContent"`
	IV               string `json:"iv"`
	Timestamp        int64  `json:"timestamp"`
}

type GetMessagesResponse struct {
	Messages []Message `json:"messages"`
}

// Session management
type Session struct {
	ID        string
	CreatedAt time.Time
	Messages  []Message
}

// Session management for peer-to-peer key exchange
type PeerSession struct {
	ID         string
	AliceKey   string
	BobKey     string
	AliceReady bool
	BobReady   bool
	CreatedAt  time.Time
	Messages   []Message
}

// Global state (in production, use Redis or database)
var (
	sessions     = make(map[string]*Session)
	peerSessions = make(map[string]*PeerSession)
	mu           sync.RWMutex
)

// ECDH Service for Go
type ECDHService struct {
	privateKey *ecdh.PrivateKey
	publicKey  *ecdh.PublicKey
}

func NewECDHService() (*ECDHService, error) {
	privateKey, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	return &ECDHService{
		privateKey: privateKey,
		publicKey:  privateKey.PublicKey(),
	}, nil
}

func (e *ECDHService) ExportPublicKey() string {
	publicKeyBytes := e.publicKey.Bytes()
	return base64.StdEncoding.EncodeToString(publicKeyBytes)
}

func (e *ECDHService) ImportPublicKey(publicKeyBase64 string) (*ecdh.PublicKey, error) {
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return nil, err
	}

	return ecdh.P256().NewPublicKey(publicKeyBytes)
}

func (e *ECDHService) DeriveSharedSecret(theirPublicKey *ecdh.PublicKey) ([]byte, error) {
	return e.privateKey.ECDH(theirPublicKey)
}

func (e *ECDHService) DeriveAESKey(sharedSecret []byte) ([]byte, error) {
	// Use HKDF to derive AES key - matching the frontend implementation
	hkdf := hkdf.New(sha256.New, sharedSecret, nil, []byte("ECDH-AES-GCM"))
	aesKey := make([]byte, 32) // 256-bit key
	_, err := io.ReadFull(hkdf, aesKey)
	return aesKey, err
}

func (e *ECDHService) EncryptMessage(message string, aesKey []byte) (string, string, error) {
	// Generate random IV
	iv := make([]byte, 12) // 96-bit IV for GCM
	if _, err := rand.Read(iv); err != nil {
		return "", "", err
	}

	// Create AES-GCM cipher
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", err
	}

	// Encrypt the message
	ciphertext := gcm.Seal(nil, iv, []byte(message), nil)

	// Return base64 encoded ciphertext and IV
	return base64.StdEncoding.EncodeToString(ciphertext),
		base64.StdEncoding.EncodeToString(iv), nil
}

func (e *ECDHService) DecryptMessage(encryptedData, ivBase64 string, aesKey []byte) (string, error) {
	// Decode base64 data
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", err
	}

	iv, err := base64.StdEncoding.DecodeString(ivBase64)
	if err != nil {
		return "", err
	}

	// Create AES-GCM cipher
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Decrypt the message
	plaintext, err := gcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// HTTP Handlers
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func handleKeyExchange(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req KeyExchangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("🔄 Key exchange request from user: %s to connect with: %s", req.MyUserId, req.OtherUserId)

	// Create or find peer session based on sorted user IDs for consistency
	var sessionId string
	if req.MyUserId < req.OtherUserId {
		sessionId = fmt.Sprintf("peer_%s_%s", req.MyUserId, req.OtherUserId)
	} else {
		sessionId = fmt.Sprintf("peer_%s_%s", req.OtherUserId, req.MyUserId)
	}

	mu.Lock()
	peerSession, exists := peerSessions[sessionId]
	if !exists {
		// Create new peer session
		peerSession = &PeerSession{
			ID:        sessionId,
			CreatedAt: time.Now(),
			Messages:  make([]Message, 0),
		}
		peerSessions[sessionId] = peerSession
	}

	var otherUserKey string
	var isReady bool

	// Store this user's key and check if the other user is ready
	if req.MyUserId == "alice" {
		if peerSession.AliceReady && peerSession.AliceKey != "" {
			if peerSession.AliceKey == req.MyPublicKey {
				log.Printf("⚠️  Alice reconnecting with SAME key: %s...", peerSession.AliceKey[:16])
			} else {
				log.Printf("🔄 Alice connecting with DIFFERENT key - clearing Bob's key for fresh exchange")
				log.Printf("   Old Alice key: %s...", peerSession.AliceKey[:16])
				log.Printf("   New Alice key: %s...", req.MyPublicKey[:16])
				// Clear Bob's key since Alice has a new key pair
				peerSession.BobKey = ""
				peerSession.BobReady = false
				peerSession.AliceKey = req.MyPublicKey
				peerSession.AliceReady = true // Make sure Alice is marked as ready
			}
		} else {
			peerSession.AliceKey = req.MyPublicKey
			peerSession.AliceReady = true
			log.Printf("✅ Alice connected with NEW key: %s...", req.MyPublicKey[:16])
		}
		otherUserKey = peerSession.BobKey
		isReady = peerSession.BobReady && peerSession.BobKey != ""
		log.Printf("🔍 Alice checking Bob: BobReady=%v, BobKey='%s'", peerSession.BobReady, peerSession.BobKey)
	} else if req.MyUserId == "bob" {
		if peerSession.BobReady && peerSession.BobKey != "" {
			if peerSession.BobKey == req.MyPublicKey {
				log.Printf("⚠️  Bob reconnecting with SAME key: %s...", peerSession.BobKey[:16])
			} else {
				log.Printf("🔄 Bob connecting with DIFFERENT key - clearing Alice's key for fresh exchange")
				log.Printf("   Old Bob key: %s...", peerSession.BobKey[:16])
				log.Printf("   New Bob key: %s...", req.MyPublicKey[:16])
				// Clear Alice's key since Bob has a new key pair
				peerSession.AliceKey = ""
				peerSession.AliceReady = false
				peerSession.BobKey = req.MyPublicKey
				peerSession.BobReady = true // Make sure Bob is marked as ready
			}
		} else {
			peerSession.BobKey = req.MyPublicKey
			peerSession.BobReady = true
			log.Printf("✅ Bob connected with NEW key: %s...", req.MyPublicKey[:16])
		}
		otherUserKey = peerSession.AliceKey
		isReady = peerSession.AliceReady && peerSession.AliceKey != ""
		log.Printf("🔍 Bob checking Alice: AliceReady=%v, AliceKey='%s'", peerSession.AliceReady, peerSession.AliceKey)
	} else {
		mu.Unlock()
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	log.Printf("📊 Session status - Alice ready: %v (key: %s...), Bob ready: %v (key: %s...)",
		peerSession.AliceReady,
		func() string {
			if peerSession.AliceKey != "" {
				return peerSession.AliceKey[:16]
			} else {
				return "none"
			}
		}(),
		peerSession.BobReady,
		func() string {
			if peerSession.BobKey != "" {
				return peerSession.BobKey[:16]
			} else {
				return "none"
			}
		}())

	log.Printf("🔍 isReady check: isReady=%v, otherUserKey='%s'", isReady, otherUserKey)

	mu.Unlock()

	if !isReady || otherUserKey == "" {
		// Other user hasn't connected yet, wait and respond with empty key
		log.Printf("⏳ Waiting for the other user to connect...")
		response := KeyExchangeResponse{
			TheirPublicKey: "",
			SessionId:      sessionId,
			Success:        false,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Both users are ready - provide the other user's public key
	log.Printf("🎉 Both users connected! Peer-to-peer key exchange complete")
	log.Printf("🔑 Alice will use Bob's key, Bob will use Alice's key")
	log.Printf("💡 Both will derive the SAME shared secret for E2EE")

	response := KeyExchangeResponse{
		TheirPublicKey: otherUserKey,
		SessionId:      sessionId,
		Success:        true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleSendMessage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received request to send message")
	enableCORS(w)

	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("📥 Received encrypted message for session: %s", req.SessionId)

	// Check if it's a peer session first
	mu.RLock()
	peerSession, isPeer := peerSessions[req.SessionId]
	session, isRegular := sessions[req.SessionId]
	mu.RUnlock()

	if !isPeer && !isRegular {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Store the encrypted message without decrypting it
	message := Message{
		ID:               fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		Sender:           "client",
		EncryptedContent: req.EncryptedData,
		IV:               req.IV,
		Timestamp:        req.Timestamp,
	}

	mu.Lock()
	if isPeer {
		peerSession.Messages = append(peerSession.Messages, message)
	} else {
		session.Messages = append(session.Messages, message)
	}
	mu.Unlock()

	log.Printf("📦 Stored encrypted message (ID: %s) - Backend never decrypts!", message.ID)
	log.Printf("🔒 Message remains end-to-end encrypted")

	// Send back a simple acknowledgment (no echo functionality)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "success",
		"messageId": message.ID,
		"info":      "Message stored encrypted - backend never decrypts",
	})
}

func handleGetMessages(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionId := r.URL.Query().Get("sessionId")
	if sessionId == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	mu.RLock()
	peerSession, isPeer := peerSessions[sessionId]
	session, isRegular := sessions[sessionId]
	mu.RUnlock()

	if !isPeer && !isRegular {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	var messages []Message
	if isPeer {
		messages = peerSession.Messages
	} else {
		messages = session.Messages
	}

	response := GetMessagesResponse{
		Messages: messages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "healthy",
		"timestamp":    time.Now().Unix(),
		"sessions":     len(sessions),
		"peerSessions": len(peerSessions),
		"crypto":       "ECDH P-256 + AES-GCM-256",
		"mode":         "Peer-to-peer key exchange",
	})
}

func handleReset(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	mu.Lock()
	sessions = make(map[string]*Session)
	peerSessions = make(map[string]*PeerSession)
	mu.Unlock()

	log.Printf("🔄 All sessions reset!")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "reset complete",
		"info":   "All sessions cleared",
	})
}

func main() {
	log.Println("🚀 Starting ECDH Backend Server...")
	log.Println("🔐 Using ECDH P-256 + AES-GCM-256")
	log.Println("🤝 Peer-to-peer key exchange mode")
	log.Println("🌐 CORS enabled for http://localhost:3000")

	// Setup routes
	http.HandleFunc("/api/key-exchange", handleKeyExchange)
	http.HandleFunc("/api/messages", handleSendMessage)
	http.HandleFunc("/api/messages/get", handleGetMessages)
	http.HandleFunc("/api/health", handleHealth)
	http.HandleFunc("/api/reset", handleReset)

	// Serve static info page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
		<html>
		<head><title>ECDH Backend Server</title></head>
		<body style="font-family: Arial; padding: 20px;">
			<h1>🔐 ECDH P-256 Backend Server</h1>
			<p><strong>Status:</strong> ✅ Running</p>
			<p><strong>Mode:</strong> 🤝 Peer-to-peer key exchange</p>
			<p><strong>Crypto:</strong> ECDH P-256 + AES-GCM-256</p>
			<p><strong>Frontend:</strong> <a href="http://localhost:3000">http://localhost:3000</a></p>
			<h2>API Endpoints:</h2>
			<ul>
				<li><code>POST /api/key-exchange</code> - ECDH key exchange (peer-to-peer)</li>
				<li><code>POST /api/messages</code> - Send encrypted message</li>
				<li><code>GET /api/messages/get?sessionId=X</code> - Get messages</li>
				<li><code>GET /api/health</code> - Health check</li>
				<li><code>POST /api/reset</code> - Reset sessions</li>
			</ul>
			<p><em>Regular Sessions: %d | Peer Sessions: %d</em></p>
		</body>
		</html>
		`, len(sessions), len(peerSessions))
	})

	port := ":8080"
	log.Printf("📡 Server listening on http://localhost%s", port)
	log.Printf("🔗 Connect your frontend at http://localhost:3000")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("❌ Server failed to start:", err)
	}
}
