//! Stateless apply engine — pure function that transforms artifacts.
//!
//! `apply(artifact?, envelope) → (Artifact, Handle)`
//!
//! Synthesize: creates artifact from body. Edit: applies ops to existing artifact.
//! Both return the artifact and a handle envelope.

use anyhow::{bail, Context, Result};

use crate::aap::{
    Artifact, EditOp, Envelope, HandleContentItem, Name, OpType, Meta, SynthesizeContentItem,
    Target, PROTOCOL_VERSION,
};

// ── Resolve trait ────────────────────────────────────────────────────────

/// Content resolution — how to find and replace targeted regions.
pub trait Resolve {
    type Content: Clone;

    fn find_by_id(&self, content: &Self::Content, id: &str) -> Result<(usize, usize)>;
    fn find_by_pointer(&self, content: &Self::Content, pointer: &str) -> Result<(usize, usize)>;
    fn replace(&self, content: &mut Self::Content, start: usize, end: usize, replacement: &str);
    fn insert(&self, content: &mut Self::Content, pos: usize, text: &str);
    fn delete(&self, content: &mut Self::Content, start: usize, end: usize);
    fn to_string(&self, content: &Self::Content) -> String;
    fn from_string(&self, s: &str) -> Self::Content;
}

// ── Text resolver ────────────────────────────────────────────────────────

/// Text-based resolver using `<aap:target id="...">` markers.
pub struct TextResolver {
    pub format: String,
}

impl Resolve for TextResolver {
    type Content = String;

    fn find_by_id(&self, content: &String, id: &str) -> Result<(usize, usize)> {
        crate::markers::find_target_range(content, id, &self.format)
    }

    fn find_by_pointer(&self, content: &String, pointer: &str) -> Result<(usize, usize)> {
        let value: serde_json::Value = serde_json::from_str(content)
            .context("pointer targeting requires valid JSON content")?;
        let serialized = serde_json::to_string_pretty(&value)?;
        let _ = value
            .pointer(pointer)
            .with_context(|| format!("pointer not found: {pointer}"))?;
        Ok((0, serialized.len()))
    }

    fn replace(&self, content: &mut String, start: usize, end: usize, replacement: &str) {
        *content = format!("{}{}{}", &content[..start], replacement, &content[end..]);
    }

    fn insert(&self, content: &mut String, pos: usize, text: &str) {
        *content = format!("{}{}{}", &content[..pos], text, &content[pos..]);
    }

    fn delete(&self, content: &mut String, start: usize, end: usize) {
        *content = format!("{}{}", &content[..start], &content[end..]);
    }

    fn to_string(&self, content: &String) -> String {
        content.clone()
    }

    fn from_string(&self, s: &str) -> String {
        s.to_string()
    }
}

// ── Apply engine ─────────────────────────────────────────────────────────

fn extract_synthesize_item(envelope: &Envelope) -> Result<SynthesizeContentItem> {
    serde_json::from_value(
        envelope
            .content
            .first()
            .context("synthesize: empty content array")?
            .clone(),
    )
    .context("synthesize: failed to parse content item")
}

fn build_handle_envelope(artifact: &Artifact) -> Result<Envelope> {
    let handle = HandleContentItem {
        id: artifact.id.clone(),
        version: artifact.version,
        token_count: Some(artifact.body.len() as u64 / 4), // rough estimate
        state: None,
        content: None,
    };
    Ok(Envelope {
        protocol: PROTOCOL_VERSION.to_string(),
        id: artifact.id.clone(),
        version: artifact.version,
        name: Name::Handle,
        meta: Meta {
            format: Some(artifact.format.clone()),
            tokens_used: None,
            checksum: None,
            state: None,
        },
        content: vec![
            serde_json::to_value(handle).context("failed to serialize handle")?
        ],
    })
}

