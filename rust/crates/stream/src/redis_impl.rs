use crate::{StreamInfo, StreamManager};
use async_trait::async_trait;
use chrono::Utc;
use redis::{AsyncCommands, Client};
use std::time::Duration;
use tokio::time::sleep;

#[derive(Clone)]
pub struct RedisStreamManager {
    client: Client,
    ttl: Duration,
    prefix: String,
}

impl RedisStreamManager {
    pub async fn new(redis_addr: &str, redis_password: Option<&str>, redis_db: Option<i64>, prefix: Option<&str>, ttl: Option<Duration>) -> anyhow::Result<Self> {
        // Build connection URL: redis://[:password@]host:port/db
        let mut url = if let Some(pw) = redis_password { format!("redis://:{}@{}", pw, redis_addr) } else { format!("redis://{}", redis_addr) };
        if let Some(db) = redis_db { url = format!("{}/{}", url.trim_end_matches('/'), db); }
        let client = Client::open(url)?;

        // Ping to validate connection
        let mut conn = client.get_async_connection().await?;
        let: () = redis::cmd("PING").query_async(&mut conn).await?;

        let ttl = ttl.unwrap_or_else(|| Duration::from_secs(24 * 3600));
        let prefix = prefix.map(str::to_string).unwrap_or_else(|| "stream:".to_string());

        Ok(Self { client, ttl, prefix })
    }

    fn build_key(&self, session_id: &str, request_id: &str) -> String {
        format!("{}:{}:{}", self.prefix, session_id, request_id)
    }
}

#[async_trait]
impl StreamManager for RedisStreamManager {
    async fn register_stream(&self, session_id: &str, request_id: &str, query: &str) -> anyhow::Result<()> {
        let mut conn = self.client.get_async_connection().await?;
        let info = StreamInfo {
            session_id: session_id.to_string(),
            request_id: request_id.to_string(),
            query: query.to_string(),
            content: String::new(),
            knowledge_references: None,
            last_updated: Utc::now(),
            is_completed: false,
        };
        let key = self.build_key(session_id, request_id);
        let data = serde_json::to_vec(&info)?;
        let _: () = redis::cmd("SET").arg(&key).arg(data).arg("EX").arg(self.ttl.as_secs()).query_async(&mut conn).await?;
        Ok(())
    }

    async fn update_stream(&self, session_id: &str, request_id: &str, content: &str, references_json: Option<&str>) -> anyhow::Result<()> {
        let mut conn = self.client.get_async_connection().await?;
        let key = self.build_key(session_id, request_id);
        let data: Option<Vec<u8>> = conn.get(&key).await?;
        if data.is_none() { return Ok(()); }
        let mut info: StreamInfo = serde_json::from_slice(&data.unwrap())?;
        info.content.push_str(content);
        if let Some(r) = references_json { if !r.is_empty() { info.knowledge_references = Some(r.to_string()); } }
        info.last_updated = Utc::now();
        let new_data = serde_json::to_vec(&info)?;
        let _: () = redis::cmd("SET").arg(&key).arg(new_data).arg("EX").arg(self.ttl.as_secs()).query_async(&mut conn).await?;
        Ok(())
    }

    async fn complete_stream(&self, session_id: &str, request_id: &str) -> anyhow::Result<()> {
        let mut conn = self.client.get_async_connection().await?;
        let key = self.build_key(session_id, request_id);
        let data: Option<Vec<u8>> = conn.get(&key).await?;
        if data.is_none() { return Ok(()); }
        let mut info: StreamInfo = serde_json::from_slice(&data.unwrap())?;
        info.is_completed = true;
        info.last_updated = Utc::now();
        let new_data = serde_json::to_vec(&info)?;
        let _: () = redis::cmd("SET").arg(&key).arg(new_data).arg("EX").arg(self.ttl.as_secs()).query_async(&mut conn).await?;

        // schedule deletion after 30s (similar to Go)
        let client = self.client.clone();
        let key_s = key.clone();
        tokio::spawn(async move {
            sleep(Duration::from_secs(30)).await;
            if let Ok(mut c) = client.get_async_connection().await {
                let _ : Result<(), _> = async {
                    let _: () = redis::cmd("DEL").arg(&key_s).query_async(&mut c).await?;
                    Ok(())
                }.await;
            }
        });
        Ok(())
    }

    async fn get_stream(&self, session_id: &str, request_id: &str) -> anyhow::Result<Option<StreamInfo>> {
        let mut conn = self.client.get_async_connection().await?;
        let key = self.build_key(session_id, request_id);
        let data: Option<Vec<u8>> = conn.get(&key).await?;
        if let Some(raw) = data { Ok(Some(serde_json::from_slice(&raw)?)) } else { Ok(None) }
    }
}
