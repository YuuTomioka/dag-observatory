State Hash Rules (Draft)

Purpose
- Deterministic comparison for backtest/replay
- Audit signal (not a cryptographic proof)

Hash 대상 (Stable State)
- Deterministic values only
- Exclude or normalize:
  - time.Time / timestamps
  - unordered map types (must be normalized)
  - random/UUID/external I/O volatile values

Normalization rules
1) Key ordering
   - Sort by StableID then Name (string ascending)
2) Value encoding
   - Prefer a stable codec (JSON with fixed field names)
   - Avoid encoding that depends on map iteration order
3) Compatibility
   - Changing key StableID/Name or codec breaks hash compatibility

Responsibility boundaries
- domain: state.Hasher interface only
- infrastructure: Memory/Bolt hasher implementations
- application: selects hasher strategy per workflow
