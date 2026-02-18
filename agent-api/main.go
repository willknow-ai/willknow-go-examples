package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	aiassistant "github.com/willknow-ai/willknow-go"
)

// Task represents a task in the system
type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // pending, in_progress, completed
	Priority    string    `json:"priority"` // low, medium, high
	CreatedAt   time.Time `json:"createdAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
}

// TaskStore manages tasks in memory
type TaskStore struct {
	mu    sync.RWMutex
	tasks map[string]*Task
}

func NewTaskStore() *TaskStore {
	return &TaskStore{
		tasks: make(map[string]*Task),
	}
}

func (s *TaskStore) Create(task *Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[task.ID] = task
}

func (s *TaskStore) Get(id string) (*Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, ok := s.tasks[id]
	return task, ok
}

func (s *TaskStore) List(statusFilter string) []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Task
	for _, task := range s.tasks {
		if statusFilter == "" || task.Status == statusFilter {
			result = append(result, task)
		}
	}
	return result
}

func (s *TaskStore) Update(id string, updates map[string]interface{}) (*Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return nil, false
	}

	if title, ok := updates["title"].(string); ok {
		task.Title = title
	}
	if desc, ok := updates["description"].(string); ok {
		task.Description = desc
	}
	if status, ok := updates["status"].(string); ok {
		task.Status = status
	}
	if priority, ok := updates["priority"].(string); ok {
		task.Priority = priority
	}

	return task, true
}

func (s *TaskStore) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[id]; !ok {
		return false
	}
	delete(s.tasks, id)
	return true
}

// Logger writes logs to file
type Logger struct {
	file *os.File
}

func NewLogger(path string) (*Logger, error) {
	os.MkdirAll("/var/log", 0755)

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &Logger{file: file}, nil
}

func (l *Logger) Log(requestID, level, message string) {
	timestamp := time.Now().Format(time.RFC3339)
	logLine := fmt.Sprintf("[%s] [%s] [RequestID: %s] %s\n", timestamp, level, requestID, message)
	l.file.WriteString(logLine)
	fmt.Print(logLine)
}

func (l *Logger) Close() {
	l.file.Close()
}

var (
	logger *Logger
	store  *TaskStore
)

func main() {
	// Initialize logger
	var err error
	logger, err = NewLogger("/var/log/app.log")
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	logger.Log("STARTUP", "INFO", "Task Management System starting...")

	// Initialize task store
	store = NewTaskStore()

	// Add some sample tasks
	sampleTasks := []*Task{
		{
			ID:          "task-1",
			Title:       "Setup development environment",
			Description: "Install Go, Docker, and configure IDE",
			Status:      "completed",
			Priority:    "high",
			CreatedAt:   time.Now().Add(-48 * time.Hour),
			CompletedAt: timePtr(time.Now().Add(-24 * time.Hour)),
		},
		{
			ID:          "task-2",
			Title:       "Implement authentication",
			Description: "Add JWT-based user authentication",
			Status:      "in_progress",
			Priority:    "high",
			CreatedAt:   time.Now().Add(-24 * time.Hour),
		},
		{
			ID:          "task-3",
			Title:       "Write unit tests",
			Description: "Achieve 80% code coverage",
			Status:      "pending",
			Priority:    "medium",
			CreatedAt:   time.Now().Add(-12 * time.Hour),
		},
	}

	for _, task := range sampleTasks {
		store.Create(task)
	}

	logger.Log("STARTUP", "INFO", fmt.Sprintf("Loaded %d sample tasks", len(sampleTasks)))

	// Start AI Assistant with APISpec
	go func() {
		apiKey := os.Getenv("AI_API_KEY")
		if apiKey == "" {
			log.Println("[WARNING] AI_API_KEY not set. AI Assistant will not work properly.")
			log.Println("[INFO] Set it with: export AI_API_KEY=your-key")
		}

		provider := os.Getenv("AI_PROVIDER")
		if provider == "" {
			provider = "anthropic"
		}

		assistant, err := aiassistant.New(aiassistant.Config{
			SourcePath:  "/app/source",
			LogFiles:    []string{"/var/log/app.log"},
			Port:        8888,
			Provider:    provider,
			APIKey:      apiKey,
			APISpec:     "/app/source/openapi.yaml",
			HostBaseURL: "http://localhost:8080",
			Auth: aiassistant.AuthConfig{
				GetUser: aiassistant.NoAuth, // Open access for demo
			},
			EnableCodeIndex: false, // Disable for faster startup
		})

		if err != nil {
			log.Printf("[ERROR] Failed to initialize AI Assistant: %v", err)
			return
		}

		if err := assistant.Start(); err != nil {
			log.Printf("[ERROR] AI Assistant failed: %v", err)
		}
	}()

	// Setup HTTP handlers
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/api/tasks", handleTasks)
	http.HandleFunc("/api/tasks/", handleTaskByID)

	logger.Log("STARTUP", "INFO", "Server starting on :8080")
	log.Println(strings.Repeat("=", 70))
	log.Println("🚀 Task Management System is running!")
	log.Println("  - Main API: http://localhost:8080")
	log.Println("  - AI Agent UI: http://localhost:8888")
	log.Println("  - Agent Info: http://localhost:8888/willknow/info")
	log.Println("  - Agent Chat: POST http://localhost:8888/willknow/chat")
	log.Println(strings.Repeat("=", 70))

	if err := http.ListenAndServe(":8080", nil); err != nil {
		logger.Log("FATAL", "ERROR", fmt.Sprintf("Server failed: %v", err))
		log.Fatal(err)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	requestID := generateRequestID()
	logger.Log(requestID, "INFO", "Home page accessed")

	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Task Management System</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
            max-width: 1000px;
            margin: 50px auto;
            padding: 20px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        }
        .card {
            background: white;
            padding: 30px;
            border-radius: 15px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
            margin-bottom: 20px;
        }
        h1 { color: #333; margin-top: 0; }
        h2 { color: #667eea; border-bottom: 2px solid #667eea; padding-bottom: 10px; }
        .endpoint {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 8px;
            margin: 10px 0;
            border-left: 4px solid #667eea;
        }
        .method {
            display: inline-block;
            padding: 4px 10px;
            border-radius: 4px;
            font-weight: bold;
            font-size: 12px;
            margin-right: 10px;
        }
        .get { background: #28a745; color: white; }
        .post { background: #007bff; color: white; }
        .put { background: #ffc107; color: #333; }
        .delete { background: #dc3545; color: white; }
        a {
            color: #667eea;
            text-decoration: none;
        }
        a:hover { text-decoration: underline; }
        .ai-section {
            background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
            color: white;
            padding: 20px;
            border-radius: 10px;
            margin-top: 20px;
        }
        .ai-section h2 { color: white; border-bottom-color: white; }
        .ai-section a { color: white; font-weight: bold; }
        code {
            background: #f0f0f0;
            padding: 2px 8px;
            border-radius: 4px;
            font-family: 'Courier New', monospace;
        }
    </style>
</head>
<body>
    <div class="card">
        <h1>📋 Task Management System</h1>
        <p>A demo application showing Willknow as an AI Agent that can manage tasks via natural language.</p>

        <h2>Available API Endpoints</h2>

        <div class="endpoint">
            <span class="method get">GET</span>
            <a href="/api/tasks">/api/tasks</a>
            <p>List all tasks (supports ?status=pending|in_progress|completed)</p>
        </div>

        <div class="endpoint">
            <span class="method post">POST</span>
            <code>/api/tasks</code>
            <p>Create a new task</p>
        </div>

        <div class="endpoint">
            <span class="method get">GET</span>
            <code>/api/tasks/{taskId}</code>
            <p>Get a specific task by ID</p>
        </div>

        <div class="endpoint">
            <span class="method put">PUT</span>
            <code>/api/tasks/{taskId}</code>
            <p>Update a task</p>
        </div>

        <div class="endpoint">
            <span class="method delete">DELETE</span>
            <code>/api/tasks/{taskId}</code>
            <p>Delete a task</p>
        </div>

        <div class="endpoint">
            <span class="method post">POST</span>
            <code>/api/tasks/{taskId}/complete</code>
            <p>Mark a task as completed</p>
        </div>
    </div>

    <div class="card ai-section">
        <h2>🤖 AI Agent Interface</h2>
        <p>Instead of calling these APIs manually, you can talk to the AI agent in natural language!</p>

        <h3>Try these commands:</h3>
        <ul>
            <li>"Show me all pending tasks"</li>
            <li>"Create a new task: Deploy to production"</li>
            <li>"Mark task-2 as completed"</li>
            <li>"Update task-3 priority to high"</li>
            <li>"Delete task task-1"</li>
        </ul>

        <p><strong>Access the AI:</strong></p>
        <ul>
            <li><a href="http://localhost:8888" target="_blank">Web UI: http://localhost:8888</a></li>
            <li>Agent Info: <a href="http://localhost:8888/willknow/info" target="_blank">GET /willknow/info</a></li>
            <li>Agent Chat: POST /willknow/chat (for external AI systems)</li>
        </ul>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func handleTasks(w http.ResponseWriter, r *http.Request) {
	requestID := generateRequestID()

	switch r.Method {
	case http.MethodGet:
		status := r.URL.Query().Get("status")
		logger.Log(requestID, "INFO", fmt.Sprintf("Listing tasks (status filter: %s)", status))

		tasks := store.List(status)

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-ID", requestID)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"tasks": tasks,
		})

		logger.Log(requestID, "INFO", fmt.Sprintf("Returned %d tasks", len(tasks)))

	case http.MethodPost:
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Log(requestID, "ERROR", fmt.Sprintf("Failed to parse request: %v", err))
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		title, _ := req["title"].(string)
		if title == "" {
			logger.Log(requestID, "ERROR", "Title is required")
			http.Error(w, "Title is required", http.StatusBadRequest)
			return
		}

		task := &Task{
			ID:          fmt.Sprintf("task-%d", time.Now().Unix()),
			Title:       title,
			Description: getStringOrDefault(req, "description", ""),
			Status:      "pending",
			Priority:    getStringOrDefault(req, "priority", "medium"),
			CreatedAt:   time.Now(),
		}

		store.Create(task)
		logger.Log(requestID, "INFO", fmt.Sprintf("Created task: %s (ID: %s)", task.Title, task.ID))

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-ID", requestID)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleTaskByID(w http.ResponseWriter, r *http.Request) {
	requestID := generateRequestID()

	// Extract task ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	parts := strings.Split(path, "/")
	taskID := parts[0]

	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	// Check for /complete endpoint
	if len(parts) > 1 && parts[1] == "complete" && r.Method == http.MethodPost {
		handleCompleteTask(w, r, taskID, requestID)
		return
	}

	switch r.Method {
	case http.MethodGet:
		logger.Log(requestID, "INFO", fmt.Sprintf("Getting task: %s", taskID))

		task, ok := store.Get(taskID)
		if !ok {
			logger.Log(requestID, "WARN", fmt.Sprintf("Task not found: %s", taskID))
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-ID", requestID)
		json.NewEncoder(w).Encode(task)

	case http.MethodPut:
		var updates map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			logger.Log(requestID, "ERROR", fmt.Sprintf("Failed to parse request: %v", err))
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		logger.Log(requestID, "INFO", fmt.Sprintf("Updating task: %s", taskID))

		task, ok := store.Update(taskID, updates)
		if !ok {
			logger.Log(requestID, "WARN", fmt.Sprintf("Task not found: %s", taskID))
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		logger.Log(requestID, "INFO", fmt.Sprintf("Task updated: %s", taskID))

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-ID", requestID)
		json.NewEncoder(w).Encode(task)

	case http.MethodDelete:
		logger.Log(requestID, "INFO", fmt.Sprintf("Deleting task: %s", taskID))

		if !store.Delete(taskID) {
			logger.Log(requestID, "WARN", fmt.Sprintf("Task not found: %s", taskID))
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		logger.Log(requestID, "INFO", fmt.Sprintf("Task deleted: %s", taskID))
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleCompleteTask(w http.ResponseWriter, r *http.Request, taskID, requestID string) {
	logger.Log(requestID, "INFO", fmt.Sprintf("Marking task as completed: %s", taskID))

	now := time.Now()
	updates := map[string]interface{}{
		"status": "completed",
	}

	task, ok := store.Update(taskID, updates)
	if !ok {
		logger.Log(requestID, "WARN", fmt.Sprintf("Task not found: %s", taskID))
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	task.CompletedAt = &now
	logger.Log(requestID, "INFO", fmt.Sprintf("Task completed: %s", taskID))

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	json.NewEncoder(w).Encode(task)
}

func generateRequestID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func getStringOrDefault(m map[string]interface{}, key, defaultVal string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return defaultVal
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
