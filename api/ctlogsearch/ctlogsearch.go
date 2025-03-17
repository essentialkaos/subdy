package ctlogsearch

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2025 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"strings"
	"time"

	"github.com/essentialkaos/ek/v13/req"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// cert contains cert info
type cert struct {
	IssuedName string `json:"issuedname"`
}

// search contains search result
type search struct {
	Data []*cert `json:"data"`
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Find tries to find subdomains using CTLogSearch API
func Find(domain string) ([]string, error) {
	resp, err := req.Request{
		URL:         "https://ctlogsearch.com/api/v1/search/domain/valid/" + domain,
		Query:       req.Query{"_": time.Now().Unix()},
		Accept:      req.CONTENT_TYPE_JSON,
		AutoDiscard: true,
	}.Get()

	if err != nil {
		return nil, fmt.Errorf("Can't send request to CTLogSearch API: %w", err)
	}

	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("CTLogSearch API returned non-ok status code %d", resp.StatusCode)
	}

	certs := &search{}
	err = resp.JSON(certs)

	if err != nil {
		return nil, fmt.Errorf("Can't decode CTLogSearch API response: %w", err)
	}

	var subdomains []string

	for _, cert := range certs.Data {
		if strings.HasPrefix(cert.IssuedName, "*") ||
			!strings.HasPrefix(cert.IssuedName, domain) {
			continue
		}

		subdomains = append(subdomains, cert.IssuedName)
	}

	return subdomains, nil
}
