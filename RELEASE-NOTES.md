# JEP Core 0.7 migration

- Default API path is now `/v0.7/events/*`.
- Events use stable `id` and no longer require Core `nonce`.
- Validation results use `status`, independent `checks`, `event_identity`, and optional `acceptance`.
- D/T/V Core minimum shapes are validated before requests are sent.
- Historical pre-0.7 verification remains explicit; there is no automatic fallback.


The earlier 0.6 release added `conformance_class` passthrough. Current releases retain it while using the Core 0.7 profile; the wire major remains `jep: "1"`.
