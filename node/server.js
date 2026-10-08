"use strict";
const http = require("node:http");
const mysql = require("mysql2/promise");
const SQL = "SELECT u.id,u.username,u.balance_cents,p.bio,o.amount_cents FROM users AS u INNER JOIN profiles AS p ON p.user_id=u.id INNER JOIN orders AS o ON o.user_id=u.id WHERE u.id = ? LIMIT 1";
async function main() {
  const pool = mysql.createPool({
    host: process.env.DB_HOST || "127.0.0.1", port: 3306,
    user: process.env.DB_USER || "bench", password: process.env.DB_PASSWORD,
    database: "bench", connectionLimit: 32, waitForConnections: true,
    queueLimit: 0, enableKeepAlive: true, connectTimeout: 5000
  });
  await pool.query("SELECT 1");
  const server = http.createServer(async (req, res) => {
    const parsed = new URL(req.url, "http://localhost");
    if (req.method === "GET" && parsed.pathname === "/health") {res.writeHead(200);res.end("ok");return;}
    if (req.method !== "GET" || parsed.pathname !== "/lookup") {res.writeHead(404);res.end();return;}
    const raw = parsed.searchParams.get("id");
    if (!/^[0-9]+$/.test(raw || "")) {res.writeHead(400);res.end("invalid id");return;}
    const id = Number(raw);
    if (!Number.isSafeInteger(id) || id < 1 || id > 1000000) {res.writeHead(400);res.end("invalid id");return;}
    try {
      const [rows] = await pool.execute(SQL, [id]);
      if (rows.length !== 1) {res.writeHead(404);res.end();return;}
      res.writeHead(200, {"content-type":"application/json"});
      res.end(JSON.stringify(rows[0]));
    } catch (e) {
      console.error("db error",e.message);res.writeHead(503);res.end("database unavailable");
    }
  });
  server.listen(8080, "127.0.0.1");
  process.on("SIGTERM", () => server.close(() => pool.end()));
}
main().catch(e => {console.error(e);process.exit(1);});
