# Protocol Allocation Reduction Progress

Offline benches (no DB). Run:

```bash
go test ./ms/ -bench='BenchmarkUconv|BenchmarkPacket|BenchmarkMessage|BenchmarkToken|BenchmarkEncode' -benchmem -count=3 -run='^$'
```

## Baseline (as-is, before optimizations)

Date: 2026-07-14  
Commit: working tree before Tier-1 changes  
CPU: AMD Ryzen 7 5700G  

| Benchmark | ns/op | B/op | allocs/op |
|---|---:|---:|---:|
| UconvEncodeString | 391.4 | 656 | **4** |
| UconvEncodeBytes | 337.1 | 512 | **4** |
| UconvDecodeToBytes | 313.7 | 176 | **2** |
| UconvDecodeToString | 336.0 | 240 | **3** |
| PacketWriterSmallMessage | 186.0 | 240 | **1** |
| PacketWriterMultiPacket | 2289 | 12400 | **4** |
| MessageReaderFetch | 853.5 | 5280 | **6** |
| MessageReaderMultiPacket | 2889 | 17824 | **9** |
| TokenStreamDecodeRows (50 rows) | 15682 | 12124 | **216** |
| EncodeParamString | 388.0 | 112 | **5** |

### Notes on baseline

- `uconv` encode: ~4 allocs (string→[]byte, []rune, []uint16, output []byte).
- `PacketWriter` multi-packet: 1 `make` per outgoing TDS packet (~4KB frame).
- `MessageReader`: copies every packet body out of `sbuffer` (see TODO in `fill`).
- Token stream: 216 allocs / 50 rows ≈ **4.3 allocs/row** (int + nvarchar + overhead).

## Change log

| # | Change | Result | Notes |
|---|---|---|---|
| 0 | Baseline benches + this file | — | Starting point |
| 1 | PacketWriter: reuse frame + intScratch; pre-size message buffer | **progress** | Small/Multi: **0 allocs** (was 1/4). EncodeParamString: 5→4 allocs, 112→64 B |

### After #1 PacketWriter frame

| Benchmark | ns/op | B/op | allocs/op | Δ allocs |
|---|---:|---:|---:|---:|
| PacketWriterSmallMessage | 140.6 | 0 | **0** | −1 |
| PacketWriterMultiPacket | 680.3 | 0 | **0** | −4 |
| EncodeParamString | 361.1 | 64 | **4** | −1 |