/// Stateless apply: `f(artifact?, envelope) → (Artifact, Handle)`.
pub fn apply(artifact: Option<&Artifact>, envelope: &Envelope) -> Result<(Artifact, Envelope)> {
    let format = envelope
        .meta
        .format
        .as_deref()
        .unwrap_or("text/html");

    let resolver = TextResolver {
        format: format.to_string(),
    };

    let result_artifact = match envelope.name {
        Name::Synthesize => {
            let item = extract_synthesize_item(envelope)?;
            Artifact {
                id: envelope.id.clone(),
                version: envelope.version,
                format: format.to_string(),
                body: item.body,
            }
        }
        Name::Edit => {
            let art = artifact.context("edit requires a base artifact")?;
            let ops: Vec<EditOp> = envelope
                .content
                .iter()
                .map(|v| serde_json::from_value(v.clone()))
                .collect::<std::result::Result<Vec<_>, _>>()
                .context("edit: failed to parse content items")?;

            let has_pointer = ops.iter().any(|op| matches!(op.target, Target::Pointer(_)));
            let body = if has_pointer {
                apply_edit_pointers(&art.body, &ops)?
            } else {
                apply_edit(&resolver, &art.body, &ops)?
            };

            Artifact {
                id: envelope.id.clone(),
                version: envelope.version,
                format: format.to_string(),
                body,
            }
        }
        Name::Handle => {
            bail!("handle is an output envelope, not an input operation")
        }
    };

    let handle = build_handle_envelope(&result_artifact)?;
    Ok((result_artifact, handle))
}

/// Apply edit operations using the Resolve trait (ID-based targeting).
pub fn apply_edit<R: Resolve<Content = String>>(
    resolver: &R,
    base: &str,
    operations: &[EditOp],
) -> Result<String> {
    let mut content = resolver.from_string(base);

    for (i, op) in operations.iter().enumerate() {
        let (start, end) = resolve_target(resolver, &content, &op.target)
            .with_context(|| format!("operation {i}: target not found"))?;

        match op.op {
            OpType::Replace => {
                let replacement = op.content.as_deref().unwrap_or("");
                resolver.replace(&mut content, start, end, replacement);
            }
            OpType::Delete => {
                resolver.delete(&mut content, start, end);
            }
            OpType::InsertBefore => {
                let text = op.content.as_deref().unwrap_or("");
                resolver.insert(&mut content, start, text);
            }
            OpType::InsertAfter => {
                let text = op.content.as_deref().unwrap_or("");
                resolver.insert(&mut content, end, text);
            }
        }
    }

    Ok(resolver.to_string(&content))
}

fn resolve_target<R: Resolve<Content = String>>(
    resolver: &R,
    content: &String,
    target: &Target,
) -> Result<(usize, usize)> {
    match target {
        Target::Id(id) => resolver.find_by_id(content, id),
        Target::Pointer(pointer) => resolver.find_by_pointer(content, pointer),
    }
}

/// Apply edit operations using JSON Pointer targeting.
fn apply_edit_pointers(base: &str, operations: &[EditOp]) -> Result<String> {
    let mut value: serde_json::Value =
        serde_json::from_str(base).context("pointer targeting requires valid JSON content")?;

    for (i, op) in operations.iter().enumerate() {
        let pointer = match &op.target {
            Target::Pointer(p) => p.as_str(),
            _ => bail!("operation {i}: expected pointer target"),
        };

        match op.op {
            OpType::Replace => {
                let content = op.content.as_deref().context("replace requires content")?;
                let new_val: serde_json::Value =
                    serde_json::from_str(content).context("content must be valid JSON")?;
                let target = value
                    .pointer_mut(pointer)
                    .with_context(|| format!("pointer not found: {pointer}"))?;
                *target = new_val;
            }
            OpType::Delete => {
                let (parent_ptr, key) = split_pointer(pointer).context("cannot delete root")?;
                let parent = value
                    .pointer_mut(&parent_ptr)
                    .with_context(|| format!("parent not found: {parent_ptr}"))?;
                remove_child(parent, &key)?;
            }
            OpType::InsertBefore | OpType::InsertAfter => {
                let content = op.content.as_deref().context("insert requires content")?;
                let new_val: serde_json::Value =
                    serde_json::from_str(content).context("content must be valid JSON")?;
                let (parent_ptr, key) = split_pointer(pointer).context("cannot insert at root")?;
                let parent = value
                    .pointer_mut(&parent_ptr)
                    .with_context(|| format!("parent not found: {parent_ptr}"))?;
                let arr = parent
                    .as_array_mut()
                    .context("insert requires array parent")?;
                let index: usize = key.parse().context("insert requires numeric array index")?;
                let insert_at = if op.op == OpType::InsertAfter { index + 1 } else { index };
                arr.insert(insert_at, new_val);
            }
        }
    }

    serde_json::to_string_pretty(&value).context("failed to re-serialize JSON")
}

