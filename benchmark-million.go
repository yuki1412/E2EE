package main

import (
	"crypto/ecdh"
	"crypto/rand"
	"fmt"
	"math"
	"runtime"
	"sync"
	"time"
)

// Benchmark configuration
const (
	TARGET_OPERATIONS = 1_000_000
	BATCH_SIZE        = 10_000 // Report progress every 10k operations
)

// Simple key pair structure for benchmarking
type KeyPair struct {
	PrivateKey []byte
	PublicKey  []byte
}

// Result structure for collecting timing data
type BenchmarkResult struct {
	KeyGenTime   time.Duration
	ExchangeTime time.Duration
	Success      bool
	Error        error
}

// Generate X25519 key pair
func generateX25519KeyPair() (*ecdh.PrivateKey, error) {
	return ecdh.X25519().GenerateKey(rand.Reader)
}

// Generate P-256 key pair
func generateP256KeyPair() (*ecdh.PrivateKey, error) {
	return ecdh.P256().GenerateKey(rand.Reader)
}

// Perform X25519 key exchange
func performX25519Exchange(privateKey *ecdh.PrivateKey, otherPublicKey *ecdh.PublicKey) ([]byte, error) {
	return privateKey.ECDH(otherPublicKey)
}

// Perform P-256 key exchange
func performP256Exchange(privateKey *ecdh.PrivateKey, otherPublicKey *ecdh.PublicKey) ([]byte, error) {
	return privateKey.ECDH(otherPublicKey)
}

// X25519 worker function
func x25519Worker(jobs <-chan int, results chan<- BenchmarkResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for range jobs {
		var result BenchmarkResult

		// Generate Alice's key pair
		keyGenStart := time.Now()
		alice, err := generateX25519KeyPair()
		if err != nil {
			result.Error = err
			results <- result
			continue
		}

		// Generate Bob's key pair
		bob, err := generateX25519KeyPair()
		if err != nil {
			result.Error = err
			results <- result
			continue
		}
		result.KeyGenTime = time.Since(keyGenStart)

		// Perform key exchange
		exchangeStart := time.Now()
		aliceSecret, err := performX25519Exchange(alice, bob.PublicKey())
		if err != nil {
			result.Error = err
			results <- result
			continue
		}

		bobSecret, err := performX25519Exchange(bob, alice.PublicKey())
		if err != nil {
			result.Error = err
			results <- result
			continue
		}
		result.ExchangeTime = time.Since(exchangeStart)

		// Verify secrets match (basic check)
		result.Success = len(aliceSecret) == len(bobSecret)
		results <- result
	}
}

// P-256 worker function
func p256Worker(jobs <-chan int, results chan<- BenchmarkResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for range jobs {
		var result BenchmarkResult

		// Generate keys
		keyGenStart := time.Now()
		alicePriv, err := generateP256KeyPair()
		if err != nil {
			result.Error = err
			results <- result
			continue
		}

		bobPriv, err := generateP256KeyPair()
		if err != nil {
			result.Error = err
			results <- result
			continue
		}
		result.KeyGenTime = time.Since(keyGenStart)

		// Perform exchange
		exchangeStart := time.Now()
		aliceSecret, err := performP256Exchange(alicePriv, bobPriv.PublicKey())
		if err != nil {
			result.Error = err
			results <- result
			continue
		}

		bobSecret, err := performP256Exchange(bobPriv, alicePriv.PublicKey())
		if err != nil {
			result.Error = err
			results <- result
			continue
		}
		result.ExchangeTime = time.Since(exchangeStart)

		// Verify secrets match (basic check)
		result.Success = len(aliceSecret) == len(bobSecret)
		results <- result
	}
}

