use std::{collections::HashMap};
use tokio::sync::RwLock;
use chrono::{DateTime, Utc};
use std::sync::Arc;
use crate::{StreamInfo, StreamManager};
use async_trait::async_trait;

#[derive(Clone, Default)]
pub struct MemoryStreamManager {
    // session_id -> request_id -> info
    inner: Arc<RwLock<HashMap<String, HashMap<String, StreamInfo>>>>,
}

impl MemoryStreamManager {
    pub fn new() -> Self { Self { inner: Default::default() } }
}

#[async_trait]
impl StreamManager for MemoryStreamManager {
    async fn register_stream(&self, session_id: &str, request_id: &str, query: &str) -> anyhow::Result<()> {
        let mut guard = self.inner.write().await;
        let reqs = guard.entry(session_id.to_string()).or_default();
        reqs.insert(request_id.to_string(), StreamInfo {
            session_id: session_id.to_string(),
            request_id: request_id.to_string(),
            query: query.to_string(),
            content: String::new(),
            knowledge_references: None,
            last_updated: Utc::now(),
            is_completed: false,
        });
        Ok(())
    }

    async fn update_stream(&self, session_id: &str, request_id: &str, content: &str, references_json: Option<&str>) -> anyhow::Result<()> {
        let mut guard = self.inner.write().await;
        if let Some(reqs) = guard.get_mut(session_id) {
            if let Some(info) = reqs.get_mut(request_id) {
                info.content.push_str(content);
                if let Some(r) = references_json { if !r.is_empty() { info.knowledge_references = Some(r.to_string()); } }
                info.last_updated = Utc::now();
            }
        }
        Ok(())
    }

    async fn complete_stream(&self, session_id: &str, request_id: &str) -> anyhow::Result<()> {
        let mut guard = self.inner.write().await;
        if let Some(reqs) = guard.get_mut(session_id) {
            if let Some(info) = reqs.get_mut(request_id) {
                info.is_completed = true;
            }
        }
        Ok(())
    }

    async fn get_stream(&self, session_id: &str, request_id: &str) -> anyhow::Result<Option<StreamInfo>> {
        let guard = self.inner.read().await;
        Ok(guard.get(session_id).and_then(|m| m.get(request_id)).cloned())
    }
}
