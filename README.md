# JEP Go SDK — JEP Core 0.7

Go client for the current [JEP Core 0.7](https://github.com/hjs-spec/jep-core) reference API.

Default endpoints:

```text
POST /v0.7/events/create
POST /v0.7/events/verify
GET  /health
```

Historical pre-0.7 verification is explicit through `VerifyEventLegacy`; no failed 0.7 event is automatically reinterpreted as 0.6.

## Status

Experimental reference SDK. It does not define new Core semantics or determine
truth, legal effect, authorization validity, causality, or policy outcome.

## Installation

```bash
go get github.com/hjs-spec/sdk-go
```

## JEP Core 0.7 model

- Event Identity is `(who,id)`; `id` is required.
- Core does not require a top-level nonce.
- Event Hash identifies an exact signed artifact, not the logical event identity.
- Validation returns `status` and independent `checks`, not a Validation Level.
- Acceptance may return `accepted` or `already_accepted`.
- D requires `delegatee + scope`.
- T requires `ref + termination_scope`.
- V requires `ref + verification_scope + result`.

## Quick start

```go
client := jep.NewClientWithURL("http://127.0.0.1:8000", "")

resp, err := client.CreateEvent(&jep.CreateEventRequest{
    Verb: jep.VerbJudgment,
    Who:  "did:example:agent-789",
    What: map[string]interface{}{"claim": "approve"},
})
if err != nil {
    log.Fatal(err)
}

result, err := client.VerifyEvent(&jep.VerifyEventRequest{
    Event: resp.Event,
    Mode:  "archival",
})
if err != nil {
    log.Fatal(err)
}

fmt.Println(resp.Event.ID)
fmt.Println(resp.EventHash)
fmt.Println(result.Status, result.Checks)
```

The reference event schema is maintained in [jep-core](https://github.com/hjs-spec/jep-core/blob/main/schemas/jep-event.schema.json).

## Legacy 0.6

Use `VerifyEventLegacy` only when the caller already knows the artifact is historical pre-0.7. Decoder selection must be explicit.

## Testing

```bash
go test ./...
```

## Related repositories

- JEP Core 0.7: https://github.com/hjs-spec/jep-core
- JEP API: https://github.com/hjs-spec/jep-api
- Python SDK: https://github.com/hjs-spec/sdk-py
- JavaScript SDK: https://github.com/hjs-spec/sdk-js

## License

MIT
