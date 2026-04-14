// Package gap provides GAP protocol types mirroring the Rust gap.rs definitions.
package gap

import "encoding/json"

// Target types for edit operations.
type Target struct {
	Type  string `json:"type"`  // "id" or "pointer"
	Value string `json:"value"` // marker ID or JSON Pointer (RFC 6901)
}

// EditOp is a single edit operation within an edit envelope.
type EditOp struct {
	Op      string  `json:"op"`               // "replace", "insert_before", "insert_after", "delete"
	Target  Target  `json:"target"`            //
	Content *string `json:"content,omitempty"` // absent for delete
}

// SynthesizeContentItem is a content item for name=synthesize.
type SynthesizeContentItem struct {
	Body string `json:"body"`
}

// Meta holds envelope metadata.
type Meta struct {
	Format     string  `json:"format,omitempty"`
	TokensUsed *int    `json:"tokens_used,omitempty"`
	Checksum   *string `json:"checksum,omitempty"`
	State      *string `json:"state,omitempty"`
}

// Envelope is the universal GAP envelope — covers synthesize, edit, and handle.
// Content is kept as raw JSON and decoded via typed accessors.
type Envelope struct {
	Protocol string          `json:"protocol"` // "gap/0.1"
	ID       string          `json:"id"`
	Version  int             `json:"version"`
	Name     string          `json:"name"` // "synthesize", "edit", "handle"
	Meta     Meta            `json:"meta"`
	Content  json.RawMessage `json:"content"`
}

// AsEditOps decodes Content as a list of edit operations.
func (e *Envelope) AsEditOps() ([]EditOp, error) {
	var ops []EditOp
	if err := json.Unmarshal(e.Content, &ops); err != nil {
		return nil, err
	}
	return ops, nil
}

// AsSynthesizeBody decodes Content and returns the synthesize body string.
func (e *Envelope) AsSynthesizeBody() (string, error) {
	var items []SynthesizeContentItem
	if err := json.Unmarshal(e.Content, &items); err != nil {
		return "", err
	}
	if len(items) == 0 {
		return "", nil
	}
	return items[0].Body, nil
}

// IsLLMEnvelope returns true if the envelope is a valid LLM output (synthesize or edit).
func (e *Envelope) IsLLMEnvelope() bool {
	return e.Name == "synthesize" || e.Name == "edit"
}

// TargetInfo is included in handle envelope content.
type TargetInfo struct {
	ID      string  `json:"id"`
	Label   *string `json:"label,omitempty"`
	Accepts *string `json:"accepts,omitempty"`
}

// HandleContentItem is a content item for name=handle.
type HandleContentItem struct {
	ID         string       `json:"id"`
	Version    int          `json:"version"`
	TokenCount *int         `json:"token_count,omitempty"`
	State      *string      `json:"state,omitempty"`
	Content    *string      `json:"content,omitempty"`
	Targets    []TargetInfo `json:"targets,omitempty"`
}

// AsHandleContent decodes Content as a list of handle content items.
func (e *Envelope) AsHandleContent() ([]HandleContentItem, error) {
	var items []HandleContentItem
	if err := json.Unmarshal(e.Content, &items); err != nil {
		return nil, err
	}
	return items, nil
}

// LLMContentItem is a unified content item for structured LLM output.
// For synthesize: only Body is set. For edit: Op, Target, and optionally Content are set.
type LLMContentItem struct {
	Body    string  `json:"body,omitempty" description:"Full artifact body (synthesize only)"`
	Op      string  `json:"op,omitempty" description:"Edit operation: replace, insert_before, insert_after, delete" enum:"replace,insert_before,insert_after,delete"`
	Target  *Target `json:"target,omitempty" description:"Edit target (id or pointer)"`
	Content *string `json:"content,omitempty" description:"Replacement content for edit ops"`
}

// LLMEnvelopeSchema is the struct used to derive a JSON schema for structured LLM output.
// It represents the union of synthesize and edit envelopes without the handle variant.
type LLMEnvelopeSchema struct {
	Protocol string            `json:"protocol" description:"Must be gap/0.1"`
	ID       string            `json:"id" description:"Artifact identifier"`
	Version  int               `json:"version" description:"Envelope version number"`
	Name     string            `json:"name" description:"Envelope type: synthesize or edit" enum:"synthesize,edit"`
	Meta     Meta              `json:"meta" description:"Envelope metadata"`
	Content  []LLMContentItem  `json:"content" description:"Array of content items"`
}
