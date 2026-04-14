// Package apply bridges Go to the Rust GAP apply engine via CGo.
package apply

/*
#cgo LDFLAGS: -L${SRCDIR}/../../../target/release -lgap
#include <stdlib.h>

extern char* gap_resolve_envelope(const char* op_json, const char* art_json);
extern void  gap_free_string(char* s);
*/
import "C"
import (
	"encoding/json"
	"errors"
	"fmt"
	"unsafe"
)

// ApplyEnvelope resolves a GAP envelope against an artifact via the Rust engine.
// For synthesize envelopes, pass an empty artifact string.
// Returns the resolved artifact body.
func ApplyEnvelope(artifact string, envelopeJSON []byte, format string, envelopeID string, envelopeVersion int) (string, error) {
	cOp := C.CString(string(envelopeJSON))
	defer C.free(unsafe.Pointer(cOp))

	var cArt *C.char
	if artifact != "" {
		artObj := map[string]any{
			"id":      envelopeID,
			"version": envelopeVersion - 1,
			"format":  format,
			"body":    artifact,
		}
		artJSON, err := json.Marshal(artObj)
		if err != nil {
			return "", fmt.Errorf("marshal artifact: %w", err)
		}
		cArt = C.CString(string(artJSON))
		defer C.free(unsafe.Pointer(cArt))
	}

	result := C.gap_resolve_envelope(cOp, cArt)
	if result == nil {
		return "", errors.New("gap_resolve_envelope returned NULL (apply failed)")
	}
	defer C.gap_free_string(result)

	goResult := C.GoString(result)
	var parsed struct {
		Artifact struct {
			Body string `json:"body"`
		} `json:"artifact"`
	}
	if err := json.Unmarshal([]byte(goResult), &parsed); err != nil {
		return "", fmt.Errorf("unmarshal result: %w", err)
	}
	return parsed.Artifact.Body, nil
}
