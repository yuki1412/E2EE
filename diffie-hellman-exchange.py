import time
from cryptography.hazmat.primitives.asymmetric import dh
from cryptography.hazmat.primitives import serialization, hashes
from cryptography.hazmat.primitives.kdf.hkdf import HKDF

class Person:
    def __init__(self, name):
        self.name = name
        self.private_key = None
        self.public_key = None
        self.shared_secret = None
        self.derived_key = None

    def generate_keys(self, parameters):
        """Generate private and public keys using real DH parameters"""
        self.private_key = parameters.generate_private_key()
        self.public_key = self.private_key.public_key()

    def get_public_key_bytes(self):
        """Serialize public key for transmission"""
        return self.public_key.public_bytes(
            encoding=serialization.Encoding.PEM,
            format=serialization.PublicFormat.SubjectPublicKeyInfo
        )

    def generate_shared_secret(self, other_public_key_bytes):
        """Generate shared secret and derive encryption key"""
        # Deserialize the other party's public key
        other_public_key = serialization.load_pem_public_key(other_public_key_bytes)
        
        # Perform the key exchange
        self.shared_secret = self.private_key.exchange(other_public_key)
        
        # Derive a proper encryption key from the shared secret
        salt = b'diffie_hellman_demo'  # In production, use random salt
        kdf = HKDF(
            algorithm=hashes.SHA256(),
            length=32,  # 256-bit key
            salt=salt,
            info=b'encryption_key'
        )
        self.derived_key = kdf.derive(self.shared_secret)

def main():
    print("=== Real Diffie-Hellman Key Exchange ===\n")
    
    generate_keys_time = time.time()

    # Generate DH parameters (2048-bit for security)
    print("Generating DH parameters (2048-bit)...")
    parameters = dh.generate_parameters(generator=2, key_size=2048)
    
    # Create participants
    alice = Person("Alice")
    bob = Person("Bob")

    print(f"Participants: {alice.name} and {bob.name}\n")

    # Generate private/public key pairs
    print("Generating private and public keys...")
    alice.generate_keys(parameters)
    bob.generate_keys(parameters)


    # Serialize public keys for "transmission over internet"
    alice_public_bytes = alice.get_public_key_bytes()
    bob_public_bytes = bob.get_public_key_bytes()

    print("Public keys generated and ready for exchange")
    print(f"Alice's public key size: {len(alice_public_bytes)} bytes")
    print(f"Bob's public key size: {len(bob_public_bytes)} bytes")
    print("\n--- Keys transmitted over the internet ---")
    print("(In real scenario, these would be sent over insecure channel)")

    # Measure exchange time
    start_time = time.time()
    
    # Exchange public keys and generate shared secrets
    print("\nGenerating shared secrets...")
    alice.generate_shared_secret(bob_public_bytes)
    bob.generate_shared_secret(alice_public_bytes)

    exchange_time = time.time() - start_time

    # Verify both parties have the same shared secret
    secrets_match = alice.shared_secret == bob.shared_secret
    derived_keys_match = alice.derived_key == bob.derived_key

    print(f"\n=== Results ===")
    print(f"Shared secrets match: {secrets_match}")
    print(f"Derived encryption keys match: {derived_keys_match}")
    print(f"Key generation time: {time.time() - generate_keys_time:.4f} seconds")
    print(f"Key exchange time: {exchange_time:.4f} seconds")
    
    if secrets_match:
        print(f"✅ Success! Both parties have established the same shared secret")
        print(f"Shared secret length: {len(alice.shared_secret)} bytes")
        print(f"Derived key length: {len(alice.derived_key)} bytes")
        print(f"Derived key (hex): {alice.derived_key.hex()[:32]}...")
    else:
        print("❌ Error: Shared secrets don't match!")

    # Security information
    print(f"\n=== Security Notes ===")
    print(f"• Using 2048-bit DH parameters (recommended minimum)")
    print(f"• Private keys are never transmitted")
    print(f"• Shared secret is derived using HKDF with SHA-256")
    print(f"• Result is a 256-bit encryption key ready for use")

if __name__ == "__main__":
    main()