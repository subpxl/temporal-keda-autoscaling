// goapi/cmd/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.temporal.io/sdk/client"
)

var (
	TemporalAddress string = "temporal-dev:7233"
	TaskQueue       string = "default-activity-queue"
	WorkflowName    string = "FourActivityWorkflow"
)

// API request/response
type TaskRequest struct {
	Name string `json:"name"`
}
type TaskResponse struct {
	WorkflowID string `json:"workflow_id"`
	RunID      string `json:"run_id"`
}

func main() {

	// Connect to Temporal server
	c, err := client.Dial(client.Options{HostPort: TemporalAddress})
	if err != nil {
		log.Fatalln("Unable to create Temporal client", err)
	}
	defer c.Close()

	// Setup HTTP server
	r := mux.NewRouter()

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Create task → starts workflow
	r.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		var req TaskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		workflowOptions := client.StartWorkflowOptions{
			ID:        fmt.Sprintf("%s-%d", WorkflowName, time.Now().Unix()),
			TaskQueue: TaskQueue,
		}

		we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, WorkflowName, req.Name)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := TaskResponse{WorkflowID: we.GetID(), RunID: we.GetRunID()}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}).Methods("POST")

	log.Println("API running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
