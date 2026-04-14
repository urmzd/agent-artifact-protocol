package runner

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/urmzd/saige/agent/types"
)

func parseTurnNum(name string) int {
	parts := strings.SplitN(name, "-", 2)
	if len(parts) < 2 {
		return 0
	}
	n, _ := strconv.Atoi(parts[1])
	return n
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// extractStreamError scans deltas for ErrorDelta and returns a combined error, or nil.
func extractStreamError(deltas []types.Delta) error {
	var errs []string
	for _, d := range deltas {
		if e, ok := d.(types.ErrorDelta); ok {
			errs = append(errs, e.Error.Error())
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("stream errors: %s", strings.Join(errs, "; "))
}
