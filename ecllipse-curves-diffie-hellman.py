from cryptography.hazmat.primitives.asymmetric import ec, x25519
from cryptography.hazmat.primitives import serialization, hashes
from cryptography.hazmat.primitives.kdf.hkdf import HKDF
import time

class Person:
    def __init__(self, name):
        self.name = name
        self.private_key = None
        self.public_key = None
        self.shared_secret = None
        self.derived_key = None
        self.curve_type = None

    def generate_keys(self, curve_name='P-256'):
        """Generate private and public keys using ECDH or X25519"""
        self.curve_type = curve_name
        
        if curve_name == 'X25519':
            # X25519 has different API
            self.private_key = x25519.X25519PrivateKey.generate()
            self.public_key = self.private_key.public_key()
        else:
            # Regular ECDH curves
            curves = {
                'P-256': ec.SECP256R1(),
                'P-384': ec.SECP384R1(), 
                'P-521': ec.SECP521R1(),
            }
            curve = curves[curve_name]
            self.private_key = ec.generate_private_key(curve)
            self.public_key = self.private_key.public_key()

    def get_public_key_bytes(self):
        """Serialize public key for transmission"""
        if self.curve_type == 'X25519':
            # X25519 uses Raw encoding
            return self.public_key.public_bytes(
                encoding=serialization.Encoding.Raw,
                format=serialization.PublicFormat.Raw
            )
        else:
            # Regular ECDH uses PEM
            return self.public_key.public_bytes(
                encoding=serialization.Encoding.PEM,
                format=serialization.PublicFormat.SubjectPublicKeyInfo
            )

    def generate_shared_secret(self, other_public_key_bytes):
        """Generate shared secret and derive encryption key"""
        if self.curve_type == 'X25519':
            # X25519 key exchange
            other_public_key = x25519.X25519PublicKey.from_public_bytes(other_public_key_bytes)
            self.shared_secret = self.private_key.exchange(other_public_key)
        else:
            # Regular ECDH key exchange
            other_public_key = serialization.load_pem_public_key(other_public_key_bytes)
            self.shared_secret = self.private_key.exchange(ec.ECDH(), other_public_key)
        
        # Derive a proper encryption key from the shared secret
        salt = b'ecdh_demo_salt'
        kdf = HKDF(
            algorithm=hashes.SHA256(),
            length=32,  # 256-bit key
            salt=salt,
            info=b'encryption_key'
        )
        self.derived_key = kdf.derive(self.shared_secret)

def main():
    print("=== ECDH Key Exchange with Multiple Curves ===\n")
    
    # Available curves
    available_curves = {
        'P-256': 'NIST P-256 (secp256r1) - Most widely supported',
        'P-384': 'NIST P-384 (secp384r1) - Higher security', 
        'P-521': 'NIST P-521 (secp521r1) - Highest NIST security',
        'X25519': 'Curve25519 (X25519) - Used by WhatsApp, Signal, etc.'
    }
    
    # Choose X25519 as requested
    curve_name = 'P-521'
    
    print(f"Using curve: {curve_name}")
    print(f"Description: {available_curves[curve_name]}")
    print()
    
    # Create participants
    alice = Person("Alice")
    bob = Person("Bob")
    print(f"Participants: {alice.name} and {bob.name}\n")

    # Measure key generation time
    start_time = time.perf_counter()
    
    print(f"Generating {curve_name} key pairs...")
    alice.generate_keys(curve_name)
    bob.generate_keys(curve_name)
    
    keygen_time = time.perf_counter() - start_time

    # Serialize public keys for "transmission over internet"
    alice_public_bytes = alice.get_public_key_bytes()
    bob_public_bytes = bob.get_public_key_bytes()

    print("Public keys generated and ready for exchange")
    print(f"Alice's public key size: {len(alice_public_bytes)} bytes")
    print(f"Bob's public key size: {len(bob_public_bytes)} bytes")
    print(f"Key generation time: {keygen_time:.6f} seconds")
    print("\n--- Keys transmitted over the internet ---")
    print("(X25519 has very compact 32-byte public keys!)")

    # Measure exchange time
    start_time = time.perf_counter()
    
    print(f"\nPerforming {curve_name} key exchange...")
    alice.generate_shared_secret(bob_public_bytes)
    bob.generate_shared_secret(alice_public_bytes)
    
    exchange_time = time.perf_counter() - start_time

    # Verify both parties have the same shared secret
    secrets_match = alice.shared_secret == bob.shared_secret
    derived_keys_match = alice.derived_key == bob.derived_key

    print(f"\n=== Results ===")
    print(f"Shared secrets match: {secrets_match}")
    print(f"Derived encryption keys match: {derived_keys_match}")
    print(f"Key exchange time: {exchange_time:.6f} seconds")
    
    if secrets_match:
        print(f"✅ Success! Both parties have established the same shared secret")
        print(f"Shared secret length: {len(alice.shared_secret)} bytes")
        print(f"Derived key length: {len(alice.derived_key)} bytes")
        print(f"Derived key (hex): {alice.derived_key.hex()[:32]}...")
    else:
        print("❌ Error: Shared secrets don't match!")

    # X25519 specific information
    print(f"\n=== X25519 (Curve25519) Advantages ===")
    print(f"• Ultra-compact: 32-byte public keys (vs 178+ bytes for NIST curves)")
    print(f"• Lightning fast: Optimized for speed")
    print(f"• Security: ~128-bit security level")
    print(f"• Real-world proven: Used by Signal, WhatsApp, WireGuard")
    print(f"• Side-channel resistant: Safer against timing attacks")
    print(f"• No cofactor issues: Simpler and safer implementation")
    print(f"• Perfect for: Mobile apps, IoT, real-time communication")

if __name__ == "__main__":
    main()