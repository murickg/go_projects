package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
)

func main() {
	// Флаги коммандной строки
	urlsFile := flag.String("urls", "urls.txt", "файл со списком URL")
	flag.Parse()

	// Загрузка URl
	urls, err := loadUrls(*urlsFile)
	if err != nil {
		fmt.Printf("Ошибка при чтении файла с URL: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Загружено %d URL\n", len(urls))
	for _, url := range urls {
		fmt.Println(url)
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
