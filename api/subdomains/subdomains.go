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

// Find tries to find subdomains using subdomain.center API
func Find(domain string) ([]string, error) {
	resp, err := req.Request{
		URL:         "https://api.subdomain.center",
		Query:       req.Query{"domain": domain},
		Accept:      req.CONTENT_TYPE_JSON,
		AutoDiscard: true,
	}.Get()

	if err != nil {
		return nil, fmt.Errorf("Can't send request to subdomain.center API: %w", err)
	}

	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("subdomain.center API returned non-ok status code %d", resp.StatusCode)
	}

	subdomains := make([]string, 0)
	err = resp.JSON(&subdomains)

	if err != nil {
		return nil, fmt.Errorf("Can't decode API response: %w", err)
	}

	return subdomains, nil
}

// ////////////////////////////////////////////////////////////////////////////////// //
