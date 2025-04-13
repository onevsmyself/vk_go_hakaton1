package search

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"worker/config"
	"worker/internal/models"

	"golang.org/x/time/rate"
)

func StreamSearch(request models.SearchRequest, filePath string) (models.SearchStreamResponse, error) {
	data, err := getDataFromTxt(filePath)
	if err != nil {
		log.Printf("Error get data from file: %v", err)
		return models.SearchStreamResponse{}, err
	}

	log.Println("Data loaded from file:", len(data))

	result, err := sortDataByGroup(data, request.Group, request.Fields)
	if err != nil {
		log.Printf("Error sorting data by group: %v", err)
		return models.SearchStreamResponse{}, err
	}

	log.Println("Data sorted by group:", len(result))

	return models.SearchStreamResponse{
		Data: result,
	}, nil
}

func StatSearch(request models.SearchRequest, filePath string) (models.SearchStatResponse, error) {

	data, err := StreamSearch(request, filePath)
	if err != nil {
		log.Printf("Error in StreamSearch: %v", err)
		return models.SearchStatResponse{}, err
	}

	stats := make(map[string]int64)
	limiter := rate.NewLimiter(rate.Limit(config.Cfg.MaxRecordsPerSecond), config.Cfg.MaxRecordsPerSecond)

	maxMinute := 0

	for _, record := range data.Data {
		_ = limiter.Wait(context.Background())
		minute := record.Time[:16]
		stats[minute]++
		min := int(record.Time[14:16][0]-'0')*10 + int(record.Time[14:16][1]-'0')
		if maxMinute < min {
			maxMinute = min
		}
	}

	for minute, count := range stats {
		log.Printf("Minute: %s, Count: %d", minute, count)
		// Send to WEB
	}

	return models.SearchStatResponse{
		Query:        request.Query,
		WorkTime:     fmt.Sprintf("%d minutes", maxMinute),
		ResultsCount: int64(len(data.Data)),
		Active:       fmt.Sprintf("%d records", len(data.Data)),
	}, nil
}

func getDataFromTxt(filePath string) ([]models.Data, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data []models.Data

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		record := models.Data{}

		if err := json.Unmarshal([]byte(line), &record); err != nil {
			log.Printf("Error parsing JSON line: %v", err)
			continue
		}
		data = append(data, record)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return data, nil
}

func sortDataByGroup(data []models.Data, group string, fields []string) ([]models.Data, error) {
	limiter := rate.NewLimiter(rate.Limit(config.Cfg.MaxRecordsPerSecond), config.Cfg.MaxRecordsPerSecond)
	result := make([]models.Data, 0)

	for _, d := range data {
		if d.Group == group || group == "" {
			_ = limiter.Wait(context.Background())
			data := models.Data{}

			for _, field := range fields {
				switch field {
				case "id":
					data.ID = d.ID
				case "asin":
					data.ASIN = d.ASIN
				case "title":
					data.Title = d.Title
				case "group":
					data.Group = d.Group
				case "salesrank":
					data.SalesRank = d.SalesRank
				case "time":
					data.Time = d.Time
				}
			}

			result = append(result, data)
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no data found for group: %s", group)
	}

	return result, nil
}
