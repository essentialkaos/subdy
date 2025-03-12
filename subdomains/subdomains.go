package subdomains

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2025 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"

	"github.com/essentialkaos/ek/v13/req"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const API_URL = "https://api.subdomain.center"

// ////////////////////////////////////////////////////////////////////////////////// //

type response []string

// ////////////////////////////////////////////////////////////////////////////////// //

// Find fetches subdomains for given domain
func Find(domain string) ([]string, error) {
	resp, err := req.Request{
		URL: API_URL,
		Query: req.Query{
			"domain": domain,
		},
		Accept:      req.CONTENT_TYPE_JSON,
		AutoDiscard: true,
	}.Get()

	if err != nil {
		return nil, fmt.Errorf("Can't fetch subdomains data: %w", err)
	}

	subdomains := make([]string, 0)
	err = resp.JSON(&subdomains)

	if err != nil {
		return nil, fmt.Errorf("Can't decode response: %w", err)
	}

	return subdomains, nil
}

// ////////////////////////////////////////////////////////////////////////////////// //
