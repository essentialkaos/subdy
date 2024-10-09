package dns

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2024 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"strings"

	"github.com/essentialkaos/ek/v13/req"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Resolver is DoH resolver
type Resolver struct {
	URL string
}

// ////////////////////////////////////////////////////////////////////////////////// //

type info struct {
	Status int      `json:"Status"`
	Answer []answer `json:"Answer"`
}

type answer struct {
	Data string `json:"data"`
	Type int    `json:"type"`
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Resolve returns IP or chain for given domain
func (r *Resolver) Resolve(domain string, simple bool) (string, error) {
	resp, err := req.Request{
		URL:    "https://" + r.URL,
		Accept: "application/dns-json",
		Query: req.Query{
			"name": domain,
		},
		AutoDiscard: true,
	}.Get()

	if err != nil {
		return "", err
	}

	info := &info{}
	err = resp.JSON(info)

	if err != nil {
		return "", fmt.Errorf("Can't decode response: %v", err)
	}

	return formatInfo(info, simple), nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// formatInfo formats info data
func formatInfo(info *info, simple bool) string {
	if info.Status != 0 || len(info.Answer) == 0 {
		return ""
	}

	var result string

	for _, a := range info.Answer {
		if a.Type == 5 {
			if !simple {
				result += a.Data + " → "
			}
		} else {
			if simple {
				result += a.Data + " "
			} else {
				result += a.Data + " / "
			}
		}

	}

	return strings.TrimRight(result, "/→ ")
}
