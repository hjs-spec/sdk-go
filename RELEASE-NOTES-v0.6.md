# JEP Go SDK v0.6.0 Release Notes

## Summary

This release upgrades the earlier Go SDK draft into a JEP v0.6 API SDK seed.

## Added

- JEP v0.6 event model.
- `/events/create` client.
- `/events/verify` client.
- `/health` client.
- JEP event, create event request, event response, and validation result types.
- J/D/T/V verb constants.
- Convenience helpers for Judgment, Delegation, Termination, and Verification.
- `httptest`-based unit tests.
- GitHub Actions workflow.

## Changed

- Replaced legacy `/judgments` and `/verify` API assumptions.
- Updated module path to `github.com/hjs-spec/jep-sdk-go`.
- Updated README and examples to current JEP v0.6 API seed.

## Boundary

This SDK is an implementation seed. It does not define new protocol semantics or claim production conformance.
