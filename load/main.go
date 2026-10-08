package main

import (
 "encoding/json"
 "flag"
 "fmt"
 "io"
 "math"
 "net/http"
 "os"
 "sort"
 "sync"
 "sync/atomic"
 "time"
)

type Record struct {
 ID uint32 `json:"id"`
 Username string `json:"username"`
 BalanceCents uint32 `json:"balance_cents"`
 Bio string `json:"bio"`
 AmountCents uint32 `json:"amount_cents"`
}
type Result struct {
 Backend string `json:"backend"`
 Concurrency int `json:"concurrency"`
 DurationSeconds float64 `json:"duration_seconds"`
 Requests int `json:"requests"`
 Success int `json:"success"`
 Errors int `json:"errors"`
 RPS float64 `json:"rps"`
 P50ms float64 `json:"p50_ms"`
 P95ms float64 `json:"p95_ms"`
 P99ms float64 `json:"p99_ms"`
}
func percentile(v []float64,p float64) float64 {
 if len(v)==0{return 0}
 idx:=int(math.Ceil(p*float64(len(v))))-1
 if idx<0 {idx=0};return v[idx]
}
func main(){
 backend:=flag.String("backend","","backend name")
 concurrency:=flag.Int("concurrency",16,"concurrent connections")
 duration:=flag.Duration("duration",12*time.Second,"measured duration")
 output:=flag.String("output","","output path")
 flag.Parse()
 if *concurrency<1 || *concurrency>512 || *duration<=0 || *backend=="" {fmt.Fprintln(os.Stderr,"invalid benchmark args");os.Exit(2)}
 transport:=&http.Transport{MaxIdleConns:1024,MaxIdleConnsPerHost:1024,MaxConnsPerHost:1024,IdleConnTimeout:30*time.Second}
 defer transport.CloseIdleConnections()
 client:=&http.Client{Transport:transport,Timeout:5*time.Second}
 var serial uint64
 var mu sync.Mutex
 var latencies []float64
 var successes, errors int64
 var wg sync.WaitGroup
 start:=time.Now()
 deadline:=start.Add(*duration)
 for i:=0;i<*concurrency;i++ {
  wg.Add(1)
  go func(){
   defer wg.Done()
   local:=make([]float64,0,4096)
   var ok,bad int64
   for time.Now().Before(deadline) {
    number:=atomic.AddUint64(&serial,1)
    id:=uint32((number*6364136223846793005+1442695040888963407)%1000000)+1
    url:=fmt.Sprintf("http://127.0.0.1:8080/lookup?id=%d",id)
    started:=time.Now()
    resp,err:=client.Get(url)
    if err!=nil {bad++;continue}
    payload,err:=io.ReadAll(io.LimitReader(resp.Body,4096))
    _=resp.Body.Close()
    if err!=nil || resp.StatusCode!=200 {bad++;continue}
    var rec Record
    if json.Unmarshal(payload,&rec)!=nil ||
      rec.ID!=id || rec.Username!=fmt.Sprintf("user_%d",id) ||
      rec.BalanceCents!=id*37%100000 || rec.Bio=="" ||
      rec.AmountCents<100 {bad++;continue}
    ok++
    local=append(local,float64(time.Since(started).Microseconds())/1000)
   }
   atomic.AddInt64(&successes,ok);atomic.AddInt64(&errors,bad)
   mu.Lock();latencies=append(latencies,local...);mu.Unlock()
  }()
 }
 wg.Wait()
 elapsed:=time.Since(start).Seconds()
 sort.Float64s(latencies)
 res:=Result{
  Backend:*backend,Concurrency:*concurrency,DurationSeconds:elapsed,
  Requests:int(successes+errors),Success:int(successes),Errors:int(errors),
  RPS:float64(successes)/elapsed,
  P50ms:percentile(latencies,.5),P95ms:percentile(latencies,.95),P99ms:percentile(latencies,.99),
 }
 data,err:=json.MarshalIndent(res,"","  ");if err!=nil{panic(err)}
 fmt.Println(string(data))
 if *output!="" {
  if err=os.WriteFile(*output,append(data, byte(10)),0644);err!=nil {panic(err)}
 }
 if successes==0 {os.Exit(1)}
}
