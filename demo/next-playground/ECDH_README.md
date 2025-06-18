# 🔐 ECDH P-256 End-to-End Encryption Implementation

This project demonstrates a complete implementation of **Elliptic Curve Diffie-Hellman (ECDH)** key exchange using the **P-256 curve** with **AES-GCM** encryption for secure end-to-end encrypted communication.

## 🎯 What You've Built

### Frontend (React/TypeScript)
- ✅ **ECDH Service**: Complete TypeScript implementation with P-256 curve
- ✅ **Key Generation**: Ephemeral key pair generation for each session
- ✅ **Key Exchange**: Secure public key exchange simulation
- ✅ **AES Derivation**: HKDF-based AES-256-GCM key derivation from shared secret
- ✅ **Message Encryption/Decryption**: AES-GCM with random IVs
- ✅ **Interactive Demo**: Two comprehensive UI demos
- ✅ **Backend Simulation**: Mock backend integration for testing

## 🔧 Architecture Overview

```
┌─────────────────┐    Key Exchange    ┌─────────────────┐
│   Frontend      │◄------------------►│   Backend       │
│   (React)       │                    │   (Golang)      │
├─────────────────┤                    ├─────────────────┤
│ • Generate Keys │                    │ • Generate Keys │
│ • ECDH P-256    │                    │ • ECDH P-256    │
│ • AES-GCM       │                    │ • AES-GCM       │
│ • UI/UX         │                    │ • Session Mgmt  │
└─────────────────┘                    └─────────────────┘
```

## 🛡️ Security Features

- **🔑 Perfect Forward Secrecy**: New ephemeral keys for each session
- **🔐 ECDH P-256**: Industry-standard elliptic curve cryptography
- **🎯 AES-GCM-256**: Authenticated encryption with additional data
- **🔀 HKDF**: Proper key derivation from shared secret
- **🎲 Random IVs**: Unique initialization vector for each message
- **⚡ Zero Key Storage**: Private keys never leave the client/server

## 📁 File Structure

```
src/
├── app/
│   ├── services/
│   │   └── ecdh.ts              # Core ECDH implementation
│   ├── components/
│   │   ├── ECDHDemo.tsx         # Interactive ECDH demonstration
│   │   └── BackendECDH.tsx      # Backend integration simulation
│   └── page.tsx                 # Main app with tab navigation
└── ECDH_README.md              # This documentation
```

## 🚀 How to Run

1. **Start the development server:**
   ```bash
   cd nextjs/next-playground
   npm run dev
   ```

2. **Open your browser:**
   ```
   http://localhost:3000
   ```

3. **Try both demos:**
   - **ECDH Demo**: Basic key exchange and encryption
   - **Backend Integration**: Simulated client-server communication

## 🔍 How It Works

### 1. Key Exchange Process
```typescript
// 1. Generate ephemeral key pairs (both client & server)
const keyPair = await crypto.subtle.generateKey({
  name: "ECDH",
  namedCurve: "P-256",
}, true, ["deriveKey", "deriveBits"]);

// 2. Exchange public keys
const publicKeyBase64 = btoa(String.fromCharCode(...exported));

// 3. Derive shared secret using ECDH
const sharedSecret = await crypto.subtle.deriveBits({
  name: "ECDH",
  public: theirPublicKey,
}, myPrivateKey, 256);

// 4. Derive AES key using HKDF
const aesKey = await crypto.subtle.deriveKey({
  name: "HKDF",
  hash: "SHA-256",
  salt: new Uint8Array(),
  info: new TextEncoder().encode("ECDH-AES-GCM"),
}, keyMaterial, { name: "AES-GCM", length: 256 }, false, ["encrypt", "decrypt"]);
```

### 2. Message Encryption
```typescript
// Generate random IV
const iv = crypto.getRandomValues(new Uint8Array(12));

// Encrypt with AES-GCM
const encryptedData = await crypto.subtle.encrypt({
  name: "AES-GCM",
  iv: iv,
}, aesKey, messageData);
```

