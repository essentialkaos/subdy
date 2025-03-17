package dns

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2025 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"strings"

	"github.com/essentialkaos/ek/v13/req"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	CLOUDFLARE = "1.1.1.1/dns-query"
	GOOGLE     = "dns.google/resolve"
	QUAD9      = "9.9.9.9:5053/dns-query"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Resolver is DoH resolver
type Resolver struct {
	URL string
}

// Answer is resolver answer
type Answer struct {
	Status  int     `json:"Status"`
	Records Records `json:"Answer"`
}

// Record is DNS record
type Record struct {
	Data string `json:"data"`
	Type int    `json:"type"`
}

// Records is a slice with records
type Records []*Record

// ////////////////////////////////////////////////////////////////////////////////// //

// resolveError is resolving error
type resolveError struct {
	Error string `json:"error"`
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Resolve returns info about given domain
func (r *Resolver) Resolve(domain string) (*Answer, error) {
	resp, err := req.Request{
		URL:    "https://" + r.URL,
		Accept: "application/dns-json",
		Query:  req.Query{"name": domain},
	}.Get()

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		errInfo := &resolveError{}
		err = resp.JSON(errInfo)

		if err == nil {
			return nil, fmt.Errorf("Resolving error: %s", errInfo.Error)
		}

		return nil, fmt.Errorf("Resolver returned non-ok status code %d", resp.StatusCode)
	}

	answer := &Answer{}
	err = resp.JSON(answer)

	if err != nil {
		return nil, fmt.Errorf("Can't decode response: %v", err)
	}

	return answer, nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// ToString returns string representation of answer
func (a *Answer) ToString(simple bool) string {
	if a == nil || a.Status != 0 || len(a.Records) == 0 {
		return ""
	}

	var result string

	for _, r := range a.Records {
		if r.Type == 5 { // 5 == CNAME
			if !simple {
				result += r.Data + " → "
			}
		} else {
			if simple {
				result += r.Data + " "
			} else {
				result += r.Data + " / "
			}
		}

	}

	return strings.TrimRight(result, "/→ ")
}

// IsEmpty returns true if answer is empty
func (a *Answer) IsEmpty() bool {
	return a == nil || len(a.Records) == 0
}

// IP returns only A records
func (a *Answer) IP() []string {
	if a == nil || a.Status != 0 || len(a.Records) == 0 {
		return nil
	}

	var result []string

	for _, r := range a.Records {
		if r.Type == 1 { // 1 == A
			result = append(result, r.Data)
		}
	}

	return result
}
