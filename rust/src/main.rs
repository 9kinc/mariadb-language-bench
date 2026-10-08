use axum::{extract::{Query, State}, http::StatusCode, routing::get, Json, Router};
use serde::{Deserialize, Serialize};
use sqlx::{mysql::MySqlPoolOptions, MySqlPool, Row};
use std::{env, time::Duration};

const SQL: &str = "SELECT u.id,u.username,u.balance_cents,p.bio,o.amount_cents FROM users AS u INNER JOIN profiles AS p ON p.user_id=u.id INNER JOIN orders AS o ON o.user_id=u.id WHERE u.id = ? LIMIT 1";

#[derive(Deserialize)]
struct Params { id: u32 }

#[derive(Serialize)]
struct Record {
    id: u32,
    username: String,
    balance_cents: u32,
    bio: String,
    amount_cents: u32
}

async fn lookup(State(pool): State<MySqlPool>, Query(p): Query<Params>) -> Result<Json<Record>, StatusCode> {
    if p.id == 0 || p.id > 1_000_000 { return Err(StatusCode::BAD_REQUEST); }
    let row = sqlx::query(SQL).bind(p.id)
        .fetch_one(&pool).await.map_err(|e| {
            eprintln!("database error: {e}");
            StatusCode::SERVICE_UNAVAILABLE
        })?;
    let record = Record {
        id: row.try_get("id").map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?,
        username: row.try_get("username").map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?,
        balance_cents: row.try_get("balance_cents").map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?,
        bio: row.try_get("bio").map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?,
        amount_cents: row.try_get("amount_cents").map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?
    };
    Ok(Json(record))
}

#[tokio::main]
async fn main() {
    let url = env::var("DATABASE_URL").expect("DATABASE_URL required");
    let pool = MySqlPoolOptions::new().max_connections(32)
        .acquire_timeout(Duration::from_secs(3))
        .connect(&url).await.expect("connect MariaDB");
    let app = Router::new()
        .route("/health", get(|| async { "ok" }))
        .route("/lookup", get(lookup)).with_state(pool);
    let listener = tokio::net::TcpListener::bind("127.0.0.1:8080").await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
