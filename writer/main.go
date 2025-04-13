package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"hash/crc32"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// Review структура для отзывов
type Review struct {
	Total      int64    `json:"total"`
	Downloaded int64    `json:"downloaded"`
	Rating     float64  `json:"rating"`
	Reviews    []string `json:"reviews"`
}

// Data структура входящих данных
type Data struct {
	ID         int64    `json:"id"`
	ASIN       string   `json:"asin"`
	Title      string   `json:"title"`
	Group      string   `json:"group"`
	Salesrank  int64    `json:"salesrank"`
	Similar    []string `json:"similar"`
	Categories []string `json:"categories"`
	Reviews    Review   `json:"reviews"`
	Time       string   `json:"time"`
}

// Writer структура для управления записью в файл
type Writer struct {
	ID    int
	File  *os.File
	Mu    sync.Mutex
	Input chan *Data
	Wg    sync.WaitGroup
	Done  chan struct{}
}

var (
	writers    []*Writer
	numWriters int
)

func (w *Writer) start() {
	log.Println("starting")
	defer w.Wg.Done()
	for {
		select {
		case data := <-w.Input:
			data.Time = time.Now().UTC().Format(time.RFC3339Nano)

			jsonData, err := json.Marshal(data)
			if err != nil {
				log.Printf("Writer %d: marshal error: %v", w.ID, err)
				continue
			}

			w.Mu.Lock()
			if _, err := w.File.Write(append(jsonData, '\n')); err != nil {
				log.Printf("Writer %d: write error: %v", w.ID, err)
			}
			w.Mu.Unlock()

		case <-w.Done:
			return
		}
	}
}

func initWriters(num int) {
	writers = make([]*Writer, num)
	for i := 0; i < num; i++ {
		filename := fmt.Sprintf("worker%d.txt", i+1)
		file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Failed to open file %s: %v", filename, err)
		}

		writer := &Writer{
			ID:    i + 1,
			File:  file,
			Input: make(chan *Data, 100),
			Done:  make(chan struct{}),
		}
		writer.Wg.Add(1)
		go writer.start()
		writers[i] = writer
	}
	log.Println("writers inited")
}

func selectWriter(id int64) *Writer {
	log.Println("selectWriter")
	hash := crc32.ChecksumIEEE([]byte(fmt.Sprintf("%d", id)))
	return writers[int(hash)%numWriters]
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Start handling")
	if r.Method != http.MethodPost {
		log.Println("Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data []Data
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		log.Println("Error in writer")
		return
	}

	for i, d := range data {
		d = data[i]
		writer := selectWriter(d.ID)
		writer.Input <- &d
	}

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "Data accepted by workers")
}

func main() {
	var port int
	flag.IntVar(&port, "port", 8080, "Server port")
	flag.IntVar(&numWriters, "num", 3, "Number of writers")
	flag.Parse()

	initWriters(numWriters)
	defer func() {
		for _, w := range writers {
			close(w.Done)
			close(w.Done)
			w.Wg.Wait()
			w.File.Close()
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	server := http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      mux,
	}

	fmt.Println("starting server at", server.Addr)
	server.ListenAndServe()

	// http.HandleFunc("/", handler)
	log.Printf("Server started on :%d with %d writers", port, numWriters)
	// log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
	log.Fatal(server.ListenAndServe(), nil)
}
