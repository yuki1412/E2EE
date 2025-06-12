package main

import (
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"time"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
)

type Person struct {
	Name       string
	PrivateKey interface{}
	PublicKey  interface{}
	SharedKey  []byte
	DerivedKey []byte
	CurveType  string
}

// Generate keys based on curve type
func (p *Person) GenerateKeys(curveName string) error {
	p.CurveType = curveName

	switch curveName {
	case "X25519":
		return p.generateX25519Keys()
	case "P-256":
		return p.generateECDHKeys(ecdh.P256())
	case "P-384":
		return p.generateECDHKeys(ecdh.P384())
	case "P-521":
		return p.generateECDHKeys(ecdh.P521())
	default:
		return fmt.Errorf("unsupported curve: %s", curveName)
	}
}

// Generate X25519 keys
func (p *Person) generateX25519Keys() error {
	// Generate private key (32 bytes)
	privateKey := make([]byte, 32)
	if _, err := rand.Read(privateKey); err != nil {
		return fmt.Errorf("failed to generate private key: %v", err)
	}

	// Generate public key
	publicKey, err := curve25519.X25519(privateKey, curve25519.Basepoint)
	if err != nil {
		return fmt.Errorf("failed to generate public key: %v", err)
	}

	p.PrivateKey = privateKey
	p.PublicKey = publicKey
	return nil
}

// Generate ECDH keys for NIST curves
func (p *Person) generateECDHKeys(curve ecdh.Curve) error {
	privateKey, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate ECDH key: %v", err)
	}

	p.PrivateKey = privateKey
	p.PublicKey = privateKey.PublicKey()
	return nil
}

// Get public key bytes for transmission
func (p *Person) GetPublicKeyBytes() ([]byte, error) {
	switch p.CurveType {
	case "X25519":
		if pubKey, ok := p.PublicKey.([]byte); ok {
			return pubKey, nil
		}
		return nil, fmt.Errorf("invalid X25519 public key type")

	case "P-256", "P-384", "P-521":
		if pubKey, ok := p.PublicKey.(*ecdh.PublicKey); ok {
			return pubKey.Bytes(), nil
		}
		return nil, fmt.Errorf("invalid ECDH public key type")

	default:
		return nil, fmt.Errorf("unsupported curve type: %s", p.CurveType)
	}
}

// Generate shared secret with another party's public key
func (p *Person) GenerateSharedSecret(otherPublicKeyBytes []byte) error {
	var sharedSecret []byte
	var err error

	switch p.CurveType {
	case "X25519":
		sharedSecret, err = p.computeX25519SharedSecret(otherPublicKeyBytes)
	case "P-256", "P-384", "P-521":
		sharedSecret, err = p.computeECDHSharedSecret(otherPublicKeyBytes)
	default:
		return fmt.Errorf("unsupported curve type: %s", p.CurveType)
	}

	if err != nil {
		return err
	}

	p.SharedKey = sharedSecret

	// Derive encryption key using HKDF
	return p.deriveEncryptionKey()
}

