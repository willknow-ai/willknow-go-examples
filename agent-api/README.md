# Agent API Example - Task Management System

This example demonstrates Willknow's **Agent Mode** - where Willknow becomes an AI agent that can call your application's APIs through natural language.

## What This Demo Shows

Instead of manually calling REST APIs, external AI systems (or users) can simply talk to Willknow in natural language:

**Traditional Approach:**
```bash
curl -X POST http://localhost:8080/api/tasks \
  -H "Content-Type: application/json" \
  -d '{"title": "Deploy to production", "priority": "high"}'
```

**With Willknow Agent:**
```
User: "Create a new high-priority task: Deploy to production"
AI: *automatically calls the API and returns result*
```

## Architecture

```
┌─────────────────┐
│  External AI    │  (e.g., another Claude instance)
│  or User        │
└────────┬────────┘
         │ Natural Language
         ↓
┌─────────────────────┐
│  Willknow Agent     │  Port 8888
│  - /willknow/info   │  (discovery)
│  - /willknow/chat   │  (chat interface)
│  - Web UI           │  (human interface)
└─────────┬───────────┘
          │ API Calls
          ↓
┌─────────────────────┐
│  Task API           │  Port 8080
│  - GET /api/tasks   │
│  - POST /api/tasks  │
│  - PUT /api/tasks/  │
│  - DELETE /api/...  │
└─────────────────────┘
```

## Features Demonstrated

- ✅ **OpenAPI Auto-Discovery**: Automatically loads API endpoints from `openapi.yaml`
- ✅ **Natural Language API Calling**: AI translates requests to API calls
- ✅ **Agent Discovery**: `/willknow/info` endpoint for external systems
- ✅ **Multi-Turn Conversations**: Stateful chat via `/willknow/chat`
- ✅ **Full CRUD Operations**: Create, Read, Update, Delete tasks
- ✅ **Query Parameters**: Filter tasks by status
- ✅ **Path Parameters**: Access tasks by ID
- ✅ **Request Body**: Submit complex JSON payloads

## Quick Start

### Option 1: Using Docker (Recommended)

```bash
# 1. Navigate to agent-api directory
cd agent-api/

# 2. Build the image
docker build -t willknow-agent-demo .

# 3. Run with your API key
docker run -p 8080:8080 -p 8888:8888 \
  -e AI_API_KEY=your-api-key \
  willknow-agent-demo
```

### Option 2: Local Development

```bash
# 1. Set environment variables
export AI_API_KEY=your-api-key

# 2. Create log directory
sudo mkdir -p /var/log
sudo chmod 777 /var/log

# 3. Navigate to example directory and run
cd agent-api/
go mod download
go run main.go
```

## Usage Examples

### Access the Web UI

Open http://localhost:8888 in your browser and try these natural language commands:

#### List Tasks
```
"Show me all tasks"
"List all pending tasks"
"What are the high-priority tasks?"
```

#### Create Tasks
```
"Create a new task: Write documentation"
"Add a high-priority task called 'Fix critical bug'"
```

#### Update Tasks
```
"Mark task-2 as completed"
"Change task-3 priority to high"
"Update task-1 description to 'Deploy v2.0'"
```

#### Delete Tasks
```
"Delete task-1"
"Remove the task with ID task-3"
```

### Use Agent Discovery (for External AI)

External AI systems can discover Willknow's capabilities:

```bash
curl http://localhost:8888/willknow/info
```

Response:
```json
{
  "name": "Task Management API",
  "description": "A simple task management system...",
  "chatEndpoint": "/willknow/chat",
  "auth": {
    "required": false
  },
  "capabilities": [
    {
      "name": "listTasks",
      "description": "Returns a list of all tasks in the system"
    },
    {
      "name": "createTask",
      "description": "Creates a new task with the provided information"
    },
    ...
  ]
}
```

### Use Agent Chat API (for External AI)

External AI systems can have conversations with Willknow:

```bash
# Start a conversation
curl -X POST http://localhost:8888/willknow/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Create a task: Review pull requests"
  }'

# Response includes sessionID for multi-turn conversation
{
  "response": "I've created a new task...",
  "sessionId": "abc123...",
  "toolCalls": [...]
}

# Continue the conversation
curl -X POST http://localhost:8888/willknow/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Now mark it as high priority",
    "sessionId": "abc123..."
  }'
```

