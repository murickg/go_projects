package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"sync"
	"syscall"
	"time"

	"golang.org/x/time/rate"
)

type Result struct {
	URL   string
	Title string
	Err   error
}

func main() {
	// Флаги коммандной строки
	urlsFile := flag.String("urls", "urls.txt", "файл со списком URL")
	workers := flag.Int("workers", 5, "количество воркеров")
	rateLimit := flag.Int("rate-limit", 10, "лимит запросов в секунду")
	timeout := flag.Int("timeout", 5, "таймаут для HTTP запросов в секундах")
	output := flag.String("output", "results.csv", "файл для сохранения результатов")
	flag.Parse()

	// Загрузка URl
	urls, err := loadUrls(*urlsFile)
	if err != nil {
		fmt.Printf("Ошибка при чтении файла с URL: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Запуск парсера: %d URL, %d воркеров, %d req/sec\n", len(urls), *workers, *rateLimit)

	ctx, cancel := context.WithCancel(context.Background())
	go handleSignals(cancel)

	limiter := rate.NewLimiter(rate.Limit(*rateLimit), *rateLimit)

	jobs := make(chan string)
	results := make(chan Result)
	var wg sync.WaitGroup

	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go worker(ctx, i+1, jobs, results, limiter, time.Duration(*timeout)*time.Second, &wg)
	}

	go func() {
		for _, url := range urls {
			jobs <- url
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	saveResults(*output, results)
	fmt.Println("Парсинг завершен")
	fmt.Printf("Результаты сохранены в %s\n", *output)

}

func worker(ctx context.Context, id int, jobs <-chan string, results chan<- Result, limiter *rate.Limiter, timeout time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()
	client := &http.Client{
		Timeout: timeout,
	}

	for url := range jobs {
		if err := limiter.Wait(ctx); err != nil {
			fmt.Printf("Воркер %d: ошибка лимитера: %v\n", id, err)
			results <- Result{URL: url, Title: "", Err: err}
			continue
		}

		reqContext, cancel := context.WithTimeout(ctx, timeout)
		title, err := fetchTitle(reqContext, client, url)
		cancel()

		results <- Result{URL: url, Title: title, Err: err}
	}
}

func fetchTitle(ctx context.Context, client *http.Client, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP статус %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1_000_000))
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile("(?i)<title>(.*?)</title>")
	matches := re.FindSubmatch(body)
	if len(matches) < 2 {
		return "", fmt.Errorf("тег <title> не найден")
	}

	return string(matches[1]), nil
}

func saveResults(path string, results <-chan Result) {
	file, err := os.Create(path)
	if err != nil {
		fmt.Printf("Ошибка при создании файла: %v\n", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"URL", "Title", "Error"})

	for res := range results {
		errMsg := ""
		if res.Err != nil {
			errMsg = res.Err.Error()
		}
		writer.Write([]string{res.URL, res.Title, errMsg})
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

func handleSignals(cancel context.CancelFunc) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	fmt.Println("\nПолучен сигнал прерывания — завершаем работу...")
	cancel()
}
