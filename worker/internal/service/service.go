package service

import (
	"log"
	"sync"
	"worker/internal/models"
	"worker/internal/search"
)

type Query struct {
	taskID uint32
	query  string
}


package search

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/your_module/models"
	"github.com/your_module/statsearch"
	"github.com/your_module/streamsearch"
)

func Search(requests <-chan models.SearchRequest, response chan<- any, filePath string) {
	const maxWorkers = 5
	var wg sync.WaitGroup

	tasks := make(chan models.SearchRequest)

	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for req := range tasks {
				switch req.TypeSearch {
				case "stream":
					// вызвать декод
					// распарсить в структуру models.SearchRequest
					// отправить в StreamSearch
					// в StreamSearch чекать состояние задачи (вдруг надо завершить)

					var decoded models.SearchRequest
					if err := json.Unmarshal([]byte(req.Query), &decoded); err != nil {
						log.Printf("Error decoding request: %v", err)
						continue
					}

					res, err := streamsearch.StreamSearch(decoded, filePath)
					if err != nil {
						log.Printf("Error in StreamSearch: %v", err)
					}
					log.Println("StreamSearch response:", res)
					response <- res

				case "stat":
					res, err := statsearch.StatSearch(req, filePath)
					if err != nil {
						log.Printf("Error in StatSearch: %v", err)
					}
					log.Println("StatSearch response:", res)
					response <- res

				default:
					log.Printf("Unknown search type: %s", req.TypeSearch)
				}
			}
		}()
	}

	for request := range requests {
		tasks <- request
	}

	close(tasks)
	wg.Wait()
}
