//! C-compatible FFI for the GAP apply engine.
//!
//! Exposes `gap_resolve_envelope` for use via CGo (or any C caller).
//! Returns heap-allocated JSON strings; caller must free with `gap_free_string`.

use std::ffi::{CStr, CString};
use std::os::raw::c_char;

use crate::apply;
use crate::gap::{Artifact, Envelope};

/// Resolve a GAP envelope against an optional base artifact.
///
/// # Arguments
/// * `op_json` - Null-terminated JSON string of the operation envelope.
/// * `art_json` - Null-terminated JSON string of the base artifact, or NULL for synthesize.
///
/// # Returns
/// Heap-allocated null-terminated JSON string: `{"artifact": {...}, "handle": {...}}`.
/// Returns NULL on error. Caller must free the result with `gap_free_string`.
///
/// # Safety
/// Both pointers must be valid null-terminated C strings or NULL (for `art_json`).
#[no_mangle]
pub unsafe extern "C" fn gap_resolve_envelope(
    op_json: *const c_char,
    art_json: *const c_char,
) -> *mut c_char {
    let op_str = match unsafe { CStr::from_ptr(op_json) }.to_str() {
        Ok(s) => s,
        Err(_) => return std::ptr::null_mut(),
    };

    let art_str = if art_json.is_null() {
        None
    } else {
        match unsafe { CStr::from_ptr(art_json) }.to_str() {
            Ok(s) => Some(s),
            Err(_) => return std::ptr::null_mut(),
        }
    };

    match resolve(op_str, art_str) {
        Ok(s) => CString::new(s)
            .map(|c| c.into_raw())
            .unwrap_or(std::ptr::null_mut()),
        Err(_) => std::ptr::null_mut(),
    }
}

/// Free a string previously returned by `gap_resolve_envelope`.
///
/// # Safety
/// The pointer must have been returned by `gap_resolve_envelope` and not yet freed.
#[no_mangle]
pub unsafe extern "C" fn gap_free_string(s: *mut c_char) {
    if !s.is_null() {
        drop(unsafe { CString::from_raw(s) });
    }
}

fn resolve(op_json: &str, art_json: Option<&str>) -> anyhow::Result<String> {
    let envelope: Envelope = serde_json::from_str(op_json)?;
    let artifact = art_json
        .map(serde_json::from_str::<Artifact>)
        .transpose()?;
    let (result_artifact, handle) = apply::apply(artifact.as_ref(), &envelope)?;
    Ok(serde_json::to_string(&serde_json::json!({
        "artifact": result_artifact,
        "handle": handle,
    }))?)
}
