package app

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2025 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/essentialkaos/ek/v13/fmtc"
	"github.com/essentialkaos/ek/v13/options"
	"github.com/essentialkaos/ek/v13/req"
	"github.com/essentialkaos/ek/v13/sortutil"
	"github.com/essentialkaos/ek/v13/strutil"
	"github.com/essentialkaos/ek/v13/support"
	"github.com/essentialkaos/ek/v13/support/deps"
	"github.com/essentialkaos/ek/v13/terminal"
	"github.com/essentialkaos/ek/v13/terminal/tty"
	"github.com/essentialkaos/ek/v13/usage"
	"github.com/essentialkaos/ek/v13/usage/completion/bash"
	"github.com/essentialkaos/ek/v13/usage/completion/fish"
	"github.com/essentialkaos/ek/v13/usage/completion/zsh"
	"github.com/essentialkaos/ek/v13/usage/man"
	"github.com/essentialkaos/ek/v13/usage/update"

	"github.com/essentialkaos/subdy/api/certspotter"
	"github.com/essentialkaos/subdy/api/ctlogsearch"
	"github.com/essentialkaos/subdy/api/subdomains"
	"github.com/essentialkaos/subdy/dns"
	"github.com/essentialkaos/subdy/probe"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Basic utility info
const (
	APP  = "subdy"
	VER  = "0.3.0"
	DESC = "CLI for subdomain.center API"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Options
const (
	OPT_IP       = "I:ip"
	OPT_DNS      = "D:dns"
	OPT_PROBE    = "P:probe"
	OPT_NO_COLOR = "nc:no-color"
	OPT_HELP     = "h:help"
	OPT_VER      = "v:version"

	OPT_VERB_VER     = "vv:verbose-version"
	OPT_COMPLETION   = "completion"
	OPT_GENERATE_MAN = "generate-man"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	// ENV_CERT_SPOTTER is environment variable name with CertSpotter API token
	ENV_CERT_SPOTTER = "CT_TOKEN"

	// ENV_SUBDOMAINS is environment variable name with Subdomain Center API
	// authentication code
	ENV_SUBDOMAINS = "SD_TOKEN"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// subdomain contains subdomain info
type subdomain struct {
	name     string
	ip       *dns.Answer
	services []string
}

// ////////////////////////////////////////////////////////////////////////////////// //

// optMap contains information about all supported options
var optMap = options.Map{
	OPT_IP:       {Type: options.BOOL},
	OPT_DNS:      {Type: options.STRING, Value: "cloudflare"},
	OPT_PROBE:    {Type: options.BOOL},
	OPT_NO_COLOR: {Type: options.BOOL},
	OPT_HELP:     {Type: options.BOOL},
	OPT_VER:      {Type: options.MIXED},

	OPT_VERB_VER:     {Type: options.BOOL},
	OPT_COMPLETION:   {},
	OPT_GENERATE_MAN: {Type: options.BOOL},
}

// dohProviders is map with DoH providers
var dohProviders = map[string]string{
	"cf":         dns.CLOUDFLARE,
	"cloudflare": dns.CLOUDFLARE,
	"google":     dns.GOOGLE,
	"quad9":      dns.QUAD9,
}

// useRawOutput is raw output flag (for cli command)
var useRawOutput = false

// ////////////////////////////////////////////////////////////////////////////////// //

// Run is main utility function
func Run(gitRev string, gomod []byte) {
	preConfigureUI()

	args, errs := options.Parse(optMap)

	if !errs.IsEmpty() {
		terminal.Error("Options parsing errors:")
		terminal.Error(errs.Error(" - "))
		os.Exit(1)
	}

	configureUI()

	switch {
	case options.Has(OPT_COMPLETION):
		os.Exit(printCompletion())
	case options.Has(OPT_GENERATE_MAN):
		printMan()
		os.Exit(0)
	case options.GetB(OPT_VER):
		genAbout(gitRev).Print(options.GetS(OPT_VER))
		os.Exit(0)
	case options.GetB(OPT_VERB_VER):
		support.Collect(APP, VER).
			WithRevision(gitRev).
			WithDeps(deps.Extract(gomod)).
			WithChecks(checkAPIAvailability()).
			Print()
		os.Exit(0)
	case options.GetB(OPT_HELP) || len(args) == 0:
		genUsage().Print()
		os.Exit(0)
	}

	err := validateOptionsAndArgs(args)

	if err != nil {
		terminal.Error(err)
		os.Exit(1)
	}

	err = process(args)

	if err != nil {
		terminal.Error(err)
		os.Exit(1)
	}
}

// ////////////////////////////////////////////////////////////////////////////////// //

// preConfigureUI preconfigures UI based on information about user terminal
func preConfigureUI() {
	if !tty.IsTTY() {
		fmtc.DisableColors = true
	}
}

// configureUI configures user interface
func configureUI() {
	if options.GetB(OPT_NO_COLOR) {
		fmtc.DisableColors = true
	}

	req.SetUserAgent(APP, VER)
}

// validateOptionsAndArgs validates options and arguments
func validateOptionsAndArgs(args options.Arguments) error {
	domain := args.Get(0).String()

	if !strings.Contains(domain, ".") {
		return fmt.Errorf("%q is not valid domain", domain)
	}

	if options.Has(OPT_DNS) {
		dns := options.GetS(OPT_DNS)

		if !strings.Contains(dns, ".") && dohProviders[dns] == "" {
			return fmt.Errorf("Unknown DNS-over-HTTPS provider %q", dns)
		}
	}

	return nil
}

// process starts arguments processing
func process(args options.Arguments) error {

	domain := args.Get(0).ToLower().String()
	subdomains := searchSubdomains(domain)

	if len(subdomains) == 0 {
		terminal.Warn("There are no subdomains for this domain")
		return nil
	}

	subdomainsInfo := processSubdomains(subdomains)

	if !useRawOutput {
		printSubdomainsInfo(subdomainsInfo)
	} else {
		printRawSubdomainsInfo(subdomainsInfo)
	}

	return nil
}

// searchSubdomains searches subdomains using various sources
func searchSubdomains(domain string) []string {
	var result []string

	fmtc.If(!useRawOutput).TPrintf("{s-}Searching subdomains using subdomain.center…{!}")

	subdomains, err := subdomains.Find(domain, os.Getenv(ENV_SUBDOMAINS))

	if err != nil {
		fmtc.If(!useRawOutput).TPrintf("{r}▲ %v{!}\n", err)
	} else {
		result = append(result, subdomains...)
	}

	fmtc.If(!useRawOutput).TPrintf("{s-}Searching subdomains using CTLogSearch…{!}")

	subdomains, err = ctlogsearch.Find(domain)

	if err != nil {
		fmtc.If(!useRawOutput).TPrintf("{r}▲ %v{!}\n", err)
	} else {
		result = append(result, subdomains...)
	}

	fmtc.If(!useRawOutput).TPrintf("{s-}Searching subdomains using CertSpotter…{!}")

	subdomains, err = certspotter.Find(domain, os.Getenv(ENV_CERT_SPOTTER))

	if err != nil {
		fmtc.If(!useRawOutput).TPrintf("{r}▲ %v{!}\n", err)
	} else {
		fmtc.If(!useRawOutput).TPrintf("")
	}

	result = append(result, subdomains...)

	return result
}

// processSubdomains enriches subdomains info
func processSubdomains(subdomains []string) []*subdomain {
	var result []*subdomain

	defer fmtc.If(!useRawOutput).TPrintf("")

	resolver := getDoHResolver()

	sortutil.StringsNatural(subdomains)
	subdomains = slices.CompactFunc(subdomains, func(s1, s2 string) bool {
		return strings.ToLower(s1) == strings.ToLower(s2)
	})

	for index, name := range subdomains {
		name = strings.ToLower(name)

		if options.GetB(OPT_IP) || options.GetB(OPT_PROBE) {
			fmtc.If(!useRawOutput).TPrintf(
				"{s-}[%d/%d] Resolving %s IP…{!}",
				index, len(subdomains), name,
			)

			answer, err := resolver.Resolve(name)

			if err != nil {
				continue
			}

			result = append(result, &subdomain{name: name, ip: answer})
		} else {
			result = append(result, &subdomain{name: name})
		}
	}

	if !useRawOutput && options.GetB(OPT_PROBE) {
		for index, info := range result {
			fmtc.TPrintf(
				"{s-}[%d/%d] Probing %s…{!}",
				index, len(result), info.name,
			)

			info.services = probe.Probe(info.ip.IP())
		}
	}

	return result
}

// printSubdomainsInfo prints subdomains info
func printSubdomainsInfo(subdomains []*subdomain) {
	fmtc.NewLine()

	for _, info := range subdomains {
		if !info.ip.IsEmpty() {
			fmtc.Printf(
				" {s}•{!} %s {s-}(%s){!}",
				info.name, info.ip.ToString(false),
			)
		} else {
			fmtc.Printf(" {s}•{!} %s", info.name)
		}

		if len(info.services) != 0 {
			fmt.Print(" " + getColoredServicePorts(info.services))
		}

		fmtc.NewLine()
	}

	fmtc.NewLine()
}

// printRawSubdomainsInfo prints subdomains info for raw output
func printRawSubdomainsInfo(subdomains []*subdomain) {
	for _, info := range subdomains {
		fmt.Println(info.name, info.ip.ToString(true))
	}
}

// getDoHResolver returns DoH resolver
func getDoHResolver() *dns.Resolver {
	resolverURL, ok := dohProviders[options.GetS(OPT_DNS)]

	if !ok {
		resolverURL = strutil.Exclude(options.GetS(OPT_DNS), "https://")
	}

	return &dns.Resolver{resolverURL}
}

// getColoredServicePorts formats list of services
func getColoredServicePorts(services []string) string {
	return strutil.JoinFunc(services, " ", func(s string) string {
		colorTag := "{#152}"

		switch s {
		case "22", "23", "5800":
			colorTag = "{#67}"
		case "25", "110", "143", "220", "993", "995":
			colorTag = "{#173}"
		case "21", "115", "445", "636", "990", "3389":
			colorTag = "{#140}"
		case "80", "443", "3000", "8080", "8443", "9000":
			colorTag = "{#151}"
		case "53":
			colorTag = "{#153}"
		case "1434", "3306", "3690", "5432", "6379", "6432", "9042", "27017":
			colorTag = "{#221}"
		}

		return fmtc.Sprintf(colorTag+"[%s]{!}", s)
	})
}

// ////////////////////////////////////////////////////////////////////////////////// //

// checkAPIAvailability checks API availability
func checkAPIAvailability() support.Check {
	chk := support.Check{support.CHECK_ERROR, "subdomain.center API", ""}

	start := time.Now()
	resp, err := req.Request{
		URL:         "https://api.subdomain.center",
		AutoDiscard: true,
	}.Get()
	dur := time.Since(start)

	if err != nil {
		chk.Message = "Can't send request"
		return chk
	}

	if resp.StatusCode != 200 {
		chk.Message = fmt.Sprintf("API returned non-ok status code (%d)", resp.StatusCode)
		return chk
	}

	chk.Status = support.CHECK_OK

	if dur < 500*time.Millisecond {
		chk.Message = "accessible and healthy"
	} else {
		chk.Message = "accessible"
	}

	return chk
}

// printCompletion prints completion for given shell
func printCompletion() int {
	info := genUsage()

	switch options.GetS(OPT_COMPLETION) {
	case "bash":
		fmt.Print(bash.Generate(info, APP))
	case "fish":
		fmt.Print(fish.Generate(info, APP))
	case "zsh":
		fmt.Print(zsh.Generate(info, optMap, APP))
	default:
		return 1
	}

	return 0
}

// printMan prints man page
func printMan() {
	fmt.Println(man.Generate(genUsage(), genAbout("")))
}

// genUsage generates usage info
func genUsage() *usage.Info {
	info := usage.NewInfo("", "domain")

	info.AddOption(OPT_IP, "Resolve subdomains IP")
	info.AddOption(OPT_DNS, "DoH JSON provider {s-}({_}cloudflare{!_}|google|quad9|custom-url){!}", "name-or-url")
	info.AddOption(OPT_PROBE, "Probe subdomains for open ports")
	info.AddOption(OPT_NO_COLOR, "Disable colors in output")
	info.AddOption(OPT_HELP, "Show this help message")
	info.AddOption(OPT_VER, "Show version")

	info.AddExample(
		"go.dev", "Find all subdomains of go.dev",
	)

	info.AddExample(
		"-I go.dev", "Find all subdomains of go.dev and resolve their IPs",
	)

	info.AddExample(
		"-I -D google go.dev", "Find all subdomains of go.dev and resolve their IPs using Google DNS",
	)

	return info
}

// genAbout generates info about version
func genAbout(gitRev string) *usage.About {
	about := &usage.About{
		App:     APP,
		Version: VER,
		Desc:    DESC,
		Year:    2009,
		Owner:   "ESSENTIAL KAOS",
		License: "Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>",
	}

	about.DescSeparator = "{s}—{!}"

	if gitRev != "" {
		about.Build = "git:" + gitRev
		about.UpdateChecker = usage.UpdateChecker{"essentialkaos/subdy", update.GitHubChecker}
	}

	return about
}

// ////////////////////////////////////////////////////////////////////////////////// //
