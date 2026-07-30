package main

import (
	"flag"
	"strings"
	"testing"
)

// TestUsageTextHidesInternalFlags guards the customer-facing help text. -url
// points the exporter at a different Control API, which only Ably's own
// environments need, so it must not appear.
func TestUsageTextHidesInternalFlags(t *testing.T) {
	text := usageText(exporterFlags())

	for _, unwanted := range []string{"-url", "ABLY_URL", "production"} {
		if strings.Contains(text, unwanted) {
			t.Errorf("usage mentions %q:\n%s", unwanted, text)
		}
	}

	// The flags customers do use have to survive the filter.
	for _, wanted := range []string{"-token", "-out", "-app", "-secrets", "-imports", "-force"} {
		if !strings.Contains(text, wanted) {
			t.Errorf("usage is missing %s:\n%s", wanted, text)
		}
	}
}

// TestHiddenFlagsAreRegistered checks every hidden flag exists. A typo would
// hide nothing and go unnoticed.
func TestHiddenFlagsAreRegistered(t *testing.T) {
	flags := exporterFlags()
	for name := range hiddenFlags {
		if flags.Lookup(name) == nil {
			t.Errorf("hiddenFlags names %q, which is not a registered flag", name)
		}
	}
}

func TestAppListRejectsEmptyValues(t *testing.T) {
	var apps appList
	if err := apps.Set(""); err == nil {
		t.Error("an empty -app should fail rather than export the whole account")
	}
	if err := apps.Set(" , "); err == nil {
		t.Error("an -app of separators should fail")
	}
	if err := apps.Set("one, two"); err != nil {
		t.Fatalf("Set: %s", err)
	}
	if got := apps.String(); got != "one,two" {
		t.Errorf("apps = %q, want one,two", got)
	}
}

// exporterFlags builds the flag set the way run does, for tests that inspect it.
func exporterFlags() *flag.FlagSet {
	var apps appList
	flags, _ := registerFlags(&apps)
	return flags
}