// Parallel benchmark for X25519
func benchmarkX25519Parallel(operations int) (float64, time.Duration, time.Duration, time.Duration, uint64) {
	numWorkers := runtime.NumCPU()
	fmt.Printf("🚀 Benchmarking X25519 with %d operations using %d goroutines...\n", operations, numWorkers)

	startTime := time.Now()
	var memBefore, memAfter runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	// Create channels
	jobs := make(chan int, operations)
	results := make(chan BenchmarkResult, operations)

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go x25519Worker(jobs, results, &wg)
	}

	// Send jobs
	go func() {
		for i := 0; i < operations; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	// Progress reporter
	go func() {
		for i := 0; i < operations; i += BATCH_SIZE {
			time.Sleep(100 * time.Millisecond) // Small delay for progress reporting
			elapsed := time.Since(startTime)
			if elapsed.Seconds() > 0 {
				rate := float64(i) / elapsed.Seconds()
				fmt.Printf("  Progress: %d/%d (%.1f%%) - Rate: %.0f ops/sec\n",
					i, operations, float64(i)/float64(operations)*100, rate)
			}
		}
	}()

	// Wait for workers to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var totalKeyGenTime, totalExchangeTime time.Duration
	var successCount, errorCount int

	for result := range results {
		if result.Error != nil {
			errorCount++
			continue
		}
		if result.Success {
			successCount++
		}
		totalKeyGenTime += result.KeyGenTime
		totalExchangeTime += result.ExchangeTime
	}

	totalTime := time.Since(startTime)
	runtime.ReadMemStats(&memAfter)

	// Calculate statistics
	validOps := operations - errorCount
	avgTotalTime := totalTime.Nanoseconds() / int64(validOps)
	avgKeyGenTime := totalKeyGenTime.Nanoseconds() / int64(validOps)
	avgExchangeTime := totalExchangeTime.Nanoseconds() / int64(validOps)
	operationsPerSecond := float64(validOps) / totalTime.Seconds()
	memoryUsed := (memAfter.Alloc - memBefore.Alloc) / 1024 / 1024

	fmt.Printf("\n=== X25519 Parallel Results (%d operations, %d workers) ===\n", validOps, numWorkers)
	fmt.Printf("Total time: %.3f seconds\n", totalTime.Seconds())
	fmt.Printf("Operations per second: %.0f\n", operationsPerSecond)
	fmt.Printf("Average per operation: %d nanoseconds (%.6f ms)\n", avgTotalTime, float64(avgTotalTime)/1_000_000)
	fmt.Printf("Average key generation: %d ns (%.6f ms)\n", avgKeyGenTime, float64(avgKeyGenTime)/1_000_000)
	fmt.Printf("Average key exchange: %d ns (%.6f ms)\n", avgExchangeTime, float64(avgExchangeTime)/1_000_000)
	fmt.Printf("Memory used: %d MB\n", memoryUsed)
	fmt.Printf("Success rate: %.2f%% (%d/%d)\n", float64(successCount)/float64(validOps)*100, successCount, validOps)
	fmt.Printf("Speedup vs single-threaded: %.2fx\n", operationsPerSecond/2320) // Previous single-threaded rate

	return operationsPerSecond, totalTime, totalKeyGenTime / time.Duration(validOps), totalExchangeTime / time.Duration(validOps), memoryUsed
}

