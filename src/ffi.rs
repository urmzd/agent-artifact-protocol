//! PyO3 bindings for the AAP apply engine.
//!
//! Exposes `resolve_envelope` to Python, which takes an operation envelope
//! JSON string and an optional base artifact envelope JSON string, and
//! returns `{"artifact": {...}, "handle": {...}}` as a JSON string.

use pyo3::prelude::*;

use crate::aap::Envelope;
use crate::apply;

/// Resolve an AAP operation against an optional base artifact.
///
/// Args:
///     operation_json: JSON string of the operation envelope.
///     artifact_json: JSON string of the base artifact envelope (a `name:"synthesize"` envelope).
///         Required for edit ops, ignored for synthesize.
///
/// Returns:
///     JSON string: `{"artifact": <envelope>, "handle": <envelope>}`
///
/// Raises:
///     ValueError: If the envelope is malformed or the operation fails.
#[pyfunction]
fn resolve_envelope(operation_json: &str, artifact_json: Option<&str>) -> PyResult<String> {
    let operation: Envelope = serde_json::from_str(operation_json)
        .map_err(|e| PyErr::new::<pyo3::exceptions::PyValueError, _>(format!("invalid operation envelope JSON: {e}")))?;

    let artifact = artifact_json
        .map(|json| {
            serde_json::from_str::<Envelope>(json)
                .map_err(|e| PyErr::new::<pyo3::exceptions::PyValueError, _>(format!("invalid artifact envelope JSON: {e}")))
        })
        .transpose()?;

    let (resolved, handle) = apply::apply(artifact.as_ref(), &operation)
        .map_err(|e| PyErr::new::<pyo3::exceptions::PyValueError, _>(format!("apply failed: {e}")))?;

    let result = serde_json::json!({
        "artifact": serde_json::to_value(&resolved)
            .map_err(|e| PyErr::new::<pyo3::exceptions::PyValueError, _>(format!("artifact serialization failed: {e}")))?,
        "handle": serde_json::to_value(&handle)
            .map_err(|e| PyErr::new::<pyo3::exceptions::PyValueError, _>(format!("handle serialization failed: {e}")))?,
    });

    serde_json::to_string(&result)
        .map_err(|e| PyErr::new::<pyo3::exceptions::PyValueError, _>(format!("serialization failed: {e}")))
}

/// AAP apply engine Python module.
#[pymodule]
fn aap(m: &Bound<'_, PyModule>) -> PyResult<()> {
    m.add_function(wrap_pyfunction!(resolve_envelope, m)?)?;
    Ok(())
}
