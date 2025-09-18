use serde::{Deserialize, Serialize};

// Core domain models used across services. These mirror key Go structs in
// internal/types and service layers. Fields are optional by default to tolerate
// partial payloads during incremental migration.

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
#[serde(rename_all = "snake_case")]
pub struct Tenant {
    pub id: Option<u32>,
    pub name: Option<String>,
    pub description: Option<String>,
    pub api_key: Option<String>,
    pub storage_used: Option<i64>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
#[serde(rename_all = "snake_case")]
pub struct KnowledgeBaseConfig {
    pub chunk_size: Option<i32>,
    pub chunk_overlap: Option<i32>,
    pub separators: Option<Vec<String>>, // aka split_markers in Go config
    pub enable_multimodal: Option<bool>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum KnowledgeSource {
    #[serde(rename = "file")] File,
    #[serde(rename = "url")] Url,
    #[serde(other)] Unknown,
}

impl Default for KnowledgeSource {
    fn default() -> Self { KnowledgeSource::Unknown }
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
#[serde(rename_all = "snake_case")]
pub struct KnowledgeBase {
    pub id: Option<String>,
    pub name: Option<String>,
    pub description: Option<String>,
    pub tenant_id: Option<u32>,
    pub chunking_config: Option<KnowledgeBaseConfig>,
    pub embedding_model_id: Option<String>,
    pub summary_model_id: Option<String>,
    pub rerank_model_id: Option<String>,
    pub vlm_model_id: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum ChunkType {
    #[serde(rename = "text")] Text,
    #[serde(rename = "summary")] Summary,
    #[serde(rename = "image_caption")] ImageCaption,
    #[serde(other)] Unknown,
}

impl Default for ChunkType {
    fn default() -> Self { ChunkType::Text }
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
#[serde(rename_all = "snake_case")]
pub struct Chunk {
    pub id: Option<String>,
    pub tenant_id: Option<u32>,
    pub knowledge_id: Option<String>,
    pub knowledge_base_id: Option<String>,
    pub content: Option<String>,
    pub chunk_index: Option<i32>,
    pub start_at: Option<i32>,
    pub end_at: Option<i32>,
    pub chunk_type: Option<ChunkType>,
    pub parent_chunk_id: Option<String>,
    pub pre_chunk_id: Option<String>,
    pub next_chunk_id: Option<String>,
    pub image_info: Option<String>,
    pub is_enabled: Option<bool>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum KnowledgeType {
    #[serde(rename = "file")] File,
    #[serde(rename = "url")] Url,
    #[serde(rename = "passage")] Passage,
    #[serde(other)] Unknown,
}

impl Default for KnowledgeType {
    fn default() -> Self { KnowledgeType::File }
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
#[serde(rename_all = "snake_case")]
pub struct Knowledge {
    pub id: Option<String>,
    pub tenant_id: Option<u32>,
    pub knowledge_base_id: Option<String>,
    pub r#type: Option<KnowledgeType>,
    pub title: Option<String>,
    pub description: Option<String>,
    pub file_name: Option<String>,
    pub file_path: Option<String>,
    pub file_type: Option<String>,
    pub file_size: Option<i64>,
    pub file_hash: Option<String>,
    pub parse_status: Option<String>,
    pub enable_status: Option<String>,
    pub embedding_model_id: Option<String>,
    pub metadata: Option<String>,
    pub source: Option<String>,
    pub created_at: Option<String>,
    pub updated_at: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
#[serde(rename_all = "snake_case")]
pub struct ModelParameters {
    pub base_url: Option<String>,
    pub api_key: Option<String>,
    pub embedding_parameters: Option<EmbeddingParameters>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
#[serde(rename_all = "snake_case")]
pub struct EmbeddingParameters {
    pub dimension: Option<i32>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
#[serde(rename_all = "snake_case")]
pub struct Model {
    pub id: Option<String>,
    pub tenant_id: Option<u32>,
    pub name: Option<String>,
    pub source: Option<String>,
    pub r#type: Option<String>,
    pub parameters: Option<ModelParameters>,
    pub status: Option<String>,
}