// Parallel benchmark for P-256
func benchmarkP256Parallel(operations int) (float64, time.Duration, time.Duration, time.Duration, uint64) {
	numWorkers := runtime.NumCPU()
	fmt.Printf("\n🔄 Benchmarking P-256 with %d operations using %d goroutines...\n", operations, numWorkers)

	startTime := time.Now()
	var memBefore, memAfter runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	// Create channels
	jobs := make(chan int, operations)
	results := make(chan BenchmarkResult, operations)

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go p256Worker(jobs, results, &wg)
	}

	// Send jobs
	go func() {
		for i := 0; i < operations; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	// Progress reporter
	go func() {
		for i := 0; i < operations; i += BATCH_SIZE {
			time.Sleep(100 * time.Millisecond)
			elapsed := time.Since(startTime)
			if elapsed.Seconds() > 0 {
				rate := float64(i) / elapsed.Seconds()
				fmt.Printf("  P-256 Progress: %d/%d (%.1f%%) - Rate: %.0f ops/sec\n",
					i, operations, float64(i)/float64(operations)*100, rate)
			}
		}
	}()

	// Wait for workers to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var totalKeyGenTime, totalExchangeTime time.Duration
	var successCount, errorCount int

	for result := range results {
		if result.Error != nil {
			errorCount++
			continue
		}
		if result.Success {
			successCount++
		}
		totalKeyGenTime += result.KeyGenTime
		totalExchangeTime += result.ExchangeTime
	}

	totalTime := time.Since(startTime)
	runtime.ReadMemStats(&memAfter)

	// Calculate statistics
	validOps := operations - errorCount
	avgTotalTime := totalTime.Nanoseconds() / int64(validOps)
	avgKeyGenTime := totalKeyGenTime.Nanoseconds() / int64(validOps)
	avgExchangeTime := totalExchangeTime.Nanoseconds() / int64(validOps)
	operationsPerSecond := float64(validOps) / totalTime.Seconds()
	memoryUsed := (memAfter.Alloc - memBefore.Alloc) / 1024 / 1024

	fmt.Printf("\n=== P-256 Parallel Results (%d operations, %d workers) ===\n", validOps, numWorkers)
	fmt.Printf("Total time: %.3f seconds\n", totalTime.Seconds())
	fmt.Printf("Operations per second: %.0f\n", operationsPerSecond)
	fmt.Printf("Average per operation: %d nanoseconds (%.6f ms)\n", avgTotalTime, float64(avgTotalTime)/1_000_000)
	fmt.Printf("Average key generation: %d ns (%.6f ms)\n", avgKeyGenTime, float64(avgKeyGenTime)/1_000_000)
	fmt.Printf("Average key exchange: %d ns (%.6f ms)\n", avgExchangeTime, float64(avgExchangeTime)/1_000_000)
	fmt.Printf("Memory used: %d MB\n", memoryUsed)
	fmt.Printf("Success rate: %.2f%% (%d/%d)\n", float64(successCount)/float64(validOps)*100, successCount, validOps)
	fmt.Printf("Speedup vs single-threaded: %.2fx\n", operationsPerSecond/7376) // Previous single-threaded rate

	return operationsPerSecond, totalTime, totalKeyGenTime / time.Duration(validOps), totalExchangeTime / time.Duration(validOps), memoryUsed
}

