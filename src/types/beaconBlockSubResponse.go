package types

type BBlock struct {
	Slot                string `json:"slot,omitempty"`
	Block               string `json:"block,omitempty"`
	ExecutionOptimistic bool   `json:"execution_optimistic,omitempty"`
}
