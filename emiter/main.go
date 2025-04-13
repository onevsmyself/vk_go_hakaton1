package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Review struct {
	Total      int64    `json:"total"`
	Downloaded int64    `json:"downloaded"`
	Rating     float64  `json:"rating"`
	Reviews    []string `json:"reviews"`
}

type Data struct {
	ID         int64    `json:"id"`
	ASIN       string   `json:"asin"`
	Title      string   `json:"title"`
	Group      string   `json:"group"`
	Salesrank  int64    `json:"salesrank"`
	Similar    []string `json:"similar"`
	Categories []string `json:"categories"`
	Reviews    Review   `json:"review"`
}

func main() {
	file_data, err := os.Open("test.txt")
	if err != nil {
		fmt.Println("Ошибка чтения из файла:", err)
		return
	}
	defer file_data.Close()

	productsChan := make(chan Data)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go MakeRequests(wg, productsChan)

	scanner := bufio.NewScanner(file_data)
	var productLines []string
	var foundFirstId bool

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "Id:") {
			if len(productLines) > 0 {
				product := parseProductBlock(productLines)
				if isValidProduct(product) {
					productsChan <- product
				}
				productLines = nil
			}
			foundFirstId = true
		}

		if !foundFirstId {
			continue
		}
		productLines = append(productLines, line)
	}

	if len(productLines) > 0 {
		product := parseProductBlock(productLines)
		if isValidProduct(product) {
			productsChan <- product
		}
	}
	close(productsChan)
	wg.Wait()
	if errScan := scanner.Err(); errScan != nil {
		fmt.Println("Ошибка при сканировании файла:", errScan)
	}

	fmt.Println("Конец обработки файла")
}

func parseProductBlock(lines []string) Data {
	var d Data
	skipDiscontinued := false

	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		switch {
		case strings.HasPrefix(line, "Id:"):
			var id int64
			_, err := fmt.Sscanf(line, "Id: %d", &id)
			if err == nil {
				d.ID = id
			}

		case strings.HasPrefix(line, "ASIN:"):
			asin := strings.TrimPrefix(line, "ASIN:")
			d.ASIN = strings.TrimSpace(asin)

		case strings.HasPrefix(line, "title:"):
			title := strings.TrimPrefix(line, "title:")
			d.Title = strings.TrimSpace(title)

		case strings.HasPrefix(line, "group:"):
			group := strings.TrimPrefix(line, "group:")
			d.Group = strings.TrimSpace(group)

		case strings.HasPrefix(line, "salesrank:"):
			var rank int64
			_, err := fmt.Sscanf(line, "salesrank: %d", &rank)
			if err == nil {
				d.Salesrank = rank
			}

		case strings.HasPrefix(line, "similar:"):
			parts := strings.Fields(line)
			if len(parts) > 2 {
				d.Similar = parts[2:]
			}

		case strings.HasPrefix(line, "discontinued product"):
			skipDiscontinued = true

		case strings.HasPrefix(line, "|"):
			cat := strings.TrimLeft(line, "|")
			d.Categories = append(d.Categories, cat)

		case strings.HasPrefix(line, "reviews:"):
			parts := strings.Fields(line)
			if len(parts) >= 8 {
				total, _ := strconv.ParseInt(parts[2], 10, 64)
				downloaded, _ := strconv.ParseInt(parts[4], 10, 64)
				ratingFloat, _ := strconv.ParseFloat(parts[7], 64)
				d.Reviews.Rating = ratingFloat
				d.Reviews.Total = total
				d.Reviews.Downloaded = downloaded
			}

		case strings.Contains(line, "cutomer:") && strings.Contains(line, "rating:"):
			d.Reviews.Reviews = append(d.Reviews.Reviews, rawLine)
		}
	}

	if skipDiscontinued {
		return Data{}
	}
	return d
}

func isValidProduct(product Data) bool {
	return !(product.ID == 0 && product.ASIN == "" && product.Title == "" && product.Group == "")
}

const chunkSize = 5

var client *http.Client

func MakeRequests(wg *sync.WaitGroup, dataChan chan Data) {
	defer wg.Done()
	writerPorts := []string{"8080/", "8080"}
	var data []Data
	client = &http.Client{}
	cntReq := 0
	for item := range dataChan {
		data = append(data, item)
		if len(data) == chunkSize {
			request(writerPorts[cntReq%len(writerPorts)], data...)
			data = nil
			cntReq++
		}
	}
	if len(data) > 0 {
		request(writerPorts[cntReq%len(writerPorts)], data...)
	}
}
func request(port string, data ...Data) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("ошибка при маршалинге JSON: %v", err)
		return
	}
	url := "http://localhost:" + port
	log.Println(url)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("ошибка при создании запроса: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("ошибка при выполнении запроса: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ошибка при чтении ответа: %v", err)
		return
	}

	if resp.StatusCode >= 400 {
		log.Printf("ошибка HTTP: %s", resp.Status)
	}
	log.Println(string(body))
	return
}
