# JEP Go SDK v0.6

Go client for the [JEP-Core-0.6](https://github.com/hjs-spec/jep-v06) API (wire version `"1"`). SDK release versions are separate from the protocol version. See the protocol repository for core semantics, profiles, and public drafts.

This SDK targets the current JEP API shape:

```text
POST /events/create
POST /events/verify
GET  /health
```

## Status

Experimental implementation seed.

This SDK does not define new JEP-Core semantics and does not determine legal liability, factual truth, regulatory compliance, or complete-log availability.

## Installation

```bash
go get github.com/hjs-spec/sdk-go
```

## Quick Start

Start the [local API](https://github.com/hjs-spec/jep-api#run-locally) before running this example. Verification uses that API's configured trusted keys.

```go
package main

import (
    "fmt"
    "log"

    jep "github.com/hjs-spec/sdk-go"
)

func main() {
    client := jep.NewClientWithURL("http://127.0.0.1:8000", "")

    resp, err := client.CreateEvent(&jep.CreateEventRequest{
        Verb: jep.VerbJudgment,
        Who:  "did:example:agent-789",
        What: map[string]interface{}{
            "claim": "approve",
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(resp.EventHash)

    result, err := client.VerifyEvent(&jep.VerifyEventRequest{
        Event: resp.Event,
        Mode: "archival",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(result.Valid)
}
```

## Core Types

See [client.go](client.go) for the current event, request, and result types, including preservation of signed JSON members.

## API and helpers

The quickstart above demonstrates event creation and archival verification. The client also exposes helpers for the four verbs; see [client methods and types](client.go) for signatures and options.

Claim fields and reference requirements are defined in the [Core-0.6 event schema](https://github.com/hjs-spec/jep-v06/blob/main/schemas/jep-event.schema.json). For an event reference, use the actual returned event hash.

### Health

```go
health, err := client.Health()
```

## Validation results

Validation results preserve the API's `conformance_class` and diagnostic fields (`code`, `message`, `level`, `recoverable`). Older servers may omit the class; the SDK does not infer conformance.

## Testing

```bash
go test ./...
```

Tests use `httptest` and do not require a live API server.

## Related Repositories

- JEP v0.6: https://github.com/hjs-spec/jep-v06
- JEP API v0.6: https://github.com/hjs-spec/jep-api
- HJS v0.5: https://github.com/hjs-spec/hjs-05
- JAC v0.5: https://github.com/hjs-spec/jac-agent-02

## License

MIT
