"""AAP envelope application — delegates to Rust apply engine via PyO3 FFI.

The Rust engine (src/apply.rs) is the single source of truth for envelope
resolution. This module serialises typed Python envelopes to JSON, calls
the FFI, and deserialises the result.
"""

from __future__ import annotations

import json

from .schema import DiffEnvelope, FullEnvelope, SectionEnvelope, TemplateEnvelope

try:
    from aap_evals.aap import resolve_envelope as _rust_resolve  # type: ignore[import-not-found]
except ImportError as exc:
    raise ImportError(
        "Rust apply engine not available — build with `maturin develop`"
    ) from exc


type AnyEnvelope = FullEnvelope | DiffEnvelope | SectionEnvelope | TemplateEnvelope


def apply_envelope(artifact: str, envelope: AnyEnvelope, fmt: str) -> str:
    """Resolve a typed AAP envelope against artifact content via Rust FFI.

    Args:
        artifact: Current artifact body (raw text).
        envelope: Typed envelope from the maintain context.
        fmt: MIME type (e.g. "text/html").

    Returns:
        New artifact body after applying the operation.
    """
    operation_json = envelope.model_dump_json(exclude_none=True)

    # The Rust engine expects the base artifact as a name:"full" envelope JSON.
    artifact_envelope = json.dumps({
        "protocol": "aap/0.1",
        "id": envelope.id,
        "version": envelope.version - 1,
        "name": "full",
        "operation": {"direction": "output", "format": fmt},
        "content": [{"body": artifact}],
    })

    # For full envelopes, no base artifact is needed.
    if isinstance(envelope, FullEnvelope):
        result_json = _rust_resolve(operation_json, None)
    else:
        result_json = _rust_resolve(operation_json, artifact_envelope)

    result = json.loads(result_json)
    return result["content"][0]["body"]
