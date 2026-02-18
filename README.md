# Willknow Go Examples

Official examples repository for [Willknow Go](https://github.com/willknow-ai/willknow-go) - Your intelligent companion for Go applications.

This repository contains multiple standalone examples demonstrating different features and use cases of Willknow.

## Prerequisites

- Go 1.21 or later
- An AI API key (Anthropic Claude, OpenAI, DeepSeek, etc.)
- Docker (optional, for containerized deployment)

## Installation

Each example is self-contained with its own `go.mod` file. You can run them individually:

```bash
# Clone this repository
git clone https://github.com/willknow-ai/willknow-go-examples.git
cd willknow-go-examples

# Set your API key
export AI_API_KEY=your-api-key

# Run any example
cd basic
go mod download
go run main.go
```

## Examples

### 1. [basic](./basic/) - Basic Debugging Assistant

The simplest example showing Willknow as a debugging companion.

**Features:**
- Log analysis and error diagnosis
- Code reading and searching
- Request ID based troubleshooting

**Best for:** Getting started, understanding core debugging features

### 2. [agent-api](./agent-api/) - AI Agent with API Integration

Advanced example showing Willknow as an AI agent that can call your application's APIs.

**Features:**
- OpenAPI spec auto-discovery
- Natural language API calling
- Multi-turn agent conversations
- External AI system integration

**Best for:** Building AI agents, enabling external AI to control your system

## Quick Start

Each example has its own README with detailed instructions. Choose based on your needs:

```bash
# For basic debugging features
cd basic/
cat README.md

# For advanced agent capabilities
cd agent-api/
cat README.md
```

## Comparison

| Feature | basic | agent-api |
|---------|-------|-----------|
| Log analysis | ✅ | ✅ |
| Code reading | ✅ | ✅ |
| API calling | ❌ | ✅ |
| OpenAPI integration | ❌ | ✅ |
| Agent discovery | ❌ | ✅ |
| External AI access | ❌ | ✅ |
| Complexity | Simple | Medium |

## Next Steps

1. Start with the **basic** example to understand core features
2. Move to **agent-api** when you need API integration
3. Check the [Willknow Go documentation](https://github.com/willknow-ai/willknow-go) for full API reference
4. Join our community and contribute more examples!

## Development

### Local Development with Unpublished Willknow

If you're developing with a local version of Willknow Go:

```bash
# In each example's go.mod, uncomment the replace directive:
# replace github.com/willknow-ai/willknow-go => ../../willknow-go

# Or use go mod edit:
cd basic
go mod edit -replace=github.com/willknow-ai/willknow-go=../../willknow-go
```

### Contributing Examples

We welcome new examples! Please:
1. Create a new directory with a descriptive name
2. Include `go.mod`, `main.go`, `README.md`, and `Dockerfile`
3. Follow the structure of existing examples
4. Update this main README with your example
5. Submit a pull request

## License

Same as [Willknow Go](https://github.com/willknow-ai/willknow-go) - MIT License

## Links

- [Willknow Go Repository](https://github.com/willknow-ai/willknow-go)
- [Documentation](https://github.com/willknow-ai/willknow-go#readme)
- [Report Issues](https://github.com/willknow-ai/willknow-go/issues)
