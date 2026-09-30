# JEP Go SDK — JEP Core 0.7

Go client for the current [JEP Core 0.7](https://github.com/hjs-spec/jep-core) reference API.

## Status

Experimental HTTP client. Event creation and verification run on the configured
API. Start the [local reference API](https://github.com/hjs-spec/jep-quickstart#start-a-local-api)
before running the examples below.

## Installation

```bash
go get github.com/hjs-spec/sdk-go@v0.7.2
```

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

- Core contract and implementation path: https://github.com/hjs-spec/jep-core#current-contract
- JEP API: https://github.com/hjs-spec/jep-api
- Python SDK: https://github.com/hjs-spec/sdk-py
- JavaScript SDK: https://github.com/hjs-spec/sdk-js

## License

MIT

## Signed-event transport

Imported events retain explicitly present empty extension objects/arrays and unknown members. Unknown members must reach the verifier so they cannot be silently removed from an invalid event. Nested numeric values decode as `json.Number` to preserve the received numeric token; callers can explicitly convert after validation. Editing exported fields changes the transmitted event and requires a new signature. The SDK delegates cryptographic and profile checks to the API.
