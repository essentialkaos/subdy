package app

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2023 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"os"
	"strings"

	"github.com/essentialkaos/ek/v12/fmtc"
	"github.com/essentialkaos/ek/v12/fsutil"
	"github.com/essentialkaos/ek/v12/options"
	"github.com/essentialkaos/ek/v12/req"
	"github.com/essentialkaos/ek/v12/strutil"
	"github.com/essentialkaos/ek/v12/usage"
	"github.com/essentialkaos/ek/v12/usage/completion/bash"
	"github.com/essentialkaos/ek/v12/usage/completion/fish"
	"github.com/essentialkaos/ek/v12/usage/completion/zsh"
	"github.com/essentialkaos/ek/v12/usage/man"
	"github.com/essentialkaos/ek/v12/usage/update"

	"github.com/essentialkaos/subdy/cli/support"
	"github.com/essentialkaos/subdy/dns"
	"github.com/essentialkaos/subdy/subdomains"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Basic utility info
const (
	APP  = "subdy"
	VER  = "0.0.2"
	DESC = "CLI for subdomain.center API"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Options
const (
	OPT_IP       = "I:ip"
	OPT_DNS      = "D:dns"
	OPT_NO_COLOR = "nc:no-color"
	OPT_HELP     = "h:help"
	OPT_VER      = "v:version"

	OPT_VERB_VER     = "vv:verbose-version"
	OPT_COMPLETION   = "completion"
	OPT_GENERATE_MAN = "generate-man"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// subdomainInfo contains subdomain info
type subdomainInfo struct {
	name string
	ip   string
}

// ////////////////////////////////////////////////////////////////////////////////// //

// optMap contains information about all supported options
var optMap = options.Map{
	OPT_IP:       {Type: options.BOOL},
	OPT_DNS:      {Type: options.STRING, Value: "cloudflare"},
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
}

// useRawOutput is raw output flag (for cli command)
var useRawOutput = false

// ////////////////////////////////////////////////////////////////////////////////// //

// Run is main utility function
func Run(gitRev string, gomod []byte) {
	preConfigureUI()

	args, errs := options.Parse(optMap)

	if len(errs) != 0 {
		printError(errs[0].Error())
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
		support.Print(APP, VER, gitRev, gomod)
		os.Exit(0)
	case options.GetB(OPT_HELP) || len(args) == 0:
		genUsage().Print()
		os.Exit(0)
	}

	err := process(args)

	if err != nil {
		printError(err.Error())
		os.Exit(1)
	}
}

// ////////////////////////////////////////////////////////////////////////////////// //

// preConfigureUI preconfigures UI based on information about user terminal
func preConfigureUI() {
	term := os.Getenv("TERM")

	fmtc.DisableColors = true

	if term != "" {
		switch {
		case strings.Contains(term, "xterm"),
			strings.Contains(term, "color"),
			term == "screen":
			fmtc.DisableColors = false
		}
	}

	if !fsutil.IsCharacterDevice("/dev/stdout") && os.Getenv("FAKETTY") == "" {
		fmtc.DisableColors = true
		useRawOutput = true
	}

	if os.Getenv("NO_COLOR") != "" {
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

// process starts arguments processing
func process(args options.Arguments) error {
	domain := args.Get(0).String()

	if !strings.Contains(domain, ".") {
		return fmt.Errorf("%s is not valid domain")
	}

	fmtc.If(!useRawOutput).TPrintf("{s-}Searching subdomains…{!}")

	subdomains, err := subdomains.Find(domain)

	if err != nil {
		fmtc.TPrintf("")
		return err
	}

	if len(subdomains) == 0 {
		fmtc.TPrintf("")
		printWarn("There are no subdomains for this domain")
		return nil
	}

	subdomainsInfo := processSubdomains(subdomains)

	fmtc.TPrintf("")

	if !useRawOutput {
		printSubdomainsInfo(subdomainsInfo)
	} else {
		printRawSubdomainsInfo(subdomainsInfo)
	}

	return nil
}

// processSubdomains enriches subdomains info
func processSubdomains(subdomains []string) []subdomainInfo {
	var result []subdomainInfo

	resolver := getDoHResolver()

	for index, subdomain := range subdomains {
		if options.GetB(OPT_IP) {
			fmtc.If(!useRawOutput).TPrintf("{s-}[%d/%d] Resolving subdomain IP…{!}", index, len(subdomains))

			ip, err := resolver.Resolve(subdomain, useRawOutput)

			if err != nil && !useRawOutput {
				ip = fmt.Sprintf("error: %v", err)
			}

			result = append(result, subdomainInfo{name: subdomain, ip: ip})
		} else {
			result = append(result, subdomainInfo{name: subdomain})
		}
	}

	return result
}

// printSubdomainsInfo prints subdomains info
func printSubdomainsInfo(subdomains []subdomainInfo) {
	fmtc.NewLine()

	for _, domainInfo := range subdomains {
		if domainInfo.ip != "" {
			fmtc.Printf(" {s}•{!} %s {s-}(%s){!}\n", domainInfo.name, domainInfo.ip)
		} else {
			fmtc.Printf(" {s}•{!} %s\n", domainInfo.name)
		}
	}

	fmtc.NewLine()
}

// printRawSubdomainsInfo prints subdomains info for raw output
func printRawSubdomainsInfo(subdomains []subdomainInfo) {
	for _, domainInfo := range subdomains {
		fmt.Println(domainInfo.name, domainInfo.ip)
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

// printError prints error message to console
func printError(f string, a ...interface{}) {
	if len(a) == 0 {
		fmtc.Fprintln(os.Stderr, "{r}"+f+"{!}")
	} else {
		fmtc.Fprintf(os.Stderr, "{r}"+f+"{!}\n", a...)
	}
}

// printError prints warning message to console
func printWarn(f string, a ...interface{}) {
	if len(a) == 0 {
		fmtc.Fprintln(os.Stderr, "{y}"+f+"{!}")
	} else {
		fmtc.Fprintf(os.Stderr, "{y}"+f+"{!}\n", a...)
	}
}

// ////////////////////////////////////////////////////////////////////////////////// //

// printCompletion prints completion for given shell
func printCompletion() int {
	info := genUsage()

	switch options.GetS(OPT_COMPLETION) {
	case "bash":
		fmt.Printf(bash.Generate(info, "subdy"))
	case "fish":
		fmt.Printf(fish.Generate(info, "subdy"))
	case "zsh":
		fmt.Printf(zsh.Generate(info, optMap, "subdy"))
	default:
		return 1
	}

	return 0
}

// printMan prints man page
func printMan() {
	fmt.Println(
		man.Generate(
			genUsage(),
			genAbout(""),
		),
	)
}

// genUsage generates usage info
func genUsage() *usage.Info {
	info := usage.NewInfo("", "domain")

	info.AddOption(OPT_IP, "Resolve subdomains IP")
	info.AddOption(OPT_DNS, "DoH provider {s-}({_}cloudflare{!_}|google|url){!}", "name-or-url")
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
		App:           APP,
		Version:       VER,
		Desc:          DESC,
		Year:          2009,
		Owner:         "ESSENTIAL KAOS",
		License:       "Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>",
		UpdateChecker: usage.UpdateChecker{"essentialkaos/subdy", update.GitHubChecker},
	}

	if gitRev != "" {
		about.Build = "git:" + gitRev
	}

	return about
}

// ////////////////////////////////////////////////////////////////////////////////// //