## 🎮 Demo Features

### 🔐 ECDH Demo Tab
- **Key Generation**: Generate Alice & Bob key pairs
- **Key Exchange**: Simulate secure public key exchange
- **Shared Secret**: View the derived shared secret
- **Message Encryption**: Encrypt messages from Alice to Bob
- **Message Decryption**: Decrypt messages using derived keys
- **Process Logs**: Real-time step-by-step process visualization

### 🌐 Backend Integration Tab
- **Connection Simulation**: Mock backend key exchange
- **Chat Interface**: Send/receive encrypted messages
- **Real-time Decryption**: Messages decrypt in real-time
- **Session Management**: Simulated session handling
- **API Documentation**: Ready-to-implement endpoints

## 🚧 Next Steps: Golang Backend

### Required Golang Packages
```go
import (
    "crypto/ecdh"
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "crypto/sha256"
    "golang.org/x/crypto/hkdf"
)
```

### Backend Endpoints to Implement

#### 1. Key Exchange Endpoint
```go
// POST /api/key-exchange
type KeyExchangeRequest struct {
    MyUserId      string `json:"myUserId"`
    OtherUserId   string `json:"otherUserId"`
    MyPublicKey   string `json:"myPublicKey"`
}

type KeyExchangeResponse struct {
    TheirPublicKey string `json:"theirPublicKey"`
    SessionId      string `json:"sessionId"`
}
```

#### 2. Message Endpoints
```go
// POST /api/messages
type SendMessageRequest struct {
    SessionId       string `json:"sessionId"`
    EncryptedData   string `json:"encryptedData"`
    IV             string `json:"iv"`
    Timestamp      int64  `json:"timestamp"`
}

// GET /api/messages/:sessionId
type GetMessagesResponse struct {
    Messages []EncryptedMessage `json:"messages"`
}
```

### Backend Implementation Steps

1. **Setup ECDH P-256 in Go:**
   ```go
   // Generate P-256 key pair
   privateKey, err := ecdh.P256().GenerateKey(rand.Reader)
   publicKey := privateKey.PublicKey()
   ```

2. **Implement Key Exchange:**
   ```go
   // Perform ECDH to get shared secret
   sharedSecret, err := privateKey.ECDH(theirPublicKey)
   
   // Derive AES key using HKDF
   aesKey := deriveAESKey(sharedSecret)
   ```

3. **Message Encryption/Decryption:**
   ```go
   // AES-GCM encryption
   block, _ := aes.NewCipher(aesKey)
   gcm, _ := cipher.NewGCM(block)
   ciphertext := gcm.Seal(nil, iv, plaintext, nil)
   ```

4. **Session Management:**
   - Store session keys in Redis/database
   - Implement session cleanup
   - Handle concurrent sessions

## 🔬 Testing & Validation

### What to Test
- [ ] Key generation consistency
- [ ] Shared secret validation (Alice & Bob must match)
- [ ] Message encryption/decryption round-trip
- [ ] Different message sizes
- [ ] Concurrent sessions
- [ ] Session expiration

### Security Considerations
- [ ] Secure key storage (never log private keys)
- [ ] Session timeout implementation
- [ ] Rate limiting on key exchange
- [ ] Input validation on all endpoints
- [ ] HTTPS only in production

## 📚 References

- [ECDH RFC 6090](https://tools.ietf.org/html/rfc6090)
- [AES-GCM RFC 5116](https://tools.ietf.org/html/rfc5116)
- [HKDF RFC 5869](https://tools.ietf.org/html/rfc5869)
- [Web Crypto API](https://developer.mozilla.org/en-US/docs/Web/API/Web_Crypto_API)

## 🎉 Congratulations!

You now have a complete frontend E2EE implementation! The next step is to mirror this exact logic in your Golang backend. The beauty of ECDH is that both sides can independently implement the same cryptographic operations and arrive at the same shared secret.

**Pro Tip**: Start with the Go implementation in your `go/` directory and test the key exchange first before building the full API. 