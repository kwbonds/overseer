package test

import "github.com/skx/overseer/utils"

type Result struct {
	Input  string `json:"input"`
	Target string `json:"target"`
	Time   int64  `json:"time"`
	Type   string `json:"type"`
	Tag    string `json:"tag"`

	// If not nil, test has failed
	Error *string `json:"error"`
}

// Generated a unique identifier for the original test (e.g. to deduplicate same results)
func (result *Result) Hash() string {
	return utils.GetMD5Hash(result.Input + result.Target + result.Type + result.Tag)
}
