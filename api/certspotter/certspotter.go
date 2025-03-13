package certspotter

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

// cert contains cert info
type cert struct {
	DNSNames []string `json:"dns_names"`
}

// certs is a slice of certs
type certs []*cert

// ////////////////////////////////////////////////////////////////////////////////// //

// Find tries to find subdomains using CertSpotter API
func Find(domain string) ([]string, error) {
	resp, err := req.Request{
		URL: "https://api.certspotter.com/v1/issuances",
		Query: req.Query{
			"domain":             domain,
			"include_subdomains": true,
			"expand":             "dns_names",
		},
		Accept:      req.CONTENT_TYPE_JSON,
		AutoDiscard: true,
	}.Get()

	if err != nil {
		return nil, fmt.Errorf("Can't send request to CertSpotter API: %w", err)
	}

	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("CertSpotter API returned non-ok status code %d", resp.StatusCode)
	}

	certs := certs{}
	err = resp.JSON(&certs)

	if err != nil {
		return nil, fmt.Errorf("Can't decode API response: %w", err)
	}

	var subdomains []string

	for _, cert := range certs {
		for _, subdomain := range cert.DNSNames {
			if strings.HasPrefix(subdomain, "*") {
				continue
			}

			subdomains = append(subdomains, subdomain)
		}
	}

	return subdomains, nil
}
