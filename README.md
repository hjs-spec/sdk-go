# JEP Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/hjs-protocol/sdk-go)](https://pkg.go.dev/github.com/hjs-protocol/sdk-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/hjs-protocol/sdk-go)](https://golang.org)

Go SDK for [JEP: A Judgment Event Protocol](https://github.com/hjs-protocol/spec).

Implements all 4 core primitives: **Judgment**, **Delegation**, **Termination**, **Verification**.

## 📦 Installation

```bash
go get github.com/jep-protocol/sdk-go
```

## 🚀 Quick Start

```go
package main

import (
    "fmt"
    "log"
    "github.com/jep-protocol/sdk-go"
)

func main() {
    // Create client with API key
    client := jep.NewClient("your-api-key")

    // 1. Record a judgment
    judgment, err := client.Judgment(&hjs.JudgmentRequest{
        Entity: "alice@bank.com",
        Action: "loan_approved",
        Scope: map[string]interface{}{
            "amount": 100000,
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✅ Judgment recorded: %s\n", judgment.ID)

    // 2. Create a delegation
    delegation, err := client.Delegation(&jep.DelegationRequest{
        Delegator: "manager@company.com",
        Delegatee: "employee@company.com",
        Scope: map[string]interface{}{
            "permissions": []string{"approve_under_1000"},
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✅ Delegation created: %s\n", delegation.ID)

    // 3. Verify the record
    verification, err := client.Verify(delegation.ID)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✅ Verification result: %s\n", verification.Status) // "VALID" or "INVALID"
}
```

## 📚 API Reference

### Creating a Client

```go
// Create client with default settings (https://api.jep.sh)
client := jep.NewClient("your-api-key")

// Create client with custom base URL
client := jep.NewClientWithURL("https://your-jep-instance.com", "your-api-key")
```

### Core Primitives

#### 1. Judgment — Record structured decisions

```go
resp, err := client.Judgment(&jep.JudgmentRequest{
    Entity: "user@example.com",           // Required: who is making the judgment
    Action: "approve",                     // Required: what action
    Scope: map[string]interface{}{         // Optional: additional context
        "amount": 1000,
        "currency": "USD",
    },
    Immutability: map[string]interface{}{  // Optional: anchor to blockchain
        "type": "ots",
    },
})
```

**Response:**
```go
type JudgmentResponse struct {
    ID        string    `json:"id"`
    Status    string    `json:"status"`
    Protocol  string    `json:"protocol"`
    Timestamp time.Time `json:"timestamp"`
}
```

#### 2. Delegation — Transfer authority

```go
resp, err := client.Delegation(&jep.DelegationRequest{
    Delegator:  "manager@company.com",     // Required: who delegates
    Delegatee:  "employee@company.com",    // Required: who receives
    JudgmentID: "jgd_xxx",                  // Optional: linked judgment
    Scope: map[string]interface{}{          // Optional: delegation scope
        "permissions": []string{"read", "write"},
    },
    Expiry: "2026-12-31T23:59:59Z",        // Optional: expiration time
})
```

#### 3. Termination — End responsibility

```go
resp, err := client.Termination(&hjs.TerminationRequest{
    Terminator: "admin@company.com",       // Required: who terminates
    TargetID:   "dlg_1234567890abcd",      // Required: what to terminate
    TargetType: "delegation",               // Required: "judgment" or "delegation"
    Reason:     "Employee left company",    // Optional: reason
})
```

#### 4. Verification — Validate records

```go
// Detailed verification
resp, err := client.Verification(&jep.VerificationRequest{
    Verifier:   "auditor@company.com",
    TargetID:   "dlg_1234567890abcd",
    TargetType: "delegation",  // "judgment", "delegation", or "termination"
})

// Quick verify (auto-detects type from ID prefix)
resp, err := client.Verify("dlg_1234567890abcd")
```

### Query Methods

```go
// Get by ID
judgment, err := client.GetJudgment("jgd_xxx")
delegation, err := client.GetDelegation("dlg_xxx")
termination, err := client.GetTermination("trm_xxx")

// List with filters
judgments, err := client.ListJudgments(&jep.ListJudgmentsParams{
    Entity: "user@example.com",
    Page:   1,
    Limit:  20,
})
```

### Utility Methods

```go
// Health check
health, err := client.Health()

// API documentation
docs, err := client.Docs()

// Generate API key
key, err := client.GenerateKey("user@example.com", "my-app")
```

## 🧪 Testing

```bash
# Clone the repository
git clone https://github.com/jep-protocol/sdk-go.git
cd sdk-go

# Run tests
go test -v ./...

# Run example
go run examples/main.go
```

## ❌ Error Handling

```go
resp, err := client.Judgment(&jep.JudgmentRequest{
    Entity: "user@example.com",
    Action: "approve",
})
if err != nil {
    switch e := err.(type) {
    case *jep.APIError:
        fmt.Printf("API error: %s (status: %d)\n", e.Message, e.StatusCode)
    case *jep.ValidationError:
        fmt.Printf("Validation error: %s\n", e.Message)
    default:
        fmt.Printf("Unexpected error: %v\n", e)
    }
    return
}
```

## 🔗 Related Repositories

- [Protocol Specification](https://github.com/jep-protocol/spec)
- [Core Implementation (Rust)](https://github.com/jep-protocol/core)
- [API Service](https://github.com/jep-protocol/api)
- [Python SDK](https://github.com/jep-protocol/sdk-py)
- [JavaScript SDK](https://github.com/jep-protocol/sdk-js)
- [CLI Tool](https://github.com/jep-protocol/cli)

## 📄 License

MIT License — see [LICENSE](LICENSE) for details.

## 🤝 Contributing

Contributions are welcome! Please:

- Open an [Issue](https://github.com/hjs-protocol/sdk-go/issues) for bugs or suggestions
- Submit Pull Requests for improvements
- See our [Contributing Guide](CONTRIBUTING.md) and [Code of Conduct](CODE_OF_CONDUCT.md)

---

**© 2026 HJS Foundation Ltd.**
```
