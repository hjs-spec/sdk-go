# Release 0.7.1

- Preserve signed imported events through JSON decode/encode, including empty `ext`, empty `ext_crit`, unknown members and nested number tokens.
- Decode nested event numbers as `json.Number`; callers can convert explicitly after validation instead of accepting an implicit float64 conversion.
- Keep edits to exported event fields visible to the verifier.
- Distinguish required D `scope` and V `result` member presence from profile-defined values, including explicit null.

Core remains 0.7. The client transports events; it does not independently verify signatures or TSTO policy.