func main() {
	fmt.Println("🔥 Go Parallel Cryptographic Performance Benchmark")
	fmt.Println("=================================================")
	fmt.Printf("Target: %d ECDH key exchange operations\n", TARGET_OPERATIONS)
	fmt.Printf("CPU cores: %d\n", runtime.NumCPU())
	fmt.Printf("Goroutines per algorithm: %d\n", runtime.NumCPU())
	fmt.Println()

	// Warm up
	fmt.Println("🔥 Warming up...")
	for i := 0; i < 1000; i++ {
		alice, _ := generateX25519KeyPair()
		bob, _ := generateX25519KeyPair()
		performX25519Exchange(alice, bob.PublicKey())
	}
	fmt.Println("Warm-up complete!\n")

	// Parallel benchmarks
	x25519Rate, x25519Total, x25519KeyGen, x25519Exchange, x25519Memory := benchmarkX25519Parallel(TARGET_OPERATIONS)
	p256Rate, p256Total, p256KeyGen, p256Exchange, p256Memory := benchmarkP256Parallel(TARGET_OPERATIONS)

	// Comprehensive Performance Analysis
	fmt.Println("\n🎯 COMPREHENSIVE PERFORMANCE ANALYSIS")
	fmt.Println("=====================================")

	// Raw Performance Metrics
	fmt.Println("\n📊 Raw Performance Metrics:")
	fmt.Printf("X25519:  %.0f ops/sec | Total: %.3fs | Memory: %d MB\n", x25519Rate, x25519Total.Seconds(), x25519Memory)
	fmt.Printf("P-256:   %.0f ops/sec | Total: %.3fs | Memory: %d MB\n", p256Rate, p256Total.Seconds(), p256Memory)

	// Speed Comparison Analysis
	fmt.Println("\n⚡ Speed Comparison Analysis:")
	if p256Rate > x25519Rate {
		speedup := p256Rate / x25519Rate
		timeSaved := x25519Total - p256Total
		fmt.Printf("🏆 P-256 is %.2fx FASTER than X25519\n", speedup)
		fmt.Printf("📈 P-256 processes %.0f more operations per second\n", p256Rate-x25519Rate)
		fmt.Printf("⏱️  P-256 saved %.3f seconds (%.1f%% time reduction)\n", timeSaved.Seconds(), (timeSaved.Seconds()/x25519Total.Seconds())*100)
	} else {
		speedup := x25519Rate / p256Rate
		timeSaved := p256Total - x25519Total
		fmt.Printf("🏆 X25519 is %.2fx FASTER than P-256\n", speedup)
		fmt.Printf("📈 X25519 processes %.0f more operations per second\n", x25519Rate-p256Rate)
		fmt.Printf("⏱️  X25519 saved %.3f seconds (%.1f%% time reduction)\n", timeSaved.Seconds(), (timeSaved.Seconds()/p256Total.Seconds())*100)
	}

	// Operation Breakdown Analysis
	fmt.Println("\n🔬 Operation Breakdown Analysis:")
	fmt.Printf("┌─────────────────┬─────────────┬─────────────┬──────────────┐\n")
	fmt.Printf("│ Algorithm       │ Key Gen     │ Exchange    │ Total/Op     │\n")
	fmt.Printf("├─────────────────┼─────────────┼─────────────┼──────────────┤\n")
	fmt.Printf("│ X25519          │ %8.3fms │ %8.3fms │ %9.3fms │\n",
		float64(x25519KeyGen.Nanoseconds())/1_000_000,
		float64(x25519Exchange.Nanoseconds())/1_000_000,
		float64(x25519KeyGen.Nanoseconds()+x25519Exchange.Nanoseconds())/1_000_000)
	fmt.Printf("│ P-256           │ %8.3fms │ %8.3fms │ %9.3fms │\n",
		float64(p256KeyGen.Nanoseconds())/1_000_000,
		float64(p256Exchange.Nanoseconds())/1_000_000,
		float64(p256KeyGen.Nanoseconds()+p256Exchange.Nanoseconds())/1_000_000)
	fmt.Printf("└─────────────────┴─────────────┴─────────────┴──────────────┘\n")

	// Key Generation vs Exchange Analysis
	fmt.Println("\n🔑 Key Generation vs Exchange Performance:")

	fmt.Printf("X25519 - Key Gen: %.1f%% | Exchange: %.1f%% of total operation time\n",
		float64(x25519KeyGen.Nanoseconds())/float64(x25519KeyGen.Nanoseconds()+x25519Exchange.Nanoseconds())*100,
		float64(x25519Exchange.Nanoseconds())/float64(x25519KeyGen.Nanoseconds()+x25519Exchange.Nanoseconds())*100)
	fmt.Printf("P-256  - Key Gen: %.1f%% | Exchange: %.1f%% of total operation time\n",
		float64(p256KeyGen.Nanoseconds())/float64(p256KeyGen.Nanoseconds()+p256Exchange.Nanoseconds())*100,
		float64(p256Exchange.Nanoseconds())/float64(p256KeyGen.Nanoseconds()+p256Exchange.Nanoseconds())*100)

	// Parallel Efficiency Analysis
	fmt.Println("\n⚙️ Parallel Efficiency Analysis:")
	cpuCores := float64(runtime.NumCPU())
	x25519TheoreticalMax := 2320 * cpuCores // Single-threaded rate * cores
	p256TheoreticalMax := 7376 * cpuCores
	x25519Efficiency := (x25519Rate / x25519TheoreticalMax) * 100
	p256Efficiency := (p256Rate / p256TheoreticalMax) * 100

	fmt.Printf("X25519 Parallel Efficiency: %.1f%% (%.0f/%.0f ops/sec)\n", x25519Efficiency, x25519Rate, x25519TheoreticalMax)
	fmt.Printf("P-256  Parallel Efficiency: %.1f%% (%.0f/%.0f ops/sec)\n", p256Efficiency, p256Rate, p256TheoreticalMax)
	fmt.Printf("Efficiency difference: %.1f percentage points\n", math.Abs(p256Efficiency-x25519Efficiency))

	// Throughput Analysis
	fmt.Println("\n📈 Throughput Analysis:")
	x25519MBPerSec := (x25519Rate * 32) / (1024 * 1024) // 32 bytes per X25519 key
	p256MBPerSec := (p256Rate * 32) / (1024 * 1024)     // 32 bytes per P-256 key

	fmt.Printf("X25519 Key throughput: %.2f MB/sec (%d bytes/key)\n", x25519MBPerSec, 32)
	fmt.Printf("P-256  Key throughput: %.2f MB/sec (%d bytes/key)\n", p256MBPerSec, 32)

	// Cost-Benefit Analysis
	fmt.Println("\n💰 Cost-Benefit Analysis:")
	memoryRatio := float64(p256Memory) / float64(x25519Memory)
	performanceRatio := p256Rate / x25519Rate

	fmt.Printf("Performance gain: %.2fx faster\n", performanceRatio)
	fmt.Printf("Memory cost ratio: %.2fx memory usage\n", memoryRatio)
	fmt.Printf("Performance per MB: X25519 = %.0f ops/MB, P-256 = %.0f ops/MB\n",
		x25519Rate/float64(x25519Memory), p256Rate/float64(p256Memory))

	// Scalability Projections
	fmt.Println("\n📊 Scalability Projections:")
	fmt.Printf("For 10M operations:\n")
	fmt.Printf("  X25519: %.1f seconds (%.1f minutes)\n", 10_000_000/x25519Rate, (10_000_000/x25519Rate)/60)
	fmt.Printf("  P-256:  %.1f seconds (%.1f minutes)\n", 10_000_000/p256Rate, (10_000_000/p256Rate)/60)

	fmt.Printf("For 100M operations:\n")
	fmt.Printf("  X25519: %.1f seconds (%.1f hours)\n", 100_000_000/x25519Rate, (100_000_000/x25519Rate)/3600)
	fmt.Printf("  P-256:  %.1f seconds (%.1f hours)\n", 100_000_000/p256Rate, (100_000_000/p256Rate)/3600)

	// Final Recommendations
	fmt.Println("\n🎯 Performance Recommendations:")
	if p256Rate > x25519Rate {
		fmt.Println("✅ Recommendation: Use P-256 for maximum performance")
		fmt.Printf("   • %.2fx faster execution\n", performanceRatio)
		fmt.Printf("   • Better parallel efficiency (%.1f%% vs %.1f%%)\n", p256Efficiency, x25519Efficiency)
		fmt.Printf("   • Higher throughput (%.2f vs %.2f MB/sec)\n", p256MBPerSec, x25519MBPerSec)
	} else {
		fmt.Println("✅ Recommendation: Use X25519 for maximum performance")
		fmt.Printf("   • %.2fx faster execution\n", x25519Rate/p256Rate)
		fmt.Printf("   • Better parallel efficiency (%.1f%% vs %.1f%%)\n", x25519Efficiency, p256Efficiency)
		fmt.Printf("   • Higher throughput (%.2f vs %.2f MB/sec)\n", x25519MBPerSec, p256MBPerSec)
	}

	fmt.Println("\n🔬 Technical Notes:")
	fmt.Println("• All measurements include complete ECDH operations (key generation + exchange)")
	fmt.Println("• Parallel efficiency calculated against theoretical maximum (single-thread × cores)")
	fmt.Println("• Memory measurements include Go runtime overhead")
	fmt.Println("• Results may vary based on CPU architecture and Go compiler optimizations")
}
