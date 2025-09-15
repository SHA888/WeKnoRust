use tracing_subscriber::{fmt, EnvFilter, prelude::*, fmt::time::ChronoUtc};

pub fn init() {
    let env_filter = EnvFilter::try_from_default_env()
        .or_else(|_| EnvFilter::try_new("info"))
        .unwrap();

    let fmt_layer = fmt::layer()
        .with_timer(ChronoUtc::rfc3339())
        .with_thread_ids(true)
        .with_target(true)
        .json();

    tracing_subscriber::registry()
        .with(env_filter)
        .with(fmt_layer)
        .init();
}
