use axum::{routing::get, Router, response::IntoResponse, Json};
use serde::Serialize;
use wk_config::AppConfig;

#[derive(Serialize)]
struct HealthResponse {
    status: &'static str,
}

async fn health() -> impl IntoResponse {
    Json(HealthResponse { status: "ok" })
}

pub fn build_router(_cfg: &AppConfig) -> Router {
    Router::new()
        .route("/health", get(health))
}
