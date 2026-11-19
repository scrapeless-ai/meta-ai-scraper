// Run without arguments to submit the question. Pass a task ID to retrieve it.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const api = "https://api.scrapeless.com/api/v2/scraper"
const payload = `{"actor":"scraper.meta","input":{"prompt":"Which public museums in Chicago have architecture exhibits?","country":"US"}}`

type reply struct {
	TaskID  string `json:"task_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  struct {
		Answer  *string `json:"full_response"`
		Sources []struct {
			Title string `json:"title"`
			URL   string `json:"url"`
		} `json:"references"`
	} `json:"task_result"`
}

func run() error {
	token := os.Getenv("SCRAPELESS_API_TOKEN")
	if token == "" || len(os.Args) > 2 {
		return fmt.Errorf("set SCRAPELESS_API_TOKEN and run go run examples/example.go [task_id]")
	}
	create := len(os.Args) == 1
	method, path := http.MethodPost, "/request"
	var body io.Reader = strings.NewReader(payload)
	if !create {
		method, path, body = http.MethodGet, "/result/"+url.PathEscape(os.Args[1]), nil
	}
	req, err := http.NewRequest(method, api+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-token", token)
	client := &http.Client{Timeout: time.Minute}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("HTTP %s", res.Status)
	}
	var data reply
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return err
	}
	if create {
		if data.TaskID == "" {
			return fmt.Errorf("creation response has no task ID")
		}
		fmt.Println(data.TaskID)
		return nil
	}
	switch data.Status {
	case "pending", "running":
		fmt.Println(data.Status)
	case "success":
		if data.Result.Answer == nil {
			return fmt.Errorf("response has no full_response")
		}
		fmt.Println(*data.Result.Answer)
		for _, source := range data.Result.Sources {
			fmt.Printf("- %s: %s\n", source.Title, source.URL)
		}
	default:
		return fmt.Errorf("task status %s: %s", data.Status, data.Message)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
