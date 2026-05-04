# JEP Go SDK v0.6

Go SDK for the JEP v0.6 API seed.

This SDK targets the current JEP API shape:

```text
POST /events/create
POST /events/verify
GET  /health
```

It is aligned with:

- `draft-wang-jep-judgment-event-protocol-06`
- `draft-wang-jep-profiles-00`
- `draft-wang-jep-conformance-00`
- `hjs-spec/jep-api`

## Status

Experimental implementation seed.

This SDK does not define new JEP-Core semantics and does not determine legal liability, factual truth, regulatory compliance, or complete-log availability.

## Installation

```bash
go get github.com/hjs-spec/jep-sdk-go
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    jep "github.com/hjs-spec/jep-sdk-go"
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

```go
type JEPEvent struct {
    JEP     string                 `json:"jep"`
    Verb    Verb                   `json:"verb"`
    Who     string                 `json:"who"`
    When    int64                  `json:"when"`
    What    interface{}            `json:"what,omitempty"`
    Nonce   string                 `json:"nonce"`
    Aud     string                 `json:"aud,omitempty"`
    Ref     *string                `json:"ref,omitempty"`
    Ext     map[string]interface{} `json:"ext,omitempty"`
    ExtCrit []string               `json:"ext_crit,omitempty"`
    Sig     string                 `json:"sig,omitempty"`
}
```

Supported verbs:

```go
jep.VerbJudgment
jep.VerbDelegation
jep.VerbTermination
jep.VerbVerification
```

## API

### Create event

```go
resp, err := client.CreateEvent(&jep.CreateEventRequest{
    Verb: jep.VerbJudgment,
    Who:  "did:example:agent",
    What: "sha256:...",
})
```

### Verify event

```go
result, err := client.VerifyEvent(&jep.VerifyEventRequest{
    Event: resp.Event,
    Mode: "archival",
})
```

### Convenience helpers

```go
client.Judgment("did:example:agent", what)
client.Delegation("did:example:agent", what)
client.Termination("did:example:agent", what, &ref)
client.Verification("did:example:agent", what, ref)
```

### Health

```go
health, err := client.Health()
```

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

## Public Drafts

- JEP-Core: https://datatracker.ietf.org/doc/draft-wang-jep-judgment-event-protocol/
- JEP-Profiles: https://datatracker.ietf.org/doc/draft-wang-jep-profiles/
- JEP-Conformance: https://datatracker.ietf.org/doc/draft-wang-jep-conformance/

## License

MIT
