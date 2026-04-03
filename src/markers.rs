//! Universal XML section marker resolution.
//!
//! All formats use the same `<aap:section id="...">` / `</aap:section>` markers.
//! The `aap:` namespace prefix is uniquely identifiable and LLMs follow XML tags
//! reliably. JSON uses pointer addressing instead.

use anyhow::{bail, Context, Result};

use crate::aap::SectionDef;

/// Build start and end markers for a section ID.
///
/// All text formats use the same XML-style markers:
/// `<aap:section id="nav">` / `</aap:section>`
///
/// JSON (`application/json`) does not support text markers — use pointer addressing.
pub fn markers_for(section_id: &str, format: &str) -> Result<(String, String)> {
    if format == "application/json" {
        bail!("JSON does not support text-based section markers; use pointer addressing instead");
    }
    Ok((
        format!(r#"<aap:section id="{section_id}">"#),
        "</aap:section>".to_string(),
    ))
}

/// Resolve section markers for a given section ID and format.
///
/// If the `section_def` provides explicit `start_marker` and `end_marker`,
/// those are used (user override). Otherwise, the universal XML markers are used.
pub fn resolve_markers(
    section_id: &str,
    format: &str,
    section_def: Option<&SectionDef>,
) -> Result<(String, String)> {
    // Explicit override wins
    if let Some(def) = section_def {
        if let (Some(start), Some(end)) = (&def.start_marker, &def.end_marker) {
            return Ok((start.clone(), end.clone()));
        }
    }

    markers_for(section_id, format)
}

/// Find the byte range of a section's content within a string.
///
/// Returns `(content_start, content_end)` — the byte offsets of the content
/// between the start and end markers (exclusive of markers themselves).
pub fn find_section_range(
    content: &str,
    section_id: &str,
    format: &str,
    section_def: Option<&SectionDef>,
) -> Result<(usize, usize)> {
    let (start_marker, end_marker) = resolve_markers(section_id, format, section_def)?;
    let si = content
        .find(&start_marker)
        .with_context(|| format!("start marker not found for section: {section_id}"))?;
    let ei = content
        .find(&end_marker)
        .with_context(|| format!("end marker not found for section: {section_id}"))?;
    Ok((si + start_marker.len(), ei))
}

/// Find the byte range of a section including its markers.
///
/// Returns `(marker_start, marker_end)` — the byte offsets that include both
/// the start marker, content, and end marker.
pub fn find_section_range_inclusive(
    content: &str,
    section_id: &str,
    format: &str,
    section_def: Option<&SectionDef>,
) -> Result<(usize, usize)> {
    let (start_marker, end_marker) = resolve_markers(section_id, format, section_def)?;
    let si = content
        .find(&start_marker)
        .with_context(|| format!("start marker not found for section: {section_id}"))?;
    let ei = content
        .find(&end_marker)
        .with_context(|| format!("end marker not found for section: {section_id}"))?;
    Ok((si, ei + end_marker.len()))
}

/// Look up a `SectionDef` by ID from an optional list.
pub fn find_section_def<'a>(
    sections: Option<&'a [SectionDef]>,
    section_id: &str,
) -> Option<&'a SectionDef> {
    sections?.iter().find(|s| s.id == section_id)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_html_markers() {
        let (start, end) = resolve_markers("nav", "text/html", None).unwrap();
        assert_eq!(start, r#"<aap:section id="nav">"#);
        assert_eq!(end, "</aap:section>");
    }

    #[test]
    fn test_javascript_markers() {
        let (start, end) = resolve_markers("utils", "application/javascript", None).unwrap();
        assert_eq!(start, r#"<aap:section id="utils">"#);
        assert_eq!(end, "</aap:section>");
    }

    #[test]
    fn test_python_markers() {
        let (start, end) = resolve_markers("imports", "text/x-python", None).unwrap();
        assert_eq!(start, r#"<aap:section id="imports">"#);
        assert_eq!(end, "</aap:section>");
    }

    #[test]
    fn test_json_unsupported() {
        let result = resolve_markers("data", "application/json", None);
        assert!(result.is_err());
    }

    #[test]
    fn test_explicit_override() {
        let def = SectionDef {
            id: "custom".to_string(),
            label: None,
            start_marker: Some("/* BEGIN custom */".to_string()),
            end_marker: Some("/* END custom */".to_string()),
        };
        let (start, end) = resolve_markers("custom", "application/json", Some(&def)).unwrap();
        assert_eq!(start, "/* BEGIN custom */");
        assert_eq!(end, "/* END custom */");
    }

    #[test]
    fn test_xml_format() {
        let (start, end) = resolve_markers("data", "application/xhtml+xml", None).unwrap();
        assert_eq!(start, r#"<aap:section id="data">"#);
        assert_eq!(end, "</aap:section>");
    }

    #[test]
    fn test_unknown_text_format() {
        let (start, end) = resolve_markers("block", "text/x-unknown", None).unwrap();
        assert_eq!(start, r#"<aap:section id="block">"#);
        assert_eq!(end, "</aap:section>");
    }

    #[test]
    fn test_find_section_range() {
        let content = "before\n<aap:section id=\"stats\">\nold stats\n</aap:section>\nafter";
        let (start, end) = find_section_range(content, "stats", "text/html", None).unwrap();
        assert_eq!(&content[start..end], "\nold stats\n");
    }

    #[test]
    fn test_find_section_range_python() {
        let content = "import os\n<aap:section id=\"imports\">\nimport sys\n</aap:section>\ncode";
        let (start, end) = find_section_range(content, "imports", "text/x-python", None).unwrap();
        assert_eq!(&content[start..end], "\nimport sys\n");
    }
}
