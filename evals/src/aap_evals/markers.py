"""Universal XML section markers — mirrors src/markers.rs.

All formats use `<aap:section id="...">` / `</aap:section>`.
JSON uses pointer addressing instead.
"""

from __future__ import annotations


def markers_for(section_id: str, fmt: str) -> tuple[str, str] | None:
    """Return (start, end) marker pair, or None for JSON."""
    if fmt == "application/json":
        return None
    return f'<aap:section id="{section_id}">', "</aap:section>"


def marker_example(fmt: str) -> str:
    """Return a human-readable marker example for prompts."""
    if fmt == "application/json":
        return ""
    return '<aap:section id="ID"> ... </aap:section>'


def extract_section_content(content: str, section_id: str, fmt: str) -> str | None:
    pair = markers_for(section_id, fmt)
    if not pair:
        return None
    start, end = pair
    si = content.find(start)
    ei = content.find(end)
    if si == -1 or ei == -1:
        return None
    return content[si + len(start) : ei]
