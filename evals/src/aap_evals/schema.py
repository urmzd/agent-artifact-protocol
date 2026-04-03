"""Spec-compliant AAP Pydantic models — mirrors ../src/aap.rs.

Three envelope types: synthesize (in), edit (in), handle (out).
Artifact is a standalone content object, not an envelope.
"""

from __future__ import annotations

from typing import Annotated, Literal, Union

from pydantic import BaseModel, Field


# ── Target definitions ────────────────────────────────────────────────────


class IdTarget(BaseModel):
    """Target an <aap:target id="..."> marker by ID."""

    type: Literal["id"]
    value: str


class PointerTarget(BaseModel):
    """Target a value by JSON Pointer (RFC 6901)."""

    type: Literal["pointer"]
    value: str


DiffTarget = Annotated[
    Union[IdTarget, PointerTarget],
    Field(discriminator="type"),
]


class EditOp(BaseModel):
    """A single edit operation."""

    op: Literal["replace", "insert_before", "insert_after", "delete"]
    target: DiffTarget
    content: str | None = None


class SynthesizeContentItem(BaseModel):
    """Content item for name=synthesize."""

    body: str


# ── Operation metadata ────────────────────────────────────────────────────


class Meta(BaseModel):
    """Envelope metadata."""

    format: str = "text/html"
    tokens_used: int | None = None
    checksum: str | None = None
    state: str | None = None


# ── Typed envelope variants ───────────────────────────────────────────────


class SynthesizeEnvelope(BaseModel):
    """Envelope for name=synthesize (full artifact generation)."""

    protocol: Literal["aap/0.1"] = "aap/0.1"
    id: str
    version: int
    name: Literal["synthesize"]
    meta: Meta = Field(default_factory=Meta)
    content: list[SynthesizeContentItem]


class EditEnvelope(BaseModel):
    """Envelope for name=edit (targeted changes via id/pointer targeting)."""

    protocol: Literal["aap/0.1"] = "aap/0.1"
    id: str
    version: int
    name: Literal["edit"]
    meta: Meta = Field(default_factory=Meta)
    content: list[EditOp]


# ── Handle ────────────────────────────────────────────────────────────────


class HandleContentItem(BaseModel):
    """Content item for name=handle — lightweight artifact reference."""

    id: str
    version: int
    token_count: int | None = None
    state: str | None = None
    content: str | None = None


class HandleEnvelope(BaseModel):
    """Envelope for name=handle (lightweight artifact reference)."""

    protocol: Literal["aap/0.1"] = "aap/0.1"
    id: str
    version: int
    name: Literal["handle"]
    meta: Meta = Field(default_factory=Meta)
    content: list[HandleContentItem]


# ── Envelope union ────────────────────────────────────────────────────────


# Full envelope union (all three types)
Envelope = Annotated[
    Union[SynthesizeEnvelope, EditEnvelope, HandleEnvelope],
    Field(discriminator="name"),
]

# LLM output type — only the two input envelopes
LLMEnvelope = Annotated[
    Union[SynthesizeEnvelope, EditEnvelope],
    Field(discriminator="name"),
]
