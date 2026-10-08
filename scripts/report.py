import glob
import json
import os
import platform

os.makedirs("results",exist_ok=True)
results=[]
for file in glob.glob("results/*.json"):
    with open(file) as f:
        results.append(json.load(f))
languages={"go":0,"rust":1,"node":2,"php":3}
results.sort(key=lambda x:(x["concurrency"],languages.get(x["backend"],99)))
lines=[
 "# MariaDB language HTTP benchmark",
 "",
 "Measurements are sequential on one GitHub Actions VM; 3 million seeded synthetic records across three indexed InnoDB tables.",
 "The same 3-way parameterised indexed join executes on every successful HTTP request. No application-level cache.",
 "Load is generated on the **same runner** as the servers and DB; results reflect the complete stack and shared hardware, not pure language speed.",
 "",
 "| Concurrent clients | Backend | Successful req/s | p50 ms | p95 ms | p99 ms | Errors / total |",
 "|---:|---|---:|---:|---:|---:|---:|"
]
for r in results:
    lines.append(f'| {r["concurrency"]} | {r["backend"]} | {r["rps"]:,.1f} | {r["p50_ms"]:.2f} | {r["p95_ms"]:.2f} | {r["p99_ms"]:.2f} | {r["errors"]:,} / {r["requests"]:,} |')
lines.extend(["", f"Measurement entries: **{len(results)}/16**. Each stage: 3 s warm-up followed by 10 s measurement.",
 "Services are not simultaneously loaded; order is Go, Rust, Node.js, PHP. Repeated runs with rotated order are needed to establish stable rankings.",
 "Concurrency uses a closed-loop client; slower responses reduce request injection rate. Percentiles include successful requests only; failures are reported separately.",
 "Data integrity is checked for every 200 response, and startup smoke checks query row 1, 500000 and 1000000.",
 ""])
if len(results)==16:
    rows=[r for r in results if r["concurrency"]==64 and r["errors"]==0]
    if len(rows)==4:
        winner=max(rows,key=lambda r:r["rps"])
        lines.append(f'Highest successful throughput at 64 clients in this run: **{winner["backend"]}** ({winner["rps"]:,.1f} req/s).')
    else:
        lines.append("At least one backend had errors at 64 clients, so no error-free throughput winner is reported.")
else:
    lines.append("INCOMPLETE RUN: missing measurements; do not interpret as a final ranking.")
lines.extend(["","## Run environment",f"- OS: {platform.platform()}"])
for name in ["mariadb-version.txt","seed-timing.txt","counts.tsv"]:
    if os.path.exists(name):
        with open(name) as f:
            lines.append(f"- {name}: {f.read().strip().replace(chr(10),'; ')}")
with open("results/summary.md","w") as f:
    f.write("\n".join(lines)+"\n")
print("\n".join(lines))
