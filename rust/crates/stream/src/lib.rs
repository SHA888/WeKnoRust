pub mod memory;
pub mod redis_impl;

use async_trait::async_trait;
use serde::{Deserialize, Serialize};
use chrono::{DateTime, Utc};

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct StreamInfo {
    pub session_id: String,
    pub request_id: String,
    pub query: String,
    pub content: String,
    pub knowledge_references: Option<String>, // JSON string for simplicity; can be typed later
    pub last_updated: DateTime<Utc>,
    pub is_completed: bool,
}

#[async_trait]
pub trait StreamManager: Send + Sync {
    async fn register_stream(&self, session_id: &str, request_id: &str, query: &str) -> anyhow::Result<()>;
    async fn update_stream(&self, session_id: &str, request_id: &str, content: &str, references_json: Option<&str>) -> anyhow::Result<()>;
    async fn complete_stream(&self, session_id: &str, request_id: &str) -> anyhow::Result<()>;
    async fn get_stream(&self, session_id: &str, request_id: &str) -> anyhow::Result<Option<StreamInfo>>;
}
