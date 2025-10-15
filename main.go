package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func main() {
	// Флаги коммандной строки
	urlsFile := flag.String("urls", "urls.txt", "файл со списком URL")
	workers := flag.Int("workers", 5, "количество воркеров")
	rateLimit := flag.Int("rate-limit", 10, "лимит запросов в секунду")
	flag.Parse()

	// Загрузка URl
	urls, err := loadUrls(*urlsFile)
	if err != nil {
		fmt.Printf("Ошибка при чтении файла с URL: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Запуск парсера: %d URL, %d воркеров, %d req/sec\n", len(urls), *workers, *rateLimit)

	ctx := context.Background()
	limiter := rate.NewLimiter(rate.Limit(*rateLimit), *rateLimit)

	jobs := make(chan string)
	wg := &sync.WaitGroup{}

	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go worker(ctx, i+1, jobs, wg, limiter)
	}

	go func() {
		for _, url := range urls {
			jobs <- url
		}
		close(jobs)
	}()

	wg.Wait()
	fmt.Println("Парсинг завершен")

}

func worker(ctx context.Context, id int, jobs <-chan string, wg *sync.WaitGroup, limiter *rate.Limiter) {
	defer wg.Done()
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for url := range jobs {
		if err := limiter.Wait(ctx); err != nil {
			fmt.Printf("Воркер %d: ошибка лимитера: %v\n", id, err)
			continue
		}

		fmt.Printf("Воркер %d: обрабатываю %s\n", id, url)
		start := time.Now()
		resp, err := client.Get(url)
		if err != nil {
			fmt.Printf("Воркер %d: ошибка запроса %s: %v\n", id, url, err)
			continue
		}

		resp.Body.Close()
		fmt.Printf("Воркер %d: %s (%v)\n", id, url, time.Since(start))
	}
}

// loadUrls читает файл и возвращает слайс URL
func loadUrls(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var urls []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			urls = append(urls, line)
		}
	}

	return urls, scanner.Err()
}