fn split_pointer(pointer: &str) -> Result<(String, String)> {
    if pointer.is_empty() || !pointer.starts_with('/') {
        bail!("invalid JSON Pointer: {pointer:?}");
    }
    match pointer.rfind('/') {
        Some(0) => Ok(("".to_string(), pointer[1..].to_string())),
        Some(pos) => Ok((pointer[..pos].to_string(), pointer[pos + 1..].to_string())),
        None => bail!("invalid JSON Pointer: {pointer:?}"),
    }
}

fn remove_child(parent: &mut serde_json::Value, key: &str) -> Result<()> {
    let unescaped = key.replace("~1", "/").replace("~0", "~");
    if let Some(obj) = parent.as_object_mut() {
        if obj.remove(&unescaped).is_none() {
            bail!("key not found: {unescaped}");
        }
    } else if let Some(arr) = parent.as_array_mut() {
        let index: usize = unescaped.parse().with_context(|| format!("expected array index: {unescaped}"))?;
        if index >= arr.len() { bail!("array index out of bounds: {index}"); }
        arr.remove(index);
    } else {
        bail!("parent is neither object nor array");
    }
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    fn synth_env(id: &str, version: u64, body: &str) -> Envelope {
        Envelope {
            protocol: PROTOCOL_VERSION.to_string(),
            id: id.to_string(),
            version,
            name: Name::Synthesize,
            meta: Meta {
                    format: Some("text/html".to_string()),
                tokens_used: None, checksum: None, state: None,
            },
            content: vec![serde_json::json!({ "body": body })],
        }
    }

    fn edit_env(id: &str, version: u64, ops: Vec<EditOp>) -> Envelope {
        Envelope {
            protocol: PROTOCOL_VERSION.to_string(),
            id: id.to_string(),
            version,
            name: Name::Edit,
            meta: Meta {
                    format: Some("text/html".to_string()),
                tokens_used: None, checksum: None, state: None,
            },
            content: ops.iter().map(|o| serde_json::to_value(o).unwrap()).collect(),
        }
    }

    fn id_target(id: &str) -> Target { Target::Id(id.to_string()) }
    fn ptr_target(p: &str) -> Target { Target::Pointer(p.to_string()) }

    #[test]
    fn test_synthesize() {
        let env = synth_env("test", 1, "<div>hello</div>");
        let (art, handle) = apply(None, &env).unwrap();
        assert_eq!(art.body, "<div>hello</div>");
        assert_eq!(art.id, "test");
        assert_eq!(art.version, 1);
        assert_eq!(handle.name, Name::Handle);
    }

    #[test]
    fn test_edit_replace_by_id() {
        let env = synth_env("t", 1, r#"<aap:target id="rev">$12,340</aap:target>"#);
        let (art, _) = apply(None, &env).unwrap();

        let edit = edit_env("t", 2, vec![EditOp {
            op: OpType::Replace,
            target: id_target("rev"),
            content: Some("$15,720".to_string()),
        }]);
        let (art2, _) = apply(Some(&art), &edit).unwrap();
        assert!(art2.body.contains("$15,720"));
        assert!(!art2.body.contains("$12,340"));
        assert!(art2.body.contains(r#"<aap:target id="rev">"#));
    }

    #[test]
    fn test_edit_delete_by_id() {
        let env = synth_env("t", 1, r#"before<aap:target id="tmp">remove</aap:target>after"#);
        let (art, _) = apply(None, &env).unwrap();

        let edit = edit_env("t", 2, vec![EditOp {
            op: OpType::Delete, target: id_target("tmp"), content: None,
        }]);
        let (art2, _) = apply(Some(&art), &edit).unwrap();
        assert_eq!(art2.body, r#"before<aap:target id="tmp"></aap:target>after"#);
    }

    #[test]
    fn test_edit_insert_after() {
        let env = synth_env("t", 1, r#"<aap:target id="list">item1</aap:target>"#);
        let (art, _) = apply(None, &env).unwrap();

        let edit = edit_env("t", 2, vec![EditOp {
            op: OpType::InsertAfter, target: id_target("list"),
            content: Some(", item2".to_string()),
        }]);
        let (art2, _) = apply(Some(&art), &edit).unwrap();
        assert!(art2.body.contains("item1, item2"));
    }

    #[test]
    fn test_nested_targets() {
        let body = r#"<aap:target id="outer"><h2>Stats</h2><aap:target id="val">100</aap:target></aap:target>"#;
        let (art, _) = apply(None, &synth_env("t", 1, body)).unwrap();

        let edit = edit_env("t", 2, vec![EditOp {
            op: OpType::Replace, target: id_target("val"),
            content: Some("200".to_string()),
        }]);
        let (art2, _) = apply(Some(&art), &edit).unwrap();
        assert!(art2.body.contains("200"));
        assert!(art2.body.contains("<h2>Stats</h2>"));
    }

    #[test]
    fn test_target_serde_roundtrip() {
        let t = Target::Id("revenue".to_string());
        let json = serde_json::to_string(&t).unwrap();
        let parsed: Target = serde_json::from_str(&json).unwrap();
        assert!(matches!(parsed, Target::Id(ref s) if s == "revenue"));

        let op = EditOp {
            op: OpType::Replace,
            target: Target::Id("rev".to_string()),
            content: Some("new".to_string()),
        };
        let json = serde_json::to_string(&op).unwrap();
        let parsed: EditOp = serde_json::from_str(&json).unwrap();
        assert!(matches!(parsed.target, Target::Id(ref s) if s == "rev"));
    }

    #[test]
    fn test_edit_from_json_string() {
        let json = r#"{
            "protocol": "aap/0.1", "id": "x", "version": 2, "name": "edit",
            "meta": {"format": "text/html"},
            "content": [{"op": "replace", "target": {"type": "id", "value": "rev"}, "content": "new"}]
        }"#;
        let env: Envelope = serde_json::from_str(json).unwrap();
        let art_body = r#"<aap:target id="rev">old</aap:target>"#;
        let (art, _) = apply(None, &synth_env("x", 1, art_body)).unwrap();
        let (art2, _) = apply(Some(&art), &env).unwrap();
        assert!(art2.body.contains("new"));
        assert!(!art2.body.contains("old"));
    }

    #[test]
    fn test_pointer_replace() {
        let base = r#"{"name": "Alice", "age": 30}"#;
        let (art, _) = apply(None, &synth_env("t", 1, base)).unwrap();

        let mut edit = edit_env("t", 2, vec![EditOp {
            op: OpType::Replace, target: ptr_target("/name"),
            content: Some(r#""Bob""#.to_string()),
        }]);
        edit.meta.format = Some("application/json".to_string());
        let (art2, _) = apply(Some(&art), &edit).unwrap();
        let parsed: serde_json::Value = serde_json::from_str(&art2.body).unwrap();
        assert_eq!(parsed["name"], "Bob");
        assert_eq!(parsed["age"], 30);
    }

    #[test]
    fn test_pointer_delete() {
        let base = r#"{"name": "Alice", "temp": true}"#;
        let (art, _) = apply(None, &synth_env("t", 1, base)).unwrap();

        let mut edit = edit_env("t", 2, vec![EditOp {
            op: OpType::Delete, target: ptr_target("/temp"), content: None,
        }]);
        edit.meta.format = Some("application/json".to_string());
        let (art2, _) = apply(Some(&art), &edit).unwrap();
        let parsed: serde_json::Value = serde_json::from_str(&art2.body).unwrap();
        assert!(parsed.get("temp").is_none());
    }

    #[test]
    fn test_handle_is_not_input() {
        let env = Envelope {
            protocol: PROTOCOL_VERSION.to_string(),
            id: "t".to_string(), version: 1, name: Name::Handle,
            meta: Meta {
                format: None,
                tokens_used: None, checksum: None, state: None,
            },
            content: vec![],
        };
        assert!(apply(None, &env).is_err());
    }
}
