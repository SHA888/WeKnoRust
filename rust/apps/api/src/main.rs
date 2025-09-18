use std::sync::Arc;
use axum::Router;
use wk_logger as logger;
use wk_config::AppConfig;
use wk_web::{build_router_with_state, AppState};
use wk_repos::init_pool;
use wk_stream::{self, StreamManager};
use std::env;

#[tokio::main]
async fn main() {
    logger::init();

    let cfg = AppConfig::load().expect("load config");
    let addr = cfg.bind_addr();
    let mut state = AppState { cfg: cfg.clone(), pool: None, stream: None };

    // Optional DB pool
    if env::var("DATABASE_URL").is_ok() {
        match init_pool().await {
            Ok(pool) => state.pool = Some(pool),
            Err(err) => tracing::warn!(?err, "failed to init DB pool"),
        }
    }

    // Stream manager selection via env: STREAM_MANAGER_TYPE=redis|memory
    let sm_type = env::var("STREAM_MANAGER_TYPE").unwrap_or_else(|_| "memory".to_string());
    state.stream = if sm_type.eq_ignore_ascii_case("redis") {
        let addr = env::var("REDIS_ADDR").unwrap_or_else(|_| "127.0.0.1:6379".into());
        let pw = env::var("REDIS_PASSWORD").ok();
        let db = env::var("REDIS_DB").ok().and_then(|s| s.parse::<i64>().ok());
        let prefix = env::var("REDIS_PREFIX").ok();
        let ttl = None; // default in impl
        match wk_stream::redis_impl::RedisStreamManager::new(&addr, pw.as_deref(), db, prefix.as_deref(), ttl).await {
            Ok(mgr) => Some(Arc::new(mgr) as Arc<dyn StreamManager>),
            Err(err) => { tracing::warn!(?err, "failed to init redis stream manager, fallback to memory"); Some(Arc::new(wk_stream::memory::MemoryStreamManager::new()) as Arc<dyn StreamManager>) }
        }
    } else {
        Some(Arc::new(wk_stream::memory::MemoryStreamManager::new()) as Arc<dyn StreamManager>)
    };

    let app = build_router_with_state(Arc::new(state));

    tracing::info!(%addr, "starting api server");
    let listener = tokio::net::TcpListener::bind(&addr).await.expect("bind addr");
    axum::serve(listener, app).await.expect("serve");
}
