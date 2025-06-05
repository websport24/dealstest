package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type TestResult struct {
	TargetRPS     int
	ActualRPS     float64
	SuccessRate   float64
	TotalRequests int64
	Successful    int64
	Failed        int64
	AvgLatency    time.Duration
	P95Latency    time.Duration
}

func runTest(baseURL string, targetRPS int, duration time.Duration) TestResult {
	fmt.Printf("🔥 Testing %d RPS for %v...\n", targetRPS, duration)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var totalRequests int64
	var successful int64
	var failed int64
	var latencies []time.Duration
	var latenciesMu sync.Mutex

	startTime := time.Now()
	endTime := startTime.Add(duration)

	// Вычисляем параметры для высокой нагрузки
	numWorkers := targetRPS / 10 // Примерно 10 RPS на воркера
	if numWorkers < 1 {
		numWorkers = 1
	}
	if numWorkers > 500 { // Ограничиваем количество воркеров
		numWorkers = 500
	}

	rpsPerWorker := targetRPS / numWorkers
	interval := time.Second / time.Duration(rpsPerWorker)

	fmt.Printf("   Workers: %d, RPS per worker: %d, Interval: %v\n", numWorkers, rpsPerWorker, interval)

	var wg sync.WaitGroup
	url := fmt.Sprintf("%s/counter/1", baseURL)

	// Запускаем пул воркеров для генерации высокой нагрузки
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// Смещение старта для равномерного распределения
			startDelay := time.Duration(workerID) * interval / time.Duration(numWorkers)
			time.Sleep(startDelay)

			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					if time.Now().After(endTime) {
						return
					}

					// Выполняем запрос асинхронно
					go func() {
						requestStart := time.Now()
						resp, err := client.Get(url)
						latency := time.Since(requestStart)

						atomic.AddInt64(&totalRequests, 1)

						if err != nil {
							atomic.AddInt64(&failed, 1)
						} else {
							io.Copy(io.Discard, resp.Body)
							resp.Body.Close()

							if resp.StatusCode >= 200 && resp.StatusCode < 300 {
								atomic.AddInt64(&successful, 1)
							} else {
								atomic.AddInt64(&failed, 1)
							}
						}

						latenciesMu.Lock()
						latencies = append(latencies, latency)
						latenciesMu.Unlock()
					}()
				}
			}
		}(i)
	}

	wg.Wait()

	actualDuration := time.Since(startTime)
	finalTotal := atomic.LoadInt64(&totalRequests)
	finalSuccessful := atomic.LoadInt64(&successful)
	finalFailed := atomic.LoadInt64(&failed)

	actualRPS := float64(finalTotal) / actualDuration.Seconds()
	successRate := 0.0
	if finalTotal > 0 {
		successRate = float64(finalSuccessful) / float64(finalTotal) * 100
	}

	// Вычисляем латентность
	latenciesMu.Lock()
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	var avgLatency, p95 time.Duration
	if len(latencies) > 0 {
		var total time.Duration
		for _, lat := range latencies {
			total += lat
		}
		avgLatency = total / time.Duration(len(latencies))
		p95 = latencies[len(latencies)*95/100]
	}
	latenciesMu.Unlock()

	result := TestResult{
		TargetRPS:     targetRPS,
		ActualRPS:     actualRPS,
		SuccessRate:   successRate,
		TotalRequests: finalTotal,
		Successful:    finalSuccessful,
		Failed:        finalFailed,
		AvgLatency:    avgLatency,
		P95Latency:    p95,
	}

	fmt.Printf("📈 Results:\n")
	fmt.Printf("   Target RPS: %d\n", result.TargetRPS)
	fmt.Printf("   Actual RPS: %.1f\n", result.ActualRPS)
	fmt.Printf("   Success Rate: %.1f%%\n", result.SuccessRate)
	fmt.Printf("   Total Requests: %d\n", result.TotalRequests)
	fmt.Printf("   Successful: %d\n", result.Successful)
	fmt.Printf("   Failed: %d\n", result.Failed)
	fmt.Printf("   Avg Latency: %v\n", result.AvgLatency)
	fmt.Printf("   P95 Latency: %v\n", result.P95Latency)
	fmt.Printf("\n")

	return result
}

