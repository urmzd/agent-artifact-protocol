//! Universal XML marker resolution for `<aap:target>`.
//!
//! All formats use `<aap:target id="...">` / `</aap:target>` markers.
//! The `aap:` namespace prefix is uniquely identifiable and LLMs follow XML tags
//! reliably. JSON uses pointer addressing instead.

use anyhow::{bail, Context, Result};
use regex::Regex;

/// Build start and end markers for a target ID.
///
/// `<aap:target id="nav">` / `</aap:target>`
///
/// JSON (`application/json`) does not support text markers — use pointer addressing.
pub fn markers_for(target_id: &str, format: &str) -> Result<(String, String)> {
    if format == "application/json" {
        bail!("JSON does not support text-based markers; use pointer addressing instead");
    }
    Ok((
        format!(r#"<aap:target id="{target_id}">"#),
        "</aap:target>".to_string(),
    ))
}

/// Find the byte range of a target's content within a string.
///
/// Returns `(content_start, content_end)` — byte offsets between markers (exclusive of markers).
pub fn find_target_range(
    content: &str,
    target_id: &str,
    format: &str,
) -> Result<(usize, usize)> {
    let (start_marker, end_marker) = markers_for(target_id, format)?;
    let si = content
        .find(&start_marker)
        .with_context(|| format!("start marker not found for target: {target_id}"))?;
    let content_start = si + start_marker.len();
    let ei = content[content_start..]
        .find(&end_marker)
        .map(|i| content_start + i)
        .with_context(|| format!("end marker not found for target: {target_id}"))?;
    Ok((content_start, ei))
}

/// Find the byte range of a target including its markers.
///
/// Returns `(marker_start, marker_end)` — byte offsets including both markers and content.
pub fn find_target_range_inclusive(
    content: &str,
    target_id: &str,
    format: &str,
) -> Result<(usize, usize)> {
    let (start_marker, end_marker) = markers_for(target_id, format)?;
    let si = content
        .find(&start_marker)
        .with_context(|| format!("start marker not found for target: {target_id}"))?;
    let ei = content[si..]
        .find(&end_marker)
        .map(|i| si + i + end_marker.len())
        .with_context(|| format!("end marker not found for target: {target_id}"))?;
    Ok((si, ei))
}

/// Extract all target IDs from artifact content by scanning for `<aap:target id="...">` markers.
///
/// Returns IDs in document order. JSON format returns an empty vec (uses pointer addressing).
pub fn extract_targets(content: &str, format: &str) -> Vec<String> {
    if format == "application/json" {
        return Vec::new();
    }
    let re = Regex::new(r#"<aap:target id="([^"]+)">"#).expect("valid regex");
    re.captures_iter(content)
        .map(|cap| cap[1].to_string())
        .collect()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_html_markers() {
        let (start, end) = markers_for("nav", "text/html").unwrap();
        assert_eq!(start, r#"<aap:target id="nav">"#);
        assert_eq!(end, "</aap:target>");
    }

    #[test]
    fn test_python_markers() {
        let (start, end) = markers_for("imports", "text/x-python").unwrap();
        assert_eq!(start, r#"<aap:target id="imports">"#);
        assert_eq!(end, "</aap:target>");
    }

    #[test]
    fn test_json_unsupported() {
        assert!(markers_for("data", "application/json").is_err());
    }

    #[test]
    fn test_find_target_range() {
        let content = r#"before<aap:target id="stats">old stats</aap:target>after"#;
        let (start, end) = find_target_range(content, "stats", "text/html").unwrap();
        assert_eq!(&content[start..end], "old stats");
    }

    #[test]
    fn test_find_target_range_nested() {
        let content = r#"<aap:target id="outer"><aap:target id="inner">val</aap:target></aap:target>"#;
        let (start, end) = find_target_range(content, "inner", "text/html").unwrap();
        assert_eq!(&content[start..end], "val");
    }

    #[test]
    fn test_find_target_range_inclusive() {
        let content = r#"before<aap:target id="x">data</aap:target>after"#;
        let (start, end) = find_target_range_inclusive(content, "x", "text/html").unwrap();
        assert_eq!(&content[start..end], r#"<aap:target id="x">data</aap:target>"#);
    }

    #[test]
    fn test_extract_targets_flat() {
        let content = r#"<aap:target id="a">x</aap:target><aap:target id="b">y</aap:target>"#;
        assert_eq!(extract_targets(content, "text/html"), vec!["a", "b"]);
    }

    #[test]
    fn test_extract_targets_nested() {
        let content = r#"<aap:target id="outer"><aap:target id="inner">v</aap:target></aap:target>"#;
        assert_eq!(extract_targets(content, "text/html"), vec!["outer", "inner"]);
    }

    #[test]
    fn test_extract_targets_empty() {
        assert!(extract_targets("no markers here", "text/html").is_empty());
    }

    #[test]
    fn test_extract_targets_json_returns_empty() {
        let content = r#"<aap:target id="a">x</aap:target>"#;
        assert!(extract_targets(content, "application/json").is_empty());
    }
}
