"""
Agent-Artifact Protocol (AAP) data model — Python implementation of aap/1.0.

Provides dataclasses for envelopes, diff operations, section updates,
template bindings, chunk frames, and token budgets.
"""
from __future__ import annotations

import hashlib
import json
import uuid
from dataclasses import dataclass, field, asdict
from datetime import datetime, timezone
from typing import Literal

Mode = Literal["full", "diff", "section", "template", "composite", "manifest"]
OpType = Literal["replace", "insert_before", "insert_after", "delete"]
Priority = Literal["completeness", "brevity", "fidelity"]

PROTOCOL_VERSION = "aap/1.0"

ArtifactState = Literal["draft", "published", "archived"]
RelationshipType = Literal["depends_on", "parent", "child", "derived_from", "supersedes", "related"]


@dataclass
class Permissions:
    read: list[str] = field(default_factory=list)
    write: list[str] = field(default_factory=list)
    admin: list[str] = field(default_factory=list)

    def to_dict(self) -> dict:
        d: dict = {}
        if self.read:
            d["read"] = self.read
        if self.write:
            d["write"] = self.write
        if self.admin:
            d["admin"] = self.admin
        return d


@dataclass
class Relationship:
    type: RelationshipType
    target: str
    version: int | None = None

    def to_dict(self) -> dict:
        d: dict = {"type": self.type, "target": self.target}
        if self.version is not None:
            d["version"] = self.version
        return d


@dataclass
class EntityMetadata:
    owner: str | None = None
    created_by: str | None = None
    tags: list[str] = field(default_factory=list)
    permissions: Permissions | None = None
    collection: str | None = None
    ttl: int | None = None
    expires_at: str | None = None
    relationships: list[Relationship] = field(default_factory=list)

    def to_dict(self) -> dict:
        d: dict = {}
        if self.owner is not None:
            d["owner"] = self.owner
        if self.created_by is not None:
            d["created_by"] = self.created_by
        if self.tags:
            d["tags"] = self.tags
        if self.permissions is not None:
            d["permissions"] = self.permissions.to_dict()
        if self.collection is not None:
            d["collection"] = self.collection
        if self.ttl is not None:
            d["ttl"] = self.ttl
        if self.expires_at is not None:
            d["expires_at"] = self.expires_at
        if self.relationships:
            d["relationships"] = [r.to_dict() for r in self.relationships]
        return d


@dataclass
class AdvisoryLock:
    held_by: str
    acquired_at: str
    ttl: int

    def to_dict(self) -> dict:
        return {"held_by": self.held_by, "acquired_at": self.acquired_at, "ttl": self.ttl}


@dataclass
class SseError:
    code: str
    message: str
    fatal: bool = False
    seq: int | None = None

    def to_dict(self) -> dict:
        d: dict = {"code": self.code, "message": self.message}
        if self.fatal:
            d["fatal"] = True
        if self.seq is not None:
            d["seq"] = self.seq
        return d


@dataclass
class Target:
    section: str | None = None
    lines: tuple[int, int] | None = None
    offsets: tuple[int, int] | None = None
    search: str | None = None
    pointer: str | None = None

    def to_dict(self) -> dict:
        if self.pointer is not None:
            return {"pointer": self.pointer}
        if self.section is not None:
            return {"section": self.section}
        if self.lines is not None:
            return {"lines": list(self.lines)}
        if self.offsets is not None:
            return {"offsets": list(self.offsets)}
        if self.search is not None:
            return {"search": self.search}
        raise ValueError("Target must have exactly one addressing mode")


@dataclass
class DiffOp:
    op: OpType
    target: Target
    content: str | None = None

    def to_dict(self) -> dict:
        d: dict = {"op": self.op, "target": self.target.to_dict()}
        if self.content is not None:
            d["content"] = self.content
        return d


@dataclass
class SectionUpdate:
    id: str
    content: str

    def to_dict(self) -> dict:
        return {"id": self.id, "content": self.content}


@dataclass
class SectionDef:
    id: str
    label: str | None = None
    start_marker: str | None = None
    end_marker: str | None = None

    def to_dict(self) -> dict:
        d: dict = {"id": self.id}
        if self.label:
            d["label"] = self.label
        if self.start_marker:
            d["start_marker"] = self.start_marker
        if self.end_marker:
            d["end_marker"] = self.end_marker
        return d


@dataclass
class TokenBudget:
    max_tokens: int | None = None
    priority: Priority | None = None
    max_sections: int | None = None

    def to_dict(self) -> dict:
        d: dict = {}
        if self.max_tokens is not None:
            d["max_tokens"] = self.max_tokens
        if self.priority is not None:
            d["priority"] = self.priority
        if self.max_sections is not None:
            d["max_sections"] = self.max_sections
        return d


@dataclass
class Include:
    ref: str | None = None
    uri: str | None = None
    content: str | None = None
    hash: str | None = None

    def to_dict(self) -> dict:
        d: dict = {}
        if self.ref is not None:
            d["ref"] = self.ref
        if self.uri is not None:
            d["uri"] = self.uri
        if self.content is not None:
            d["content"] = self.content
        if self.hash is not None:
            d["hash"] = self.hash
        return d


