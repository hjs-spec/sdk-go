# JEP Go SDK v0.6

Go client for the JEP-Core-0.6 API (wire version `"1"`). SDK release versions are separate from the protocol version.

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

Supported verbs:

```go
jep.VerbJudgment
jep.VerbDelegation
jep.VerbTermination
jep.VerbVerification
```

## API and helpers

The quickstart above demonstrates event creation and archival verification. The client also exposes helpers for the four verbs; see [client methods and types](client.go) for signatures and options.

For object-form `what`, `D` requires a claim, delegatee, and scope; `T` requires a claim, target, and termination scope; `V` requires a verification scope and non-null reference. Digest-form claims are also supported. Use the actual returned event hash for an event reference. See the [event schema](https://github.com/hjs-spec/jep-v06/blob/main/schemas/jep-event.schema.json) for the full requirements.

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

## Public Drafts

- JEP-Core: https://datatracker.ietf.org/doc/draft-wang-jep-judgment-event-protocol/
- JEP-Profiles: https://datatracker.ietf.org/doc/draft-wang-jep-profiles/
- JEP-Conformance: https://datatracker.ietf.org/doc/draft-wang-jep-conformance/

## License

MIT
