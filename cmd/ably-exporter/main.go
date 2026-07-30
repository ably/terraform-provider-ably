// Command ably-exporter generates Terraform configuration for the resources in an
// Ably account. It needs an account token and nothing else:
//
//	ABLY_ACCOUNT_TOKEN=... go run ./cmd/ably-exporter -out ./ably
//
// Every Control API call it makes is a read.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ably/terraform-provider-ably/internal/exporter"
)

// hiddenFlags are accepted but left out of the usage text. -url points the
// exporter at a different Control API, which only Ably's own environments need.
var hiddenFlags = map[string]bool{"url": true}

// VERSION is reported in the exporter's User-Agent. Overridden at release time
// with -ldflags="-X main.VERSION=x.y.z", as the provider binary is.
var VERSION = "1.0.0"

// appList collects a repeatable -app flag, also accepting a comma-separated
// list in one go.
type appList []string

func (a *appList) String() string { return strings.Join(*a, ",") }

func (a *appList) Set(value string) error {
	before := len(*a)
	for part := range strings.SplitSeq(value, ",") {
		if part = strings.TrimSpace(part); part != "" {
			*a = append(*a, part)
		}
	}
	// An -app that parses to nothing would widen the export to the whole
	// account. Catches -app with an unset shell variable.
	if len(*a) == before {
		return fmt.Errorf("no app named in %q", value)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "ably-exporter: %s\n", err)
		os.Exit(1)
	}
}

// options are the flag values, gathered so that flag registration can be shared
// with the tests.
type options struct {
	token           *string
	url             *string
	out             *string
	secrets         *string
	imports         *bool
	singleFile      *bool
	providerVersion *string
	force           *bool
	showVersion     *bool
}

// registerFlags defines the command line, writing app filters into apps.
func registerFlags(apps *appList) (*flag.FlagSet, *options) {
	flags := flag.NewFlagSet("ably-exporter", flag.ContinueOnError)
	opts := &options{
		token: flags.String("token", "", "Ably account token. Defaults to $ABLY_ACCOUNT_TOKEN."),
		url:   flags.String("url", "", "Control API URL. Defaults to $ABLY_URL."),
		out:   flags.String("out", "./ably-export", "Directory to write the generated configuration into."),
		secrets: flags.String("secrets", string(exporter.SecretsInline),
			"What to do with sensitive values: inline (write them), vars (reference sensitive variables) or omit (leave them out)."),
		imports:    flags.Bool("imports", true, "Generate import blocks alongside the configuration."),
		singleFile: flags.Bool("single-file", false, "Write everything to main.tf instead of one file per app."),
		providerVersion: flags.String("provider-version", exporter.DefaultProviderVersion,
			"Version constraint for the generated required_providers block. Empty omits it."),
		force:       flags.Bool("force", false, "Write into the output directory even if it already contains .tf files."),
		showVersion: flags.Bool("version", false, "Print the exporter version and exit."),
	}
	flags.Var(apps, "app", "Only export this app, by ID or name. Repeatable, or comma-separated.")

	flags.SetOutput(os.Stderr)
	flags.Usage = func() { fmt.Fprint(os.Stderr, usageText(flags)) }
	return flags, opts
}

func run() error {
	var apps appList
	flags, opts := registerFlags(&apps)

	if err := flags.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	if *opts.showVersion {
		fmt.Println("ably-exporter " + VERSION)
		return nil
	}

	secretMode, err := exporter.ParseSecretMode(*opts.secrets)
	if err != nil {
		return err
	}

	accountToken := *opts.token
	if accountToken == "" {
		accountToken = os.Getenv("ABLY_ACCOUNT_TOKEN")
	}
	if accountToken == "" {
		return fmt.Errorf("no account token: pass -token or set ABLY_ACCOUNT_TOKEN")
	}

	controlURL := *opts.url
	if controlURL == "" {
		controlURL = os.Getenv("ABLY_URL")
	}

	result, err := exporter.Run(context.Background(), exporter.Config{
		Token:           accountToken,
		URL:             controlURL,
		Apps:            apps,
		Secrets:         secretMode,
		Imports:         *opts.imports,
		SingleFile:      *opts.singleFile,
		ProviderVersion: *opts.providerVersion,
		Version:         VERSION,
	})
	if err != nil {
		return err
	}

	if err := exporter.WriteFiles(*opts.out, result, *opts.force); err != nil {
		return err
	}

	report(result, *opts.out, secretMode)
	return nil
}

// usageText builds the usage text, skipping hiddenFlags. It walks the flag set
// rather than listing flags by hand so it can't drift from what is registered.
func usageText(flags *flag.FlagSet) string {
	var buf strings.Builder
	buf.WriteString("Usage: ably-exporter [flags]\n\n" +
		"Generates Terraform configuration for the resources in an Ably account.\n" +
		"Only reads from the Control API.\n\nFlags:\n")

	flags.VisitAll(func(f *flag.Flag) {
		if hiddenFlags[f.Name] {
			return
		}
		name, usage := flag.UnquoteUsage(f)
		if name != "" {
			name = " " + name
		}
		fmt.Fprintf(&buf, "  -%s%s\n    \t%s", f.Name, name, usage)
		// Defaults that carry information, so not "" or false.
		if f.DefValue != "" && f.DefValue != "false" {
			fmt.Fprintf(&buf, " (default %s)", f.DefValue)
		}
		buf.WriteString("\n")
	})
	return buf.String()
}

// report prints a summary to stderr, leaving stdout free to pipe.
func report(result *exporter.Result, out string, secrets exporter.SecretMode) {
	fmt.Fprint(os.Stderr, result.Summary())
	fmt.Fprintf(os.Stderr, "\nWritten to %s:\n", out)
	for _, file := range result.Files {
		fmt.Fprintf(os.Stderr, "  %s\n", file.Name)
	}

	if len(result.Warnings) > 0 {
		fmt.Fprintf(os.Stderr, "\n%d warnings:\n", len(result.Warnings))
		for _, warning := range result.Warnings {
			fmt.Fprintf(os.Stderr, "  %s\n", warning)
		}
	}

	if len(result.Missing) > 0 {
		fmt.Fprintf(os.Stderr, "\n%d required values are missing because the Control API does not return them:\n", len(result.Missing))
		for _, missing := range result.Missing {
			fmt.Fprintf(os.Stderr, "  %s\n", missing)
		}
	}

	if len(result.Sensitive) > 0 {
		switch secrets {
		case exporter.SecretsInline:
			fmt.Fprintf(os.Stderr, "\n%d sensitive values were written into the output. Treat these files as secrets.\n",
				len(result.Sensitive))
		case exporter.SecretsVars:
			if result.Variables {
				fmt.Fprintf(os.Stderr, "\n%d sensitive values were replaced with variables. Fill in variables.tf before applying.\n",
					len(result.Sensitive))
			} else {
				fmt.Fprintf(os.Stderr, "\n%d sensitive values were left out: none could be replaced with a variable. See the comments in the output.\n",
					len(result.Sensitive))
			}
		case exporter.SecretsOmit:
			fmt.Fprintf(os.Stderr, "\n%d sensitive values were left out. The config will not apply cleanly until you add them.\n",
				len(result.Sensitive))
		}
	}

	fmt.Fprintf(os.Stderr, "\nNext: cd %s && terraform init && terraform plan\n", out)
}
