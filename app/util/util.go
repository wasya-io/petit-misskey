package util

import (
	"encoding/json"
	"fmt"
)

type (
	prettiedJson struct {
		ErrorMessage string `json:"error"`
		OriginalData string `json:"data"`
	}
)

func PrittyJson(v any) string {
	j, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		p := &prettiedJson{
			ErrorMessage: "json marshal indent failed",
			OriginalData: fmt.Sprintf("%v", v),
		}
		pp, _ := json.MarshalIndent(p, "", "  ")
		return string(pp)
	}
	return string(j)
}
