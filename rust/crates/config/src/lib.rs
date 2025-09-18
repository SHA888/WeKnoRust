use config as cfg;
use dotenvy::dotenv;
use serde::{Deserialize, Serialize};
use thiserror::Error;

#[derive(Debug, Error)]
pub enum ConfigError {
    #[error("config build error: {0}")]
    Build(#[from] cfg::ConfigError),
}

// ----- Typed config structures mirroring Go internal/config/config.go -----

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct SummaryConfig {
    #[serde(default)]
    pub max_tokens: i32,
    #[serde(default)]
    pub repeat_penalty: f64,
    #[serde(default)]
    pub top_k: i32,
    #[serde(default)]
    pub top_p: f64,
    #[serde(default)]
    pub frequency_penalty: f64,
    #[serde(default)]
    pub presence_penalty: f64,
    #[serde(default)]
    pub prompt: String,
    #[serde(default)]
    pub context_template: String,
    #[serde(default)]
    pub temperature: f64,
    #[serde(default)]
    pub seed: i32,
    #[serde(default)]
    pub max_completion_tokens: i32,
    #[serde(default)]
    pub no_match_prefix: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct ConversationConfig {
    #[serde(default)]
    pub max_rounds: i32,
    #[serde(default)]
    pub keyword_threshold: f64,
    #[serde(default)]
    pub embedding_top_k: i32,
    #[serde(default)]
    pub vector_threshold: f64,
    #[serde(default)]
    pub rerank_top_k: i32,
    #[serde(default)]
    pub rerank_threshold: f64,
    #[serde(default)]
    pub fallback_strategy: String,
    #[serde(default)]
    pub fallback_response: String,
    #[serde(default)]
    pub fallback_prompt: String,
    #[serde(default)]
    pub enable_rewrite: bool,
    #[serde(default)]
    pub enable_rerank: bool,
    #[serde(default)]
    pub summary: Option<SummaryConfig>,
    #[serde(default)]
    pub generate_session_title_prompt: String,
    #[serde(default)]
    pub generate_summary_prompt: String,
    #[serde(default)]
    pub rewrite_prompt_system: String,
    #[serde(default)]
    pub rewrite_prompt_user: String,
    #[serde(default)]
    pub simplify_query_prompt: String,
    #[serde(default)]
    pub simplify_query_prompt_user: String,
    #[serde(default)]
    pub extract_entities_prompt: String,
    #[serde(default)]
    pub extract_relationships_prompt: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ServerConfig {
    #[serde(default = "default_host")] pub host: String,
    #[serde(default = "default_port")] pub port: u16,
    #[serde(default)] pub log_path: String,
    #[serde(default = "default_shutdown")] pub shutdown_timeout: String,
}

fn default_host() -> String { "0.0.0.0".into() }
fn default_port() -> u16 { 8080 }
fn default_shutdown() -> String { "30s".into() }

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct ImageProcessingConfig {
    #[serde(default)]
    pub enable_multimodal: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct KnowledgeBaseConfig {
    #[serde(default)]
    pub chunk_size: i32,
    #[serde(default)]
    pub chunk_overlap: i32,
    #[serde(default)]
    pub split_markers: Vec<String>,
    #[serde(default)]
    pub keep_separator: bool,
    #[serde(default)]
    pub image_processing: Option<ImageProcessingConfig>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct TenantConfig {
    #[serde(default)] pub default_session_name: String,
    #[serde(default)] pub default_session_title: String,
    #[serde(default)] pub default_session_description: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct ModelConfig {
    #[serde(default)] pub r#type: String,
    #[serde(default)] pub source: String,
    #[serde(default)] pub model_name: String,
    #[serde(default)] pub parameters: serde_json::Value,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct AsynqConfig {
    #[serde(default)] pub addr: String,
    #[serde(default)] pub username: String,
    #[serde(default)] pub password: String,
    #[serde(default)] pub read_timeout: String,
    #[serde(default)] pub write_timeout: String,
    #[serde(default)] pub concurrency: i32,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct VectorDatabaseConfig {
    #[serde(default)] pub driver: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct DocReaderConfig {
    #[serde(default)] pub addr: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct RedisConfig {
    #[serde(default)] pub address: String,
    #[serde(default)] pub password: String,
    #[serde(default)] pub db: i32,
    #[serde(default)] pub prefix: String,
    #[serde(default)] pub ttl: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct StreamManagerConfig {
    #[serde(default)] pub r#type: String,
    #[serde(default)] pub redis: RedisConfig,
    #[serde(default)] pub cleanup_timeout: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct AppConfig {
    #[serde(default)] pub conversation: Option<ConversationConfig>,
    #[serde(default)] pub server: Option<ServerConfig>,
    #[serde(default)] pub knowledge_base: Option<KnowledgeBaseConfig>,
    #[serde(default)] pub tenant: Option<TenantConfig>,
    #[serde(default)] pub models: Vec<ModelConfig>,
    #[serde(default)] pub asynq: Option<AsynqConfig>,
    #[serde(default)] pub vector_database: Option<VectorDatabaseConfig>,
    #[serde(default)] pub docreader: Option<DocReaderConfig>,
    #[serde(default)] pub stream_manager: Option<StreamManagerConfig>,
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

        // Defaults
        if cfg.server.is_none() {
            cfg.server = Some(ServerConfig { host: default_host(), port: default_port(), log_path: String::new(), shutdown_timeout: default_shutdown() });
        }
        Ok(cfg)
    }

    pub fn bind_addr(&self) -> String {
        let server = self.server.as_ref();
        let host = server.map(|s| s.host.clone()).unwrap_or_else(default_host);
        let port = server.map(|s| s.port).unwrap_or_else(default_port);
        format!("{}:{}", host, port)
    }
}
