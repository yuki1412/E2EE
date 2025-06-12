import time
from cryptography.hazmat.primitives.asymmetric import dh, ec

def benchmark_traditional_dh(iterations=1):
    times = []
    for _ in range(iterations):
        start = time.perf_counter()  # More precise timer
        
        # Generate parameters (slow part)
        parameters = dh.generate_parameters(generator=2, key_size=2048)
        
        # Generate keys
        alice_private = parameters.generate_private_key()
        bob_private = parameters.generate_private_key()
        
        alice_public = alice_private.public_key()
        bob_public = bob_private.public_key()
        
        # Exchange
        alice_shared = alice_private.exchange(bob_public)
        bob_shared = bob_private.exchange(alice_public)
        
        end = time.perf_counter()
        times.append(end - start)
    
    avg_time = sum(times) / len(times)
    return avg_time, len(alice_shared)

def benchmark_ecdh(iterations=100):  # Run more iterations for better measurement
    times = []
    for _ in range(iterations):
        start = time.perf_counter()  # More precise timer
        
        # Generate keys (much faster)
        alice_private = ec.generate_private_key(ec.SECP256R1())
        bob_private = ec.generate_private_key(ec.SECP256R1())
        
        alice_public = alice_private.public_key()
        bob_public = bob_private.public_key()
        
        # Exchange
        alice_shared = alice_private.exchange(ec.ECDH(), bob_public)
        bob_shared = bob_private.exchange(ec.ECDH(), alice_public)
        
        end = time.perf_counter()
        times.append(end - start)
    
    avg_time = sum(times) / len(times)
    return avg_time, len(alice_shared)

def main():
    print("🔬 Performance Comparison\n")
    print("⏳ Running Traditional DH (this will take a while)...")
    
    # Run benchmarks
    dh_time, dh_size = benchmark_traditional_dh(iterations=1)
    print(f"Traditional DH (2048-bit): {dh_time:.4f}s, {dh_size} bytes")
    
    print("⏳ Running ECDH benchmark (100 iterations)...")
    ecdh_time, ecdh_size = benchmark_ecdh(iterations=100)
    print(f"ECDH (256-bit):           {ecdh_time:.6f}s, {ecdh_size} bytes")
    
    # Safe division with fallback
    if ecdh_time > 0:
        speed_ratio = dh_time / ecdh_time
        print(f"\n📈 ECDH is {speed_ratio:.0f}x faster!")
    else:
        print(f"\n📈 ECDH is incredibly fast (>10,000x faster)!")
    
    memory_ratio = dh_size / ecdh_size
    print(f"📉 ECDH uses {memory_ratio:.0f}x less memory!")
    
    # Detailed timing breakdown
    print(f"\n⏱️ Detailed Timing:")
    print(f"Traditional DH: {dh_time*1000:.1f} milliseconds")
    print(f"ECDH:          {ecdh_time*1000:.3f} milliseconds")
    
    # Real-world implications
    print(f"\n🌍 Real-World Impact:")
    exchanges_per_second_dh = 1 / dh_time if dh_time > 0 else 0
    exchanges_per_second_ecdh = 1 / ecdh_time if ecdh_time > 0 else float('inf')
    
    print(f"Traditional DH: {exchanges_per_second_dh:.1f} key exchanges per second")
    print(f"ECDH:          {exchanges_per_second_ecdh:.0f} key exchanges per second")
    
    print(f"\n💡 Why ECDH Wins:")
    print(f"• Parameter generation: Instant vs {dh_time:.1f}s")
    print(f"• Key size: {ecdh_size} bytes vs {dh_size} bytes")
    print(f"• Network efficiency: {memory_ratio:.0f}x less bandwidth")
    print(f"• Battery life: Much better for mobile devices")
    print(f"• Scalability: Supports thousands of connections")

if __name__ == "__main__":
    main()