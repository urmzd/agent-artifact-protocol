//! Agent-Artifact Protocol (AAP) data model — Rust implementation of aap/0.1.
//!
//! Four envelope types: `synthesize` (full generation), `edit` (targeted changes),
//! `handle` (lightweight reference), `handle_result` (response from handle interaction).

use serde::{Deserialize, Serialize};

pub const PROTOCOL_VERSION: &str = "aap/0.1";

/// Operation name — four types: synthesize, edit, handle, handle_result.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "snake_case")]
pub enum Name {
    Synthesize,
    Edit,
    Handle,
    HandleResult,
}

/// Artifact lifecycle state.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum ArtifactState {
    Draft,
    Published,
    Archived,
}

/// Token budget constraints.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TokenBudget {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub max_tokens: Option<u64>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub priority: Option<String>,
}

/// Operation metadata object.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Operation {
    pub direction: String,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub format: Option<String>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub encoding: Option<String>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub content_encoding: Option<String>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub token_budget: Option<TokenBudget>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub tokens_used: Option<u64>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub checksum: Option<String>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub created_at: Option<String>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub updated_at: Option<String>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub state: Option<ArtifactState>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub state_changed_at: Option<String>,
}

/// Top-level envelope.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Envelope {
    pub protocol: String,
    pub id: String,
    pub version: u64,
    pub name: Name,
    pub operation: Operation,
    pub content: Vec<serde_json::Value>,
}

impl Envelope {
    pub fn is_envelope(s: &str) -> bool {
        let trimmed = s.trim_start();
        trimmed.starts_with('{') && trimmed.contains("\"aap/")
    }

    pub fn from_json(s: &str) -> Result<Self, serde_json::Error> {
        serde_json::from_str(s)
    }
}

/// Target definition — metadata about a named target in the artifact.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TargetDef {
    pub id: String,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub label: Option<String>,
}

/// Content item for `name: "synthesize"`.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SynthesizeContentItem {
    pub body: String,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub targets: Option<Vec<TargetDef>>,
}

/// Target addressing for edit operations — discriminated union on `type`.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type", content = "value")]
pub enum Target {
    /// Target an `<aap:target id="...">` marker by ID.
    #[serde(rename = "id")]
    Id(String),

    /// Target a value by JSON Pointer (RFC 6901).
    #[serde(rename = "pointer")]
    Pointer(String),
}

/// Edit operation type.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "snake_case")]
pub enum OpType {
    Replace,
    InsertBefore,
    InsertAfter,
    Delete,
}

/// A single edit operation (content item for `name: "edit"`).
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DiffOp {
    pub op: OpType,
    pub target: Target,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub content: Option<String>,
}

/// Content item for `name: "handle"`.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HandleContentItem {
    pub sections: Vec<String>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub token_count: Option<u64>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub state: Option<ArtifactState>,
}

/// Result status for edit operations (used in handle_result).
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum ResultStatus {
    Applied,
    Rejected,
    Partial,
    Conflict,
}

/// Description of a single change within a handle_result.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChangeDescription {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub target_id: Option<String>,
    pub description: String,
}

/// Content item for `name: "handle_result"` — discriminated union on `type`.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub enum HandleResultContentItem {
    /// Free-form text response (answers, descriptions).
    #[serde(rename = "text")]
    Text { body: String },

    /// Edit confirmation with status and change descriptions.
    #[serde(rename = "edit")]
    Edit {
        status: ResultStatus,
        changes: Vec<ChangeDescription>,
    },

    /// Error or rejection response.
    #[serde(rename = "error")]
    Error { code: String, message: String },
}