@dataclass
class SectionPrompt:
    id: str
    prompt: str
    dependencies: list[str] = field(default_factory=list)
    token_budget: int | None = None

    def to_dict(self) -> dict:
        d: dict = {"id": self.id, "prompt": self.prompt}
        if self.dependencies:
            d["dependencies"] = self.dependencies
        if self.token_budget is not None:
            d["token_budget"] = self.token_budget
        return d


@dataclass
class Envelope:
    id: str
    version: int
    format: str
    mode: Mode
    content: str | None = None
    base_version: int | None = None
    encoding: str = "utf-8"
    created_at: str | None = None
    updated_at: str | None = None
    token_budget: TokenBudget | None = None
    tokens_used: int | None = None
    checksum: str | None = None
    sections: list[SectionDef] = field(default_factory=list)
    operations: list[DiffOp] = field(default_factory=list)
    target_sections: list[SectionUpdate] = field(default_factory=list)
    template: str | None = None
    bindings: dict | None = None
    includes: list[Include] = field(default_factory=list)
    content_encoding: str | None = None
    # Manifest mode fields
    skeleton: str | None = None
    section_prompts: list[SectionPrompt] = field(default_factory=list)
    section_id: str | None = None
    # Entity state
    state: ArtifactState | None = None
    state_changed_at: str | None = None
    entity: EntityMetadata | None = None
    lock: AdvisoryLock | None = None

    def to_dict(self) -> dict:
        d: dict = {
            "protocol": PROTOCOL_VERSION,
            "id": self.id,
            "version": self.version,
            "format": self.format,
            "mode": self.mode,
        }
        if self.encoding != "utf-8":
            d["encoding"] = self.encoding
        if self.base_version is not None:
            d["base_version"] = self.base_version
        if self.created_at:
            d["created_at"] = self.created_at
        if self.updated_at:
            d["updated_at"] = self.updated_at
        if self.token_budget:
            d["token_budget"] = self.token_budget.to_dict()
        if self.tokens_used is not None:
            d["tokens_used"] = self.tokens_used
        if self.checksum:
            d["checksum"] = self.checksum
        if self.sections:
            d["sections"] = [s.to_dict() for s in self.sections]
        if self.content is not None:
            d["content"] = self.content
        if self.operations:
            d["operations"] = [op.to_dict() for op in self.operations]
        if self.target_sections:
            d["target_sections"] = [s.to_dict() for s in self.target_sections]
        if self.template is not None:
            d["template"] = self.template
        if self.bindings is not None:
            d["bindings"] = self.bindings
        if self.includes:
            d["includes"] = [i.to_dict() for i in self.includes]
        if self.skeleton is not None:
            d["skeleton"] = self.skeleton
        if self.section_prompts:
            d["section_prompts"] = [sp.to_dict() for sp in self.section_prompts]
        if self.section_id is not None:
            d["section_id"] = self.section_id
        if self.content_encoding:
            d["content_encoding"] = self.content_encoding
        if self.state is not None:
            d["state"] = self.state
        if self.state_changed_at:
            d["state_changed_at"] = self.state_changed_at
        if self.entity:
            d["entity"] = self.entity.to_dict()
        if self.lock:
            d["lock"] = self.lock.to_dict()
        return d

    def to_json(self, indent: int = 2) -> str:
        return json.dumps(self.to_dict(), indent=indent)

    def resolve_content(self, store: dict[str, str] | None = None) -> str:
        """Resolve the final content string from this envelope."""
        if self.mode == "full":
            return self.content or ""
        if self.mode == "diff":
            if not store or self.id not in store:
                raise ValueError(f"No base content for artifact {self.id}")
            return apply_diff(store[self.id], self.operations, self.format, self.sections)
        if self.mode == "section":
            if not store or self.id not in store:
                raise ValueError(f"No base content for artifact {self.id}")
            return apply_section_update(store[self.id], self.target_sections, self.format, self.sections)
        if self.mode == "template":
            return fill_template(self.template or "", self.bindings or {})
        if self.mode == "composite":
            return resolve_composite(self.includes, store or {}, self.format, self.sections)
        raise ValueError(f"Unknown mode: {self.mode}")


@dataclass
class ChunkFrame:
    seq: int
    content: str
    envelope: dict | None = None
    section_id: str | None = None
    flush: bool = False
    final: bool = False

    def to_dict(self) -> dict:
        d: dict = {"seq": self.seq, "content": self.content}
        if self.envelope:
            d["envelope"] = self.envelope
        if self.section_id:
            d["section_id"] = self.section_id
        if self.flush:
            d["flush"] = True
        if self.final:
            d["final"] = True
        return d

    def to_json(self) -> str:
        return json.dumps(self.to_dict())


# ── helper functions ──────────────────────────────────────────────────────────


def new_id() -> str:
    return str(uuid.uuid4())


def now_iso() -> str:
    return datetime.now(timezone.utc).isoformat()