func main() {
	var (
		baseURL  = flag.String("url", "http://127.0.0.1:8080", "Base URL")
		startRPS = flag.Int("start", 100, "Start RPS")
		endRPS   = flag.Int("end", 5000, "End RPS")
		step     = flag.Int("step", 100, "Step")
		duration = flag.Duration("duration", 20*time.Second, "Test duration")
	)
	flag.Parse()

	fmt.Printf("🚀 Basic Load Tester for Click Counter\n")
	fmt.Printf("=====================================\n")
	fmt.Printf("URL: %s\n", *baseURL)
	fmt.Printf("RPS Range: %d - %d (step %d)\n", *startRPS, *endRPS, *step)
	fmt.Printf("Duration: %v\n\n", *duration)

	// Проверяем сервис
	resp, err := http.Get(fmt.Sprintf("%s/health", *baseURL))
	if err != nil {
		log.Fatalf("Service not available: %v", err)
	}
	resp.Body.Close()
	fmt.Printf("✅ Service is available\n\n")

	var results []TestResult

	// Простое тестирование в указанном диапазоне
	fmt.Printf("=== Testing RPS range: %d - %d (step %d) ===\n", *startRPS, *endRPS, *step)
	for rps := *startRPS; rps <= *endRPS; rps += *step {
		result := runTest(*baseURL, rps, *duration)
		results = append(results, result)

		if result.SuccessRate < 90 {
			fmt.Printf("⚠️ Success rate dropped below 90%% at %d RPS, stopping tests\n", rps)
			break
		}

		// Пауза между тестами (кроме последнего)
		if rps < *endRPS {
			time.Sleep(3 * time.Second)
		}
	}

	// Итоговый отчет
	fmt.Printf("\n📊 FINAL SUMMARY\n")
	fmt.Printf("================\n")
	fmt.Printf("%-10s | %-10s | %-10s | %-12s | %-12s | %s\n",
		"Target", "Actual", "Success%", "AvgLatency", "P95Latency", "Status")
	fmt.Printf("------------------------------------------------------------------------\n")

	maxSuccessfulRPS := 0
	maxActualRPS := 0.0

	for _, result := range results {
		status := "✅"
		if result.SuccessRate < 95 {
			status = "⚠️"
		}
		if result.SuccessRate < 90 {
			status = "❌"
		}

		// Обновляем максимальный успешный RPS (отдельно от статуса)
		if result.SuccessRate >= 95 && result.TargetRPS > maxSuccessfulRPS {
			maxSuccessfulRPS = result.TargetRPS
		}

		if result.ActualRPS > maxActualRPS {
			maxActualRPS = result.ActualRPS
		}

		fmt.Printf("%-10d | %-10.1f | %-9.1f%% | %-12v | %-12v | %s\n",
			result.TargetRPS,
			result.ActualRPS,
			result.SuccessRate,
			result.AvgLatency.Truncate(time.Microsecond),
			result.P95Latency.Truncate(time.Microsecond),
			status)
	}

	fmt.Printf("\n🎯 CONCLUSIONS:\n")
	fmt.Printf("   Maximum RPS with 95%% success: %d\n", maxSuccessfulRPS)
	fmt.Printf("   Maximum achieved RPS: %.1f\n", maxActualRPS)
	fmt.Printf("   Total tests completed: %d\n", len(results))

	if maxSuccessfulRPS >= 1000 {
		fmt.Printf("   ✅ System meets ТЗ requirements (1000-5000 RPS)\n")
	} else {
		fmt.Printf("   ⚠️ System may not fully meet ТЗ requirements\n")
	}

	fmt.Printf("\n🎉 Load testing completed!\n")
}