## API Endpoints

### Task API (Port 8080)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/tasks` | List all tasks (optional `?status=` filter) |
| POST | `/api/tasks` | Create a new task |
| GET | `/api/tasks/{id}` | Get a specific task |
| PUT | `/api/tasks/{id}` | Update a task |
| DELETE | `/api/tasks/{id}` | Delete a task |
| POST | `/api/tasks/{id}/complete` | Mark task as completed |

### Willknow Agent (Port 8888)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/` | Web UI for humans |
| GET | `/willknow/info` | Agent discovery (public) |
| POST | `/willknow/chat` | Agent chat API (for external AI) |

## Configuration

The example configures Willknow in Agent Mode:

```go
assistant, err := aiassistant.New(aiassistant.Config{
    SourcePath:  "/app/source",
    LogFiles:    []string{"/var/log/app.log"},
    Port:        8888,

    // Agent Mode: Load OpenAPI spec
    APISpec:     "/path/to/openapi.yaml",
    HostBaseURL: "http://localhost:8080",

    // No authentication for demo
    Auth: aiassistant.AuthConfig{
        GetUser: aiassistant.NoAuth,
    },
})
```

## Understanding Agent Mode

When `APISpec` is configured:

1. **At Startup**: Willknow parses the OpenAPI spec and loads all API endpoints as tools
2. **Agent Discovery**: The `/willknow/info` endpoint exposes capabilities to external systems
3. **Natural Language Processing**: When users/AIs send messages, Willknow:
   - Understands the intent
   - Maps intent to API tools
   - Calls the appropriate endpoints
   - Returns natural language responses
4. **Multi-Turn Sessions**: Conversations maintain context across multiple exchanges

## Comparison with Basic Example

| Feature | Basic Example | Agent API Example |
|---------|---------------|-------------------|
| Primary Use | Debugging companion | AI agent |
| OpenAPI Integration | ❌ | ✅ |
| API Calling | ❌ | ✅ |
| External AI Access | ❌ | ✅ |
| Agent Discovery | ❌ | ✅ |
| Log Analysis | ✅ | ✅ |
| Code Reading | ✅ | ✅ |

## Extending This Example

### Add Authentication

Replace `NoAuth` with custom authentication:

```go
Auth: aiassistant.AuthConfig{
    GetUser: func(r *http.Request) (*aiassistant.User, error) {
        // Your auth logic here
        token := r.Header.Get("Authorization")
        return validateToken(token)
    },
},
```

### Add More Endpoints

1. Update `openapi.yaml` with new endpoints
2. Implement handlers in `main.go`
3. Restart - Willknow automatically discovers new capabilities

### Connect External AI

Any AI system can call Willknow's agent interface:

```python
# Python example: External AI calling Willknow
import requests

response = requests.post(
    "http://localhost:8888/willknow/chat",
    json={
        "message": "Create a task: Deploy v2.0",
        "sessionId": session_id  # for multi-turn
    },
    headers={"Authorization": f"Bearer {user_token}"}
)

result = response.json()
print(result["response"])
```

## Troubleshooting

### "AI_API_KEY not set"
**Solution**: Set the environment variable:
```bash
export AI_API_KEY=your-claude-api-key
```

### "Failed to parse OpenAPI spec"
**Solution**: Verify `openapi.yaml` is valid OpenAPI 3.0:
```bash
# Use a validator
npx @apidevtools/swagger-cli validate openapi.yaml
```

### "API call failed: connection refused"
**Solution**: Ensure the task API is running on port 8080. Check if the port is available:
```bash
lsof -i :8080
```

### Port already in use
**Solution**: Change ports in both `main.go` and `openapi.yaml`, or kill the process using the port

## Learn More

- See [Basic Example](../basic/) for debugging features
- Read [Main README](../../README.md) for full API documentation
- Check [OpenAPI Specification](https://swagger.io/specification/) for spec format

## Next Steps

1. Try the example and observe how natural language maps to API calls
2. Modify `openapi.yaml` to add your own endpoints
3. Integrate Willknow into your own application
4. Connect external AI systems to your Willknow-enabled app
