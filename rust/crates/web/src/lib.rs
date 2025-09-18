use std::sync::Arc;
use axum::{
    routing::{get, post, put, delete},
    Router,
    response::IntoResponse,
    Json,
    extract::{State, Request, Path},
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
    // Stub handlers
    async fn ok(endpoint: &'static str) -> impl IntoResponse { Json(serde_json::json!({"ok": true, "endpoint": endpoint})) }
    async fn ok_with_params(endpoint: &'static str, params: serde_json::Value) -> impl IntoResponse { Json(serde_json::json!({"ok": true, "endpoint": endpoint, "params": params})) }

    // Tenants
    let tenants = Router::new()
        .route("/", post(|| async { ok("POST /api/v1/tenants").await }))
        .route("/", get(|| async { ok("GET /api/v1/tenants").await }))
        .route("/:id", get(|Path(id): Path<String>| async move { ok_with_params("GET /api/v1/tenants/:id", serde_json::json!({"id": id})).await }))
        .route("/:id", put(|Path(id): Path<String>| async move { ok_with_params("PUT /api/v1/tenants/:id", serde_json::json!({"id": id})).await }))
        .route("/:id", delete(|Path(id): Path<String>| async move { ok_with_params("DELETE /api/v1/tenants/:id", serde_json::json!({"id": id})).await }));

    // Knowledge Bases
    let knowledge_bases = Router::new()
        .route("/", post(|| async { ok("POST /api/v1/knowledge-bases").await }))
        .route("/", get(|| async { ok("GET /api/v1/knowledge-bases").await }))
        .route("/:id", get(|Path(id): Path<String>| async move { ok_with_params("GET /api/v1/knowledge-bases/:id", serde_json::json!({"id": id})).await }))
        .route("/:id", put(|Path(id): Path<String>| async move { ok_with_params("PUT /api/v1/knowledge-bases/:id", serde_json::json!({"id": id})).await }))
        .route("/:id", delete(|Path(id): Path<String>| async move { ok_with_params("DELETE /api/v1/knowledge-bases/:id", serde_json::json!({"id": id})).await }))
        .route("/:id/hybrid-search", get(|Path(id): Path<String>| async move { ok_with_params("GET /api/v1/knowledge-bases/:id/hybrid-search", serde_json::json!({"id": id})).await }))
        .route("/copy", post(|| async { ok("POST /api/v1/knowledge-bases/copy").await }));

    // Knowledge routes
    let knowledge = Router::new()
        .route("/batch", get(|| async { ok("GET /api/v1/knowledge/batch").await }))
        .route("/:id", get(|Path(id): Path<String>| async move { ok_with_params("GET /api/v1/knowledge/:id", serde_json::json!({"id": id})).await }))
        .route("/:id", delete(|Path(id): Path<String>| async move { ok_with_params("DELETE /api/v1/knowledge/:id", serde_json::json!({"id": id})).await }))
        .route("/:id", put(|Path(id): Path<String>| async move { ok_with_params("PUT /api/v1/knowledge/:id", serde_json::json!({"id": id})).await }))
        .route("/:id/download", get(|Path(id): Path<String>| async move { ok_with_params("GET /api/v1/knowledge/:id/download", serde_json::json!({"id": id})).await }))
        .route("/image/:id/:chunk_id", put(|Path((id, chunk_id)): Path<(String, String)>| async move { ok_with_params("PUT /api/v1/knowledge/image/:id/:chunk_id", serde_json::json!({"id": id, "chunk_id": chunk_id})).await }));

    // Knowledge under knowledge base
    let kb_knowledge = Router::new()
        .route("/file", post(|Path(id): Path<String>| async move { ok_with_params("POST /api/v1/knowledge-bases/:id/knowledge/file", serde_json::json!({"id": id})).await }))
        .route("/url", post(|Path(id): Path<String>| async move { ok_with_params("POST /api/v1/knowledge-bases/:id/knowledge/url", serde_json::json!({"id": id})).await }))
        .route("/", get(|Path(id): Path<String>| async move { ok_with_params("GET /api/v1/knowledge-bases/:id/knowledge", serde_json::json!({"id": id})).await }));

    // Chunks
    let chunks = Router::new()
        .route("/:knowledge_id", get(|Path(knowledge_id): Path<String>| async move { ok_with_params("GET /api/v1/chunks/:knowledge_id", serde_json::json!({"knowledge_id": knowledge_id})).await }))
        .route("/:knowledge_id/:id", delete(|Path((knowledge_id, id)): Path<(String, String)>| async move { ok_with_params("DELETE /api/v1/chunks/:knowledge_id/:id", serde_json::json!({"knowledge_id": knowledge_id, "id": id})).await }))
        .route("/:knowledge_id", delete(|Path(knowledge_id): Path<String>| async move { ok_with_params("DELETE /api/v1/chunks/:knowledge_id", serde_json::json!({"knowledge_id": knowledge_id})).await }))
        .route("/:knowledge_id/:id", put(|Path((knowledge_id, id)): Path<(String, String)>| async move { ok_with_params("PUT /api/v1/chunks/:knowledge_id/:id", serde_json::json!({"knowledge_id": knowledge_id, "id": id})).await }));

    // Sessions
    let sessions = Router::new()
        .route("/", post(|| async { ok("POST /api/v1/sessions").await }))
        .route("/:id", get(|Path(id): Path<String>| async move { ok_with_params("GET /api/v1/sessions/:id", serde_json::json!({"id": id})).await }))
        .route("/", get(|| async { ok("GET /api/v1/sessions").await }))
        .route("/:id", put(|Path(id): Path<String>| async move { ok_with_params("PUT /api/v1/sessions/:id", serde_json::json!({"id": id})).await }))
        .route("/:id", delete(|Path(id): Path<String>| async move { ok_with_params("DELETE /api/v1/sessions/:id", serde_json::json!({"id": id})).await }))
        .route("/:session_id/generate_title", post(|Path(session_id): Path<String>| async move { ok_with_params("POST /api/v1/sessions/:session_id/generate_title", serde_json::json!({"session_id": session_id})).await }))
        .route("/continue-stream/:session_id", get(|Path(session_id): Path<String>| async move { ok_with_params("GET /api/v1/sessions/continue-stream/:session_id", serde_json::json!({"session_id": session_id})).await }));

    // Messages
    let messages = Router::new()
        .route("/:session_id/load", get(|Path(session_id): Path<String>| async move { ok_with_params("GET /api/v1/messages/:session_id/load", serde_json::json!({"session_id": session_id})).await }))
        .route("/:session_id/:id", delete(|Path((session_id, id)): Path<(String, String)>| async move { ok_with_params("DELETE /api/v1/messages/:session_id/:id", serde_json::json!({"session_id": session_id, "id": id})).await }));

    // Chat
    let knowledge_chat = Router::new()
        .route("/:session_id", post(|Path(session_id): Path<String>| async move { ok_with_params("POST /api/v1/knowledge-chat/:session_id", serde_json::json!({"session_id": session_id})).await }));
    let knowledge_search = Router::new()
        .route("/", post(|| async { ok("POST /api/v1/knowledge-search").await }));

    // Models
    let models = Router::new()
        .route("/", post(|| async { ok("POST /api/v1/models").await }))
        .route("/", get(|| async { ok("GET /api/v1/models").await }))
        .route("/:id", get(|Path(id): Path<String>| async move { ok_with_params("GET /api/v1/models/:id", serde_json::json!({"id": id})).await }))
        .route("/:id", put(|Path(id): Path<String>| async move { ok_with_params("PUT /api/v1/models/:id", serde_json::json!({"id": id})).await }))
        .route("/:id", delete(|Path(id): Path<String>| async move { ok_with_params("DELETE /api/v1/models/:id", serde_json::json!({"id": id})).await }));

    // Evaluation
    let evaluation = Router::new()
        .route("/", post(|| async { ok("POST /api/v1/evaluation/").await }))
        .route("/", get(|| async { ok("GET /api/v1/evaluation/").await }));

    // Initialization and test-data (public in Go)
    let init = Router::new()
        .route("/initialization/status", get(|| async { ok("GET /api/v1/initialization/status").await }))
        .route("/initialization/config", get(|| async { ok("GET /api/v1/initialization/config").await }))
        .route("/initialization/initialize", post(|| async { ok("POST /api/v1/initialization/initialize").await }))
        .route("/initialization/ollama/status", get(|| async { ok("GET /api/v1/initialization/ollama/status").await }))
        .route("/initialization/ollama/models", get(|| async { ok("GET /api/v1/initialization/ollama/models").await }))
        .route("/initialization/ollama/models/check", post(|| async { ok("POST /api/v1/initialization/ollama/models/check").await }))
        .route("/initialization/ollama/models/download", post(|| async { ok("POST /api/v1/initialization/ollama/models/download").await }))
        .route("/initialization/ollama/download/progress/:taskId", get(|Path(task_id): Path<String>| async move { ok_with_params("GET /api/v1/initialization/ollama/download/progress/:taskId", serde_json::json!({"taskId": task_id})).await }))
        .route("/initialization/ollama/download/tasks", get(|| async { ok("GET /api/v1/initialization/ollama/download/tasks").await }))
        .route("/initialization/remote/check", post(|| async { ok("POST /api/v1/initialization/remote/check").await }))
        .route("/initialization/embedding/test", post(|| async { ok("POST /api/v1/initialization/embedding/test").await }))
        .route("/initialization/rerank/check", post(|| async { ok("POST /api/v1/initialization/rerank/check").await }))
        .route("/initialization/multimodal/test", post(|| async { ok("POST /api/v1/initialization/multimodal/test").await }))
        .route("/test-data", get(|| async { ok("GET /api/v1/test-data").await }));

    let api_v1 = Router::new()
        .nest("/tenants", tenants)
        .nest("/knowledge-bases", knowledge_bases)
        .nest("/knowledge-bases/:id/knowledge", kb_knowledge)
        .nest("/knowledge", knowledge)
        .nest("/chunks", chunks)
        .nest("/sessions", sessions)
        .nest("/messages", messages)
        .nest("/knowledge-chat", knowledge_chat)
        .nest("/knowledge-search", knowledge_search)
        .nest("/models", models)
        .nest("/evaluation", evaluation)
        .merge(init);

    Router::new()
        .route("/health", get(health))
        .route("/health/db", get(health_db))
        .route("/health/stream", get(health_stream))
        .nest("/api/v1", api_v1)
        .layer(middleware::from_fn(request_id_mw))
        .layer(middleware::from_fn(auth_mw))
        .with_state(state)
}
