use axum::{Router};
use wk_logger as logger;
use wk_config::AppConfig;
use wk_web::build_router;

#[tokio::main]
async fn main() {
    logger::init();

    let cfg = AppConfig::load().expect("load config");
    let addr = cfg.bind_addr();

    let app = build_router(&cfg);

    tracing::info!(%addr, "starting api server");
    let listener = tokio::net::TcpListener::bind(&addr).await.expect("bind addr");
    axum::serve(listener, app).await.expect("serve");
}