// Compute X25519 shared secret
func (p *Person) computeX25519SharedSecret(otherPublicKey []byte) ([]byte, error) {
	privateKey, ok := p.PrivateKey.([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid X25519 private key type")
	}

	sharedSecret, err := curve25519.X25519(privateKey, otherPublicKey)
	if err != nil {
		return nil, fmt.Errorf("X25519 key exchange failed: %v", err)
	}

	return sharedSecret, nil
}

// Compute ECDH shared secret for NIST curves
func (p *Person) computeECDHSharedSecret(otherPublicKeyBytes []byte) ([]byte, error) {
	privateKey, ok := p.PrivateKey.(*ecdh.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("invalid ECDH private key type")
	}

	// Import the other party's public key
	otherPublicKey, err := privateKey.Curve().NewPublicKey(otherPublicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to import public key: %v", err)
	}

	// Perform ECDH
	sharedSecret, err := privateKey.ECDH(otherPublicKey)
	if err != nil {
		return nil, fmt.Errorf("ECDH operation failed: %v", err)
	}

	return sharedSecret, nil
}

// Derive encryption key from shared secret using HKDF
func (p *Person) deriveEncryptionKey() error {
	salt := []byte("ecdh_demo_salt")
	info := []byte("encryption_key")

	// Create HKDF
	kdf := hkdf.New(sha256.New, p.SharedKey, salt, info)

	// Derive 32-byte key
	derivedKey := make([]byte, 32)
	if _, err := io.ReadFull(kdf, derivedKey); err != nil {
		return fmt.Errorf("key derivation failed: %v", err)
	}

	p.DerivedKey = derivedKey
	return nil
}

// Get curve information
func getCurveInfo() map[string]string {
	return map[string]string{
		"P-256":  "NIST P-256 (secp256r1) - Most widely supported",
		"P-384":  "NIST P-384 (secp384r1) - Higher security",
		"P-521":  "NIST P-521 (secp521r1) - Highest NIST security",
		"X25519": "Curve25519 (X25519) - Used by WhatsApp, Signal, etc.",
	}
}

// Compare byte slices
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Convert bytes to hex string (first n bytes)
func bytesToHex(data []byte, limit int) string {
	if len(data) > limit {
		return fmt.Sprintf("%x...", data[:limit])
	}
	return fmt.Sprintf("%x", data)
}

func main() {
	fmt.Println("=== Go ECDH Key Exchange with Multiple Curves ===\n")

	// Available curves
	availableCurves := getCurveInfo()

	// Choose curve (can be changed)
	curveName := "X25519"

	fmt.Printf("Using curve: %s\n", curveName)
	fmt.Printf("Description: %s\n", availableCurves[curveName])
	fmt.Println()

	// Create participants
	alice := &Person{Name: "Alice"}
	bob := &Person{Name: "Bob"}
	fmt.Printf("Participants: %s and %s\n\n", alice.Name, bob.Name)

	// Measure key generation time
	startTime := time.Now()

	fmt.Printf("Generating %s key pairs...\n", curveName)
	if err := alice.GenerateKeys(curveName); err != nil {
		fmt.Printf("Failed to generate Alice's keys: %v\n", err)
		return
	}
	if err := bob.GenerateKeys(curveName); err != nil {
		fmt.Printf("Failed to generate Bob's keys: %v\n", err)
		return
	}

	keygenTime := time.Since(startTime)

	// Serialize public keys for "transmission over internet"
	alicePublicBytes, err := alice.GetPublicKeyBytes()
	if err != nil {
		fmt.Printf("Failed to get Alice's public key: %v\n", err)
		return
	}

	bobPublicBytes, err := bob.GetPublicKeyBytes()
	if err != nil {
		fmt.Printf("Failed to get Bob's public key: %v\n", err)
		return
	}

	fmt.Println("Public keys generated and ready for exchange")
	fmt.Printf("Alice's public key size: %d bytes\n", len(alicePublicBytes))
	fmt.Printf("Bob's public key size: %d bytes\n", len(bobPublicBytes))
	fmt.Printf("Key generation time: %.6f seconds\n", keygenTime.Seconds())
	fmt.Println("\n--- Keys transmitted over the internet ---")
	if curveName == "X25519" {
		fmt.Println("(X25519 has very compact 32-byte public keys!)")
	} else {
		fmt.Printf("(%s public keys transmitted)\n", curveName)
	}

	// Measure exchange time
	startTime = time.Now()

	fmt.Printf("\nPerforming %s key exchange...\n", curveName)
	if err := alice.GenerateSharedSecret(bobPublicBytes); err != nil {
		fmt.Printf("Alice's key exchange failed: %v\n", err)
		return
	}
	if err := bob.GenerateSharedSecret(alicePublicBytes); err != nil {
		fmt.Printf("Bob's key exchange failed: %v\n", err)
		return
	}

	exchangeTime := time.Since(startTime)

	// Verify both parties have the same shared secret
	secretsMatch := bytesEqual(alice.SharedKey, bob.SharedKey)
	derivedKeysMatch := bytesEqual(alice.DerivedKey, bob.DerivedKey)

	fmt.Println("\n=== Results ===")
	fmt.Printf("Shared secrets match: %v\n", secretsMatch)
	fmt.Printf("Derived encryption keys match: %v\n", derivedKeysMatch)
	fmt.Printf("Key exchange time: %.6f seconds\n", exchangeTime.Seconds())

	if secretsMatch {
		fmt.Println("✅ Success! Both parties have established the same shared secret")
		fmt.Printf("Shared secret length: %d bytes\n", len(alice.SharedKey))
		fmt.Printf("Derived key length: %d bytes\n", len(alice.DerivedKey))
		fmt.Printf("Derived key (hex): %s\n", bytesToHex(alice.DerivedKey, 16))
	} else {
		fmt.Println("❌ Error: Shared secrets don't match!")
	}

	// Display curve-specific advantages
	switch curveName {
	case "X25519":
		fmt.Println("\n=== X25519 (Curve25519) Advantages ===")
		fmt.Println("• Ultra-compact: 32-byte public keys (vs 65+ bytes for NIST curves)")
		fmt.Println("• Lightning fast: Optimized for speed")
		fmt.Println("• Security: ~128-bit security level")
		fmt.Println("• Real-world proven: Used by Signal, WhatsApp, WireGuard")
		fmt.Println("• Side-channel resistant: Safer against timing attacks")
		fmt.Println("• No cofactor issues: Simpler and safer implementation")
		fmt.Println("• Perfect for: Mobile apps, IoT, real-time communication")
	case "P-256":
		fmt.Println("\n=== P-256 Advantages ===")
		fmt.Println("• Industry standard: Widely supported")
		fmt.Println("• Good performance: Faster than larger curves")
		fmt.Println("• Broad compatibility: Works everywhere")
		fmt.Println("• NIST approved: Government/enterprise ready")
	case "P-384":
		fmt.Println("\n=== P-384 Advantages ===")
		fmt.Println("• Higher security: ~192-bit security level")
		fmt.Println("• Future-proof: Resistant to quantum advances")
		fmt.Println("• Enterprise grade: Used in high-security environments")
		fmt.Println("• NIST Suite B: Government approved")
	case "P-521":
		fmt.Println("\n=== P-521 Advantages ===")
		fmt.Println("• Maximum security: ~256-bit security level")
		fmt.Println("• Quantum resistant: Longest security timeline")
		fmt.Println("• Premium protection: For highest-value assets")
		fmt.Println("• Future-ready: Designed for long-term use")
	}

	// Performance summary
	fmt.Printf("\n=== Performance Summary ===\n")
	fmt.Printf("Total execution time: %.6f seconds\n", keygenTime.Seconds()+exchangeTime.Seconds())
	fmt.Printf("Key generation: %.2f%%\n", (keygenTime.Seconds()/(keygenTime.Seconds()+exchangeTime.Seconds()))*100)
	fmt.Printf("Key exchange: %.2f%%\n", (exchangeTime.Seconds()/(keygenTime.Seconds()+exchangeTime.Seconds()))*100)
}