def sha256_checksum(content: str) -> str:
    return "sha256:" + hashlib.sha256(content.encode()).hexdigest()


def _apply_pointer_op(data: dict | list, pointer: str, op: DiffOp) -> None:
    """Apply a single pointer-targeted operation to parsed JSON data."""
    parts = [p.replace("~1", "/").replace("~0", "~") for p in pointer.split("/")[1:]]
    if not parts:
        raise ValueError("Cannot operate on root pointer")

    # Navigate to parent
    parent = data
    for part in parts[:-1]:
        if isinstance(parent, list):
            parent = parent[int(part)]
        else:
            parent = parent[part]

    key = parts[-1]

    if op.op == "replace":
        new_val = json.loads(op.content)
        if isinstance(parent, list):
            parent[int(key)] = new_val
        else:
            parent[key] = new_val
    elif op.op == "delete":
        if isinstance(parent, list):
            del parent[int(key)]
        else:
            del parent[key]
    elif op.op in ("insert_before", "insert_after"):
        if not isinstance(parent, list):
            raise ValueError("insert_before/insert_after require array parent")
        idx = int(key)
        if op.op == "insert_after":
            idx += 1
        new_val = json.loads(op.content)
        parent.insert(idx, new_val)


def apply_diff(
    base: str,
    operations: list[DiffOp],
    format: str = "text/html",
    sections: list[SectionDef] | None = None,
) -> str:
    """Apply diff operations sequentially to base content."""
    from aap.markers import find_section_def, find_section_range

    # Check for pointer operations
    has_pointer = any(op.target.pointer is not None for op in operations)
    if has_pointer:
        data = json.loads(base)
        for op in operations:
            if op.target.pointer is None:
                raise ValueError("Mixing pointer and non-pointer targets is not supported")
            _apply_pointer_op(data, op.target.pointer, op)
        return json.dumps(data, indent=2)

    result = base
    for op in operations:
        t = op.target
        if t.search is not None:
            idx = result.find(t.search)
            if idx == -1:
                raise ValueError(f"Search target not found: {t.search!r}")
            start, end = idx, idx + len(t.search)
        elif t.offsets is not None:
            start, end = t.offsets
        elif t.lines is not None:
            lines = result.split("\n")
            s, e = t.lines[0] - 1, t.lines[1]
            start = sum(len(l) + 1 for l in lines[:s])
            end = sum(len(l) + 1 for l in lines[:e]) - 1
        elif t.section is not None:
            section_def = find_section_def(sections, t.section)
            start, end = find_section_range(result, t.section, format, section_def)
        else:
            raise ValueError("No addressing mode in target")

        if op.op == "replace":
            result = result[:start] + (op.content or "") + result[end:]
        elif op.op == "delete":
            result = result[:start] + result[end:]
        elif op.op == "insert_before":
            result = result[:start] + (op.content or "") + result[start:]
        elif op.op == "insert_after":
            result = result[:end] + (op.content or "") + result[end:]

    return result


def apply_section_update(
    base: str,
    updates: list[SectionUpdate],
    format: str = "text/html",
    sections: list[SectionDef] | None = None,
) -> str:
    """Replace section content in base, preserving markers and other sections."""
    from aap.markers import find_section_def, resolve_markers

    result = base
    for update in updates:
        section_def = find_section_def(sections, update.id)
        start_marker, end_marker = resolve_markers(update.id, format, section_def)
        si = result.find(start_marker)
        ei = result.find(end_marker)
        if si == -1 or ei == -1:
            raise ValueError(f"Section markers not found: {update.id}")
        before = result[: si + len(start_marker)]
        after = result[ei:]
        result = before + "\n" + update.content + "\n" + after
    return result


def fill_template(template: str, bindings: dict) -> str:
    """Simple Mustache-subset template filling (variable substitution only)."""
    result = template
    for key, value in bindings.items():
        # Unescaped triple-brace
        result = result.replace(f"{{{{{{{key}}}}}}}", str(value))
        # Regular double-brace
        result = result.replace(f"{{{{{key}}}}}", str(value))
    return result


def resolve_composite(
    includes: list[Include],
    store: dict[str, str],
    format: str = "text/html",
    sections: list[SectionDef] | None = None,
) -> str:
    """Assemble content from includes."""
    from aap.markers import find_section_def, find_section_range_inclusive

    parts = []
    for inc in includes:
        if inc.content is not None:
            parts.append(inc.content)
        elif inc.ref is not None:
            if ":" in inc.ref:
                artifact_id, section_id = inc.ref.split(":", 1)
                content = store.get(artifact_id, "")
                section_def = find_section_def(sections, section_id)
                try:
                    start, end = find_section_range_inclusive(
                        content, section_id, format, section_def
                    )
                    parts.append(content[start:end])
                except ValueError:
                    pass  # Section not found, skip
            else:
                parts.append(store.get(inc.ref, ""))
        elif inc.uri is not None:
            parts.append(f"<!-- unresolved: {inc.uri} -->")
    return "\n".join(parts)
