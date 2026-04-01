"""
Generation strategy implementations for aap/1.0.

Each function produces an Envelope using the most token-efficient mode
for the given change scenario.
"""
from __future__ import annotations

import difflib
import re

from aap.aap import (
    Envelope,
    DiffOp,
    Target,
    SectionDef,
    SectionUpdate,
    Include,
    new_id,
    now_iso,
    sha256_checksum,
)


def generate_full(
    content: str,
    format: str = "text/html",
    artifact_id: str | None = None,
    version: int = 1,
    sections: list[SectionDef] | None = None,
) -> Envelope:
    """Generate a full-mode envelope with complete content."""
    return Envelope(
        id=artifact_id or new_id(),
        version=version,
        format=format,
        mode="full",
        content=content,
        created_at=now_iso() if version == 1 else None,
        updated_at=now_iso(),
        checksum=sha256_checksum(content),
        sections=sections or [],
    )


def generate_diff(
    artifact_id: str,
    base_version: int,
    old_content: str,
    new_content: str,
    version: int | None = None,
    format: str = "text/html",
) -> Envelope:
    """Compute minimal diff operations between old and new content.

    Uses search-based targeting for changed lines — finds the old text
    and replaces with the new text.
    """
    old_lines = old_content.splitlines(keepends=True)
    new_lines = new_content.splitlines(keepends=True)
    matcher = difflib.SequenceMatcher(None, old_lines, new_lines)
    ops: list[DiffOp] = []

    for tag, i1, i2, j1, j2 in matcher.get_opcodes():
        if tag == "equal":
            continue
        old_text = "".join(old_lines[i1:i2])
        new_text = "".join(new_lines[j1:j2])
        if tag == "replace":
            ops.append(DiffOp(
                op="replace",
                target=Target(search=old_text),
                content=new_text,
            ))
        elif tag == "delete":
            ops.append(DiffOp(
                op="delete",
                target=Target(search=old_text),
            ))
        elif tag == "insert":
            # Insert before the next line in the old content
            if i1 < len(old_lines):
                anchor = old_lines[i1]
                ops.append(DiffOp(
                    op="insert_before",
                    target=Target(search=anchor),
                    content=new_text,
                ))
            else:
                # Appending at end — use offset
                ops.append(DiffOp(
                    op="insert_after",
                    target=Target(offsets=(len(old_content), len(old_content))),
                    content=new_text,
                ))

    return Envelope(
        id=artifact_id,
        version=version or base_version + 1,
        format=format,
        mode="diff",
        base_version=base_version,
        updated_at=now_iso(),
        operations=ops,
    )


def generate_section_update(
    artifact_id: str,
    base_version: int,
    updates: dict[str, str],
    version: int | None = None,
    format: str = "text/html",
) -> Envelope:
    """Generate a section-mode envelope replacing specific sections."""
    return Envelope(
        id=artifact_id,
        version=version or base_version + 1,
        format=format,
        mode="section",
        base_version=base_version,
        updated_at=now_iso(),
        target_sections=[
            SectionUpdate(id=sid, content=content)
            for sid, content in updates.items()
        ],
    )


def generate_template_fill(
    template: str,
    bindings: dict[str, str],
    artifact_id: str | None = None,
    version: int = 1,
    format: str = "text/html",
) -> Envelope:
    """Generate a template-mode envelope with slot bindings."""
    return Envelope(
        id=artifact_id or new_id(),
        version=version,
        format=format,
        mode="template",
        updated_at=now_iso(),
        template=template,
        bindings=bindings,
    )


def generate_composite(
    includes: list[Include],
    artifact_id: str | None = None,
    version: int = 1,
    format: str = "text/html",
) -> Envelope:
    """Generate a composite-mode envelope assembling sub-artifacts."""
    return Envelope(
        id=artifact_id or new_id(),
        version=version,
        format=format,
        mode="composite",
        created_at=now_iso(),
        includes=includes,
    )


def extract_sections(content: str, format: str = "text/html") -> list[SectionDef]:
    """Extract section definitions from content using format-appropriate markers."""
    from aap.markers import extract_sections_regex, resolve_markers

    pattern = extract_sections_regex(format)
    if pattern is None:
        return []
    sections = []
    for match in re.finditer(pattern, content):
        sid = match.group(1)
        start_marker, end_marker = resolve_markers(sid, format)
        sections.append(SectionDef(
            id=sid,
            start_marker=start_marker,
            end_marker=end_marker,
        ))
    return sections


def get_section_content(content: str, section_id: str, format: str = "text/html") -> str:
    """Extract the content between section markers."""
    from aap.markers import find_section_range

    start, end = find_section_range(content, section_id, format)
    return content[start:end].strip()
