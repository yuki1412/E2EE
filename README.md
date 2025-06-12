# 🔐 Diffie-Hellman Key Exchange Visualizer

An interactive demonstration of Diffie-Hellman key exchange protocols with both educational and real-world cryptographic implementations.

![Diffie-Hellman Visualizer](https://img.shields.io/badge/Crypto-Diffie--Hellman-blue) ![Web Crypto API](https://img.shields.io/badge/WebCrypto-API-green) ![Python](https://img.shields.io/badge/Python-3.8+-yellow)

## 📋 Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Components](#components)
- [Getting Started](#getting-started)
- [Usage](#usage)
- [Algorithms Supported](#algorithms-supported)
- [Performance Comparison](#performance-comparison)
- [Educational Resources](#educational-resources)
- [Browser Compatibility](#browser-compatibility)
- [Security Notes](#security-notes)
- [Contributing](#contributing)

## 🌟 Overview

This project provides multiple implementations of Diffie-Hellman key exchange:

1. **Educational Python Implementation** - Simple DH with small numbers for learning
2. **Production Python Implementation** - Real cryptographic DH and ECDH using the `cryptography` library
3. **Interactive HTML Visualizer** - Browser-based demonstration with real Web Crypto API calculations

Perfect for students, educators, and developers who want to understand how secure key exchange works!

## ✨ Features

### 🎓 Educational
- **Step-by-step visualization** of the key exchange process
- **Mathematical formulas** shown for each operation
- **Simple numbers** for easy understanding (educational mode)
- **Real cryptographic calculations** using industry standards

### 🔒 Cryptographically Secure
- **Web Crypto API** integration for real browser-based cryptography
- **Multiple curve support**: P-256, P-384, X25519
- **Performance benchmarking** between different algorithms
- **Actual key verification** - not simulated!

### 🎨 Interactive UI
- **Beautiful, modern interface** with animations
- **Progressive disclosure** of information
- **Mobile-responsive** design
- **Real-time progress tracking**

## 📁 Components

### 1. HTML Visualizer (`dh-visualizer.html`)
Interactive browser-based demonstration with four algorithms:
- Simple DH (educational)
- ECDH P-256 (industry standard)
- ECDH P-384 (high security)
- X25519 (modern, ultra-fast)

### 2. Python Implementations

#### Traditional DH (`diffie-hellman-exchange.py`)
```python
# Real 2048-bit Diffie-Hellman with proper key derivation
python diffie-hellman-exchange.py
```

#### ECDH Implementation (`ecllipse-curves-diffie-hellman.py`)
```python
# Elliptic Curve Diffie-Hellman with multiple curves
python ecllipse-curves-diffie-hellman.py
```

#### Performance Comparison (`performance_comparison.py`)
```python
# Benchmark Traditional DH vs ECDH performance
python performance_comparison.py
```

## 🚀 Getting Started

### Prerequisites

**For Python implementations:**
```bash
pip install cryptography
```

**For HTML visualizer:**
- Modern web browser with Web Crypto API support
- No additional dependencies required!

### Quick Start

1. **Clone or download** the project files
2. **For Python demos:**
   ```bash
   cd python-dh
   pip install cryptography
   python diffie-hellman-exchange.py
   ```

3. **For HTML visualizer:**
   - Open `dh-visualizer.html` in your browser
   - Select an algorithm
   - Click "🚀 Start Key Exchange"

## 🎯 Usage

### HTML Visualizer

1. **Open** `dh-visualizer.html` in a modern browser
2. **Select** from four algorithms:
   - Simple DH (educational)
   - ECDH P-256 (real crypto)
   - ECDH P-384 (higher security) 
   - X25519 (modern/fastest)
3. **Click** "Start Key Exchange"
4. **Watch** the step-by-step process with real calculations!

### Python Scripts

#### Basic DH Exchange
```bash
python diffie-hellman-exchange.py
```
**Output:**
```
=== Real Diffie-Hellman Key Exchange ===

Generating DH parameters (2048-bit)...
Participants: Alice and Bob

✅ Success! Both parties have established the same shared secret
Key generation time: 15.2341 seconds
```

#### ECDH Exchange
```bash
python ecllipse-curves-diffie-hellman.py
```
**Output:**
```
=== ECDH Key Exchange ===

Using curve: X25519
✅ Success! Both parties have established the same shared secret
Key generation time: 0.0012 seconds
```

#### Performance Comparison
```bash
python performance_comparison.py
```
**Output:**
```
🔬 Performance Comparison

Traditional DH (2048-bit): 19.493s, 256 bytes
ECDH (256-bit):           0.001s, 32 bytes

📈 ECDH is 15,653x faster!
📉 ECDH uses 8x less memory!
```

## 🔧 Algorithms Supported

| Algorithm | Key Size | Security Level | Speed | Use Case |
|-----------|----------|----------------|-------|----------|
| **Simple DH** | Custom | Educational | Slow | Learning |
| **Traditional DH** | 2048-bit | ~112-bit | Slow | Legacy systems |
| **ECDH P-256** | 256-bit | ~128-bit | Fast | Industry standard |
| **ECDH P-384** | 384-bit | ~192-bit | Fast | High security |
| **X25519** | 255-bit | ~128-bit | Fastest | Modern apps |

### Real-World Usage
- **P-256**: TLS, most web browsers
- **P-384**: Government, high-security applications
- **X25519**: Signal, WhatsApp, WireGuard VPN

## ⚡ Performance Comparison

Based on our benchmarks:

```
Algorithm        | Speed      | Key Size | Real-World Usage
-----------------|------------|----------|------------------
Traditional DH   | 19.5s      | 256B     | Legacy systems
ECDH P-256      | 0.001s     | 32B      | Web browsers
ECDH P-384      | 0.002s     | 48B      | Government
X25519          | 0.0005s    | 32B      | Signal, WhatsApp
```

**Key Insights:**
- ECDH is **15,000x faster** than traditional DH
- X25519 has **ultra-compact 32-byte keys**
- All algorithms produce cryptographically secure results

## 📚 Educational Resources

### What You'll Learn
1. **Public Key Cryptography Basics**
   - How two parties establish shared secrets
   - Why private keys never travel over the network
   - The mathematical foundation of security

2. **Elliptic Curve Advantages**
   - Why ECDH is faster and more efficient
   - Comparison of different curve types
   - Real-world performance implications

3. **Modern Cryptography**
   - X25519 and its advantages
   - Web Crypto API usage
   - Browser-based cryptography

### Step-by-Step Process
The visualizer shows exactly how DH works:

1. **📢 Public Parameters** - Establish shared mathematical constants
2. **🔑 Private Key Generation** - Each party creates secret keys
3. **📤 Public Key Calculation** - Derive shareable keys from private keys
4. **🌐 Key Exchange** - Send public keys over insecure channel
5. **🔐 Shared Secret Calculation** - Both compute same secret independently
6. **✅ Verification** - Confirm both parties have identical secrets

## 🌐 Browser Compatibility

The HTML visualizer requires Web Crypto API support:

| Browser | Version | Status |
|---------|---------|--------|
| Chrome | 37+ | ✅ Full support |
| Firefox | 34+ | ✅ Full support |
| Safari | 7+ | ✅ Full support |
| Edge | 12+ | ✅ Full support |

**Note:** X25519 support requires newer browser versions (Chrome 70+, Firefox 72+)

## 🔒 Security Notes

### ⚠️ Important Disclaimers

1. **Educational Purpose**: The simple DH implementation uses small numbers for learning - never use in production!

2. **Real Cryptography**: ECDH implementations use proper cryptographic libraries and are suitable for understanding real-world usage.

3. **Web Crypto API**: Browser implementations are secure and use the same algorithms as production systems.

4. **Key Storage**: In real applications, private keys should never be displayed or logged.

### 🛡️ Best Practices Demonstrated

- ✅ Cryptographically secure random number generation
- ✅ Proper key derivation (HKDF)
- ✅ Standard curve parameters (NIST, RFC)
- ✅ Secure key exchange verification

## 🤝 Contributing

Contributions welcome! Areas for improvement:

- Additional curve support (secp256k1, Ed25519)
- More educational animations
- Additional language implementations
- Mobile app version
- Advanced security features

### Development Setup
```bash
git clone <repository>
cd python-dh
pip install cryptography
# Open dh-visualizer.html in browser for testing
```

## 📄 License

This project is open source and available under the MIT License.

## 🙏 Acknowledgments

- **NIST** for standardized elliptic curves
- **Signal Foundation** for X25519 adoption
- **Web Crypto API** working group
- **Python Cryptography** library maintainers

---

**🔐 Secure key exchange made simple and visual!**

*Perfect for educational use, demonstrations, and understanding modern cryptography.* 