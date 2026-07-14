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

| 2 | uconv: direct UTF-8↔UTF-16 encode/decode (no []rune/utf16.Encode) | **progress** | Encode 4→1 alloc; Decode bytes 2→1, string 3→2; EncodeParam 4→2; TokenStream B back to baseline |

### After #2 uconv rewrite

| Benchmark | ns/op | B/op | allocs/op | Δ allocs vs baseline |
|---|---:|---:|---:|---:|
| UconvEncodeString | 209.9 | 144 | **1** | −3 |
| UconvEncodeBytes | 193.1 | 128 | **1** | −3 |
| UconvDecodeToBytes | 247.1 | 64 | **1** | −1 |
| UconvDecodeToString | 270.8 | 128 | **2** | −1 |
| EncodeParamString | 318.5 | 32 | **2** | −3 |
| TokenStreamDecodeRows | 14833 | 12124 | 216 | 0 (B restored after exact-size decode) |

| 3 | sbuffer 4× packet capacity | **reverted** | Same allocs; B/op rose with larger ring (17568 vs 5280 on MessageReaderFetch). No measured win; left at 1× packet. |
| 4–7 | MustCopy + MessageReader msgBuf reuse + decode emit(tds.dv) + non-NChar view | **progress** | TokenStream **216→117** allocs; MessageReader multi **9→6**; Fetch **6→5** |

### After #4–7 reader/decode/MustCopy

| Benchmark | ns/op | B/op | allocs/op | Δ allocs vs baseline |
|---|---:|---:|---:|---:|
| MessageReaderFetch | 921.1 | 5264 | **5** | −1 |
| MessageReaderMultiPacket | 2434 | 13200 | **6** | −3 |
| TokenStreamDecodeRows | 13810 | 10076 | **117** | **−99** |

| 8 | PacketWriter UCS2 scratch for encodeParam/SQL/values | **progress** | EncodeParamString **2→0** allocs |
| 9 | Connection utf8Scratch + uconv.AppendBytes for NChar decode | **progress** | TokenStream **117→68** allocs (was 216 baseline) |

### After #8–9 encode/decode scratch

| Benchmark | ns/op | B/op | allocs/op | Δ allocs vs baseline |
|---|---:|---:|---:|---:|
| EncodeParamString | 256.1 | 0 | **0** | −5 |
| TokenStreamDecodeRows | 12434 | 8900 | **68** | **−148** |

### Final summary vs baseline

| Benchmark | baseline allocs | final allocs | Δ |
|---|---:|---:|---:|
| UconvEncodeString | 4 | 1 | −3 |
| UconvEncodeBytes | 4 | 1 | −3 |
| UconvDecodeToBytes | 2 | 1 | −1 |
| UconvDecodeToString | 3 | 2 | −1 |
| PacketWriterSmallMessage | 1 | 0 | −1 |
| PacketWriterMultiPacket | 4 | 0 | −4 |
| MessageReaderFetch | 6 | 5 | −1 |
| MessageReaderMultiPacket | 9 | 6 | −3 |
| TokenStreamDecodeRows | 216 | **68** | **−148 (−69%)** |
| EncodeParamString | 5 | **0** | −5 |

### Skipped / future

| Item | Status |
|---|---|
| sbuffer 4× capacity | Reverted — higher B/op, no alloc win |
| Dual sbuffer / true zero-copy Fetch | Partial via msgBuf; full dual-buffer left for later |
| Avoid interface{} boxing for ints | Needs typed writeField/Prep path |
| True TDS prepare / fReuseMetaData | Prepare not implemented |
