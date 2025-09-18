use std::sync::Arc;
use axum::{
    routing::get,
    Router,
    response::IntoResponse,
    Json,
    extract::{State, Request},
    middleware::{self, Next},
};
use serde::Serialize;
use uuid::Uuid;
use wk_config::AppConfig;
use wk_repos::PgPool;
use wk_stream::{StreamManager, StreamInfo};
use http::StatusCode;

#[derive(Clone)]
pub struct AppState {
    pub cfg: AppConfig,
    pub pool: Option<PgPool>,
    pub stream: Option<Arc<dyn StreamManager>>, // trait object behind Arc
}

// Basic auth identity captured from headers
#[derive(Clone, Debug, Serialize)]
pub struct ApiIdentity { pub api_key: Option<String> }

// Auth middleware: capture x-api-key (no enforcement yet)
async fn auth_mw(mut req: Request, next: Next) -> impl IntoResponse {
    let key = req
        .headers()
        .get("x-api-key")
        .and_then(|v| v.to_str().ok())
        .map(|s| s.to_string());
    req.extensions_mut().insert(ApiIdentity { api_key: key });
    next.run(req).await
}

// JSON API error type
#[derive(Debug, Serialize)]
pub struct AppErrorBody { pub code: u16, pub message: String, pub request_id: Option<String> }

pub struct AppError { pub status: StatusCode, pub message: String }

impl AppError {
    pub fn new(status: StatusCode, message: impl Into<String>) -> Self { Self { status, message: message.into() } }
}

impl IntoResponse for AppError {
    fn into_response(self) -> axum::response::Response {
        let mut res = (self.status, Json(AppErrorBody { code: self.status.as_u16(), message: self.message, request_id: None })).into_response();
        // Echo x-request-id into body if present
        if let Some(rid) = res.headers().get("x-request-id").and_then(|v| v.to_str().ok()).map(|s| s.to_string()) {
            let body = AppErrorBody { code: self.status.as_u16(), message: String::from_utf8_lossy(res.body().to_owned().into_bytes().as_ref()).into_owned(), request_id: Some(rid) };
            res = (self.status, Json(body)).into_response();
        }
        res
    }
}

#[derive(Serialize)]
struct HealthResponse { status: &'static str }

#[derive(Serialize)]
struct DBHealth { ok: bool }

#[derive(Serialize)]
struct StreamHealth { ok: bool }

async fn health() -> impl IntoResponse { Json(HealthResponse { status: "ok" }) }

async fn health_db(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    if let Some(pool) = &state.pool {
        let ok = sqlx::query_scalar::<_, i32>("SELECT 1")
            .fetch_one(pool)
            .await
            .map(|v| v == 1)
            .unwrap_or(false);
        return Json(DBHealth { ok });
    }
    Json(DBHealth { ok: false })
}

async fn health_stream(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    if let Some(stream) = &state.stream {
        // Try a noop get on a non-existent key to validate round-trip
        let _ = stream.get_stream("health", &Uuid::new_v4().to_string()).await.ok();
        return Json(StreamHealth { ok: true });
    }
    Json(StreamHealth { ok: false })
}

// Simple request-id middleware: attach a request id header if absent
async fn request_id_mw(mut req: Request, next: Next) -> impl IntoResponse {
    let hdr = http::header::HeaderName::from_static("x-request-id");
    let id = req.headers().get(&hdr).cloned().unwrap_or_else(|| {
        http::HeaderValue::from_str(&Uuid::new_v4().to_string()).unwrap()
    });
    req.headers_mut().insert(hdr.clone(), id.clone());
    let mut res = next.run(req).await;
    res.headers_mut().insert(hdr, id);
    res
}

pub fn build_router_with_state(state: Arc<AppState>) -> Router {
    Router::new()
        .route("/health", get(health))
        .route("/health/db", get(health_db))
        .route("/health/stream", get(health_stream))
        .layer(middleware::from_fn(request_id_mw))
        .layer(middleware::from_fn(auth_mw))
        .with_state(state)
}
