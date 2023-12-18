package models

import (
	"encoding/json"
	"fmt"
)

// FailureFlags define where abouts in a system (if at all), something should fail.
type FailureFlags struct {
	Order      bool `json:"order"`
	Payment    bool `json:"payment"`
	Inventory  bool `json:"inventory"`
	Fulfilment bool `json:"fulfilment"`
}

// JSONToFailureFlags converts a json.RawMessage into a FailureFlags struct.
func JSONToFailureFlags(j json.RawMessage) (FailureFlags, error) {
	var ff FailureFlags
	if err := json.Unmarshal(j, &ff); err != nil {
		return FailureFlags{}, fmt.Errorf("unmarshalling FailureFlags: %w", err)
	}

	return ff, nil
}
