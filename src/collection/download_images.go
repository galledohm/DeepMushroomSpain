package main

import (
	"encoding/csv"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	imagesRootDir  = "../../data/images/"
	inputCSVPath   = "../../data/raw/inaturalist/observations-spain-2000-2026.csv"
	workerCount    = 6
	reportRate     = 100
	maxAttempts    = 3
	requestTimeout = 45 * time.Second

	headerID             = "id"
	headerImageURL       = "image_url"
	headerScientificName = "scientific_name"
)

type observation struct {
	ID             string
	ImageURL       string
	ScientificName string
}

type csvColumns struct {
	id             int
	imageURL       int
	scientificName int
}

type downloadResult struct {
	ObservationID string
	Err           error
}

func main() {
	observations, err := loadObservations(inputCSVPath)
	if err != nil {
		log.Fatal(err)
	}

	if len(observations) == 0 {
		log.Println("no downloadable observations found")
		return
	}

	if err := os.MkdirAll(imagesRootDir, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	client := &http.Client{
		Timeout: requestTimeout,
		Transport: &http.Transport{
			Proxy:               http.ProxyFromEnvironment,
			MaxIdleConns:        workerCount * 2,
			MaxIdleConnsPerHost: workerCount,
			MaxConnsPerHost:     workerCount,
		},
	}

	jobs := make(chan observation)
	results := make(chan downloadResult)

	var workers sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for observation := range jobs {
				results <- downloadResult{
					ObservationID: observation.ID,
					Err:           downloadObservation(client, observation),
				}
			}
		}()
	}

	go func() {
		for _, observation := range observations {
			jobs <- observation
		}
		close(jobs)
		workers.Wait()
		close(results)
	}()

	var completed int
	var failed int

	for result := range results {
		completed++
		if result.Err != nil {
			failed++
			log.Printf("download failed for observation %s: %v", result.ObservationID, result.Err)
		}

		if completed%reportRate == 0 {
			log.Printf("processed %d/%d observations", completed, len(observations))
		}
	}

	log.Printf("finished: %d downloaded, %d failed", completed-failed, failed)
}

func loadObservations(csvPath string) ([]observation, error) {
	csvFile, err := os.Open(csvPath)
	if err != nil {
		return nil, err
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	reader.FieldsPerRecord = -1

	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read csv header: %w", err)
	}

	columns, err := findCSVColumns(headers)
	if err != nil {
		return nil, err
	}

	observations := make([]observation, 0, 1024)
	seen := make(map[string]struct{})

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read csv row: %w", err)
		}

		observation := observation{
			ID:             getField(record, columns.id),
			ImageURL:       getField(record, columns.imageURL),
			ScientificName: getField(record, columns.scientificName),
		}

		if shouldSkipObservation(observation) {
			continue
		}

		key := observation.ScientificName + "|" + observation.ID
		if _, exists := seen[key]; exists {
			continue
		}

		seen[key] = struct{}{}
		observations = append(observations, observation)
	}

	return observations, nil
}

func findCSVColumns(headers []string) (csvColumns, error) {
	lookup := make(map[string]int, len(headers))
	for index, header := range headers {
		lookup[strings.TrimSpace(header)] = index
	}

	columns := csvColumns{
		id:             lookup[headerID],
		imageURL:       lookup[headerImageURL],
		scientificName: lookup[headerScientificName],
	}

	if _, ok := lookup[headerID]; !ok {
		return csvColumns{}, fmt.Errorf("missing required csv column %q", headerID)
	}
	if _, ok := lookup[headerImageURL]; !ok {
		return csvColumns{}, fmt.Errorf("missing required csv column %q", headerImageURL)
	}
	if _, ok := lookup[headerScientificName]; !ok {
		return csvColumns{}, fmt.Errorf("missing required csv column %q", headerScientificName)
	}

	return columns, nil
}

func shouldSkipObservation(observation observation) bool {
	if observation.ID == "" || observation.ImageURL == "" {
		return true
	}

	scientificName := strings.ToLower(observation.ScientificName)

	return strings.Contains(scientificName, "lichen")
}

func downloadObservation(client *http.Client, observation observation) error {
	speciesDir := filepath.Join(imagesRootDir, sanitizePathComponent(observation.ScientificName))
	if err := os.MkdirAll(speciesDir, os.ModePerm); err != nil {
		return fmt.Errorf("create species directory: %w", err)
	}

	fileExtension := detectFileExtension(observation.ImageURL)
	finalPath := filepath.Join(speciesDir, observation.ID+fileExtension)
	temporaryPath := finalPath + ".part"

	if _, err := os.Stat(finalPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("check existing file: %w", err)
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := downloadToFile(client, observation.ImageURL, temporaryPath); err != nil {
			lastErr = err
		} else if err := validateImageFile(temporaryPath); err != nil {
			lastErr = err
		} else if err := os.Rename(temporaryPath, finalPath); err != nil {
			lastErr = fmt.Errorf("rename temporary file: %w", err)
		} else {
			return nil
		}

		_ = os.Remove(temporaryPath)
		if attempt < maxAttempts {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}

	return fmt.Errorf("download failed after %d attempts: %w", maxAttempts, lastErr)
}

func downloadToFile(client *http.Client, url string, destination string) error {
	response, err := client.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status %s", response.Status)
	}

	file, err := os.Create(destination)
	if err != nil {
		return err
	}

	written, copyErr := io.Copy(file, response.Body)
	closeErr := file.Close()

	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	if response.ContentLength > 0 && written != response.ContentLength {
		return fmt.Errorf("incomplete download: wrote %d bytes, expected %d", written, response.ContentLength)
	}

	return nil
}

func validateImageFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, _, err = image.Decode(file)
	return err
}

func detectFileExtension(url string) string {
	cleanURL := strings.Split(url, "?")[0]
	extension := strings.ToLower(filepath.Ext(cleanURL))
	if extension == "" {
		return ".jpg"
	}
	return extension
}

func sanitizePathComponent(value string) string {
	replacer := strings.NewReplacer(
		"<", "_",
		">", "_",
		":", "_",
		"\"", "_",
		"/", "_",
		"\\", "_",
		"|", "_",
		"?", "_",
		"*", "_",
	)

	cleanValue := strings.TrimSpace(replacer.Replace(value))
	cleanValue = strings.Trim(cleanValue, ".")
	if cleanValue == "" {
		return "unknown_species"
	}

	return cleanValue
}

func getField(record []string, index int) string {
	if index < 0 || index >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[index])
}
