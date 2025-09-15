use config as cfg;
use serde::Deserialize;
use thiserror::Error;
use dotenvy::dotenv;

#[derive(Debug, Error)]
pub enum ConfigError {
    #[error("config build error: {0}")]
    Build(#[from] cfg::ConfigError),
}

#[derive(Debug, Clone, Deserialize)]
pub struct ServerConfig {
    pub host: Option<String>,
    pub port: Option<u16>,
}

#[derive(Debug, Clone, Deserialize)]
pub struct AppConfig {
    pub server: Option<ServerConfig>,
}

impl Default for AppConfig {
    fn default() -> Self {
        Self {
            server: Some(ServerConfig {
                host: Some("0.0.0.0".to_string()),
                port: Some(8080),
            }),
        }
    }
}

impl AppConfig {
    pub fn load() -> Result<Self, ConfigError> {
        // Load .env if present
        dotenv().ok();

        let builder = cfg::Config::builder()
            // optional: YAML at config/config.yaml
            .add_source(cfg::File::with_name("config/config").required(false))
            // environment variables like APP__SERVER__PORT=8080
            .add_source(cfg::Environment::with_prefix("APP").separator("__"));

        let settings = builder.build()?;
        let mut cfg: AppConfig = settings.try_deserialize().unwrap_or_default();

        // Ensure defaults filled
        if cfg.server.is_none() {
            cfg.server = Some(ServerConfig { host: Some("0.0.0.0".into()), port: Some(8080) });
        }
        Ok(cfg)
    }

    pub fn bind_addr(&self) -> String {
        let host = self.server.as_ref().and_then(|s| s.host.clone()).unwrap_or_else(|| "0.0.0.0".into());
        let port = self.server.as_ref().and_then(|s| s.port).unwrap_or(8080);
        format!("{}:{}", host, port)
    }
}
