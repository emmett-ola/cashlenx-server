# Performance Baseline

This document records the `v0.9.0` performance baseline and cache decision.
Results are regression reference points, not cross-machine service-level targets.

## Service Baseline

Recorded on 2026-07-01 with Go 1.26.1 on Windows/amd64 and an Intel Core
i7-14700F. Each value is the median of five benchmark runs. Fixtures contain
10 categories and deterministic in-memory cash-flow mapper results.

| Operation | Transactions | Median | Bytes/op | Allocs/op |
| --- | ---: | ---: | ---: | ---: |
| Cash-flow summary | 100 | 34.9 us | 37,720 | 286 |
| Cash-flow summary | 1,000 | 354.9 us | 370,361 | 2,806 |
| Cash-flow summary | 10,000 | 3.37 ms | 3,696,763 | 28,006 |
| Statistic summary | 100 | 25.5 us | 22,408 | 206 |
| Statistic summary | 1,000 | 239.3 us | 216,808 | 2,006 |
| Statistic summary | 10,000 | 2.33 ms | 2,160,811 | 20,006 |
| Dashboard overview | 100 | 145.0 us | 152,184 | 1,099 |
| Dashboard overview | 1,000 | 1.52 ms | 1,381,410 | 9,292 |
| Dashboard overview | 10,000 | 14.08 ms | 16,848,004 | 90,416 |

Run the baseline with:

```bash
go test -run '^$' -bench 'Summary|Dashboard' -benchmem -count=5 \
  ./service/cash_flow_service ./service/statistic_service
```

## Mapper Baseline

Recorded on the same host against fresh local MongoDB 7.0 and MySQL 8.0
containers. Each query benchmark seeds user-isolated fixtures and removes them
afterward.

| Backend | Operation | Rows | Time | Bytes/op | Allocs/op |
| --- | --- | ---: | ---: | ---: | ---: |
| MongoDB | Date range and user | 1,000 | 12.22 ms | 5,630,995 | 137,268 |
| MongoDB | Date range and user | 10,000 | 88.59 ms | 60,645,249 | 1,370,397 |
| MongoDB | Filtered, max 100 results | 1,000 | 2.12 ms | 558,050 | 13,932 |
| MongoDB | Filtered, max 100 results | 10,000 | 4.61 ms | 558,090 | 13,931 |
| MySQL | Date range and user | 1,000 | 4.15 ms | 1,439,817 | 30,087 |
| MySQL | Date range and user | 10,000 | 17.16 ms | 18,488,397 | 299,941 |
| MySQL | Filtered, max 100 results | 1,000 | 3.07 ms | 140,725 | 3,102 |
| MySQL | Filtered, max 100 results | 10,000 | 4.49 ms | 140,667 | 3,102 |

Database, storage, and container differences make absolute mapper timings
unsuitable as cross-machine targets. Use them for same-environment regression
comparisons with the commands in [`roadmap.md`](roadmap.md).

## v0.9.0 Cache Decision

Do not add a recent-query cache in `v0.9.0`.

The summary paths remain below 4 ms with the deliberately large 10,000-row
fixture. The dashboard reaches about 14 ms at that size, while mapper timings
range from about 17 ms on MySQL to 89 ms on MongoDB when materializing all
10,000 rows. Typical filtered reads remain below 5 ms. The dashboard's 16.8 MB
and 90,416 allocations identify repeated in-request aggregation and response
construction as the better future optimization target. A read-through cache
would require coherent invalidation for cash-flow and category writes,
imports, restores, and account deletion while serving mutable user-scoped
data. That complexity is not justified without production evidence of
frequent repeated 10,000-row reads.

Reconsider a bounded, user-scoped cache only after production traces show a
material repeated-read bottleneck. Reconsider a shared cache only with
multi-instance deployment and a defined cross-process invalidation model.
