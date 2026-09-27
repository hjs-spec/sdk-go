# Release 0.7.2

- Validate original JSON before event import: reject duplicate members at every nesting level, invalid UTF-8, unpaired surrogate escapes, trailing JSON and non-object events. Failed imports do not change the receiver.
- Reject case-insensitive aliases of Core member names before Go struct decoding can treat them as signed fields.
- Preserve legitimate surrogate pairs, escaped backslashes, unknown fields, empty optional members and exact numeric tokens. The client is a transport, not an independent signature or profile verifier.
- Add hostile-input regressions and fuzz seeds. The Core interoperability gate separately checks real signed artifacts.

Protocol remains Core 0.7 / wire 1. Historical signed artifacts and published protocol drafts are unchanged.
