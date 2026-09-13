# Implementation hardening — September 2026

Preserve signed JSON when events pass through Go clients.

## Changes

Custom JSON marshaling preserves unchanged nulls, empty members, number literals and unknown fields. JEPEvent.Reference accepts raw structured references while Ref retains its string-pointer API. SignatureObject carries object containers. VerifyEventRequest adds ExpectedAudience.

## Validation

```sh
go test -race ./...
```

## Compatibility and remaining limits

No old archive data is rewritten. JEPEvent.Reference is used when Ref is nil. New J requests permit explicit null what.
