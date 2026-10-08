package main

import (
 "database/sql"
 "context"
 "encoding/json"
 "log"
 "net/http"
 "os"
 "strconv"
 "time"
 _ "github.com/go-sql-driver/mysql"
)

const query = "SELECT u.id,u.username,u.balance_cents,p.bio,o.amount_cents FROM users AS u INNER JOIN profiles AS p ON p.user_id=u.id INNER JOIN orders AS o ON o.user_id=u.id WHERE u.id = ? LIMIT 1"

type Record struct {
 ID int64 `json:"id"`
 Username string `json:"username"`
 BalanceCents int64 `json:"balance_cents"`
 Bio string `json:"bio"`
 AmountCents int64 `json:"amount_cents"`
}
func main() {
 dsn := os.Getenv("MYSQL_DSN")
 if dsn == "" { log.Fatal("MYSQL_DSN required") }
 db,err := sql.Open("mysql", dsn); if err != nil {log.Fatal(err)}
 defer db.Close()
 db.SetMaxOpenConns(32);db.SetMaxIdleConns(32);db.SetConnMaxLifetime(4*time.Minute)
 if err=db.Ping();err!=nil{log.Fatal(err)}
 stmt,err:=db.Prepare(query);if err!=nil{log.Fatal(err)}
 defer stmt.Close()
 mux:=http.NewServeMux()
 mux.HandleFunc("GET /health",func(w http.ResponseWriter,r *http.Request){w.Write([]byte("ok"))})
 mux.HandleFunc("GET /lookup",func(w http.ResponseWriter,r *http.Request){
  raw:=r.URL.Query().Get("id"); id,e:=strconv.ParseUint(raw,10,32)
  if e!=nil||id==0||id>1000000{http.Error(w,"invalid id",http.StatusBadRequest);return}
  ctx,cancel:=context.WithTimeout(r.Context(), 3*time.Second);defer cancel()
  var out Record
  e=stmt.QueryRowContext(ctx,id).Scan(&out.ID,&out.Username,&out.BalanceCents,&out.Bio,&out.AmountCents)
  if e==sql.ErrNoRows{http.Error(w,"not found",http.StatusNotFound);return}
  if e!=nil{log.Printf("db error: %v",e);http.Error(w,"database unavailable",http.StatusServiceUnavailable);return}
  w.Header().Set("Content-Type","application/json")
  _=json.NewEncoder(w).Encode(out)
 })
 s:=&http.Server{Addr:":8080",Handler:mux,ReadHeaderTimeout:3*time.Second,IdleTimeout:30*time.Second}
 log.Fatal(s.ListenAndServe())
}
