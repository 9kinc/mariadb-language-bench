# MariaDB HTTP language benchmark

Public, disposable GitHub Actions benchmark comparing **Go**, **Rust**, **Node.js 24** and **PHP 8.5 (Nginx + PHP-FPM)** against **MariaDB 13.0.2**.

- Each of users, profiles and orders contains **1,000,000 deterministic synthetic rows**, with primary/secondary InnoDB indexes.
- Every GET /lookup?id=N performs **the exact same parameterised three-table INNER JOIN**. Every successful response is checked for ID, username, balance and non-empty joined fields.
- Go uses net/http + database/sql; Rust uses axum + sqlx; Node uses native http + mysql2; PHP uses Nginx + PHP-FPM + PDO. No production databases or third-party services.
- The backends share a 32-connection database pool ceiling (PHP-FPM max children 32). MariaDB runs in a fresh isolated service container.
- Scenarios run sequentially at **1, 16, 64 and 128 concurrent clients**, 3s warm-up + 10s measurement per stage, same runner. Counts, latency p50/p95/p99, errors and RPS are saved as JSON and in results/summary.md.
- Runs on **push to benchmark/** branches or manual workflow dispatch. All data disappears when GitHub terminates the runner. No merging of the benchmark branch is necessary.

## Important methodological limits

This compares application stacks, not isolated languages. Database-driver differences, PHP-FPM process lifecycle, SQL preparation, scheduling, cache warmth and shared-runner variability can dominate. Client and MariaDB consume CPU on the same runner. The client is **closed-loop**, not constant arrival rate, and p99 is conditional on successful requests; use error counts when interpreting results.

The first runner run is a smoke/performance baseline. For robust rankings repeat across fresh runs, rotate language order, report standard deviation/confidence intervals, and consider putting the load generator and DB on separate dedicated machines. Do not compare with internet benchmarks on unrelated hardware.

## Safety

The project uses only synthetic data and throwaway service credentials. No production repositories are modified. The workflow requests read-only repository permission and uploads results as a GitHub Actions artifact.
