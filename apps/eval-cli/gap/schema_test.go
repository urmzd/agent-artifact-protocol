package gap_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/urmzd/saige/agent/types"

	"github.com/urmzd/generative-artifact-protocol/eval-cli/gap"
)

func TestLLMEnvelopeSchema(t *testing.T) {
	schema := types.SchemaFrom[gap.LLMEnvelopeSchema]()
	b, _ := json.MarshalIndent(schema, "", "  ")
	fmt.Println(string(b))
}
