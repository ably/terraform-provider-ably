package exporter

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// resourceRef identifies an exported resource by type and Control API ID.
type resourceRef struct {
	resourceType string
	id           string
}

// referenceableAttributes maps an attribute holding another resource's ID onto
// that resource's type. The exporter rewrites these as references, so the config
// carries its dependencies: app_id = ably_app.foo.id rather than "abcdef".
//
// It is an allowlist, not a rule like "anything ending in _id", because plenty of
// attributes hold unrelated identifiers (a Bodyguard channel_id, an Azure app ID)
// that would be rewritten into nonsense on an ID collision. Pairing with the type
// also stops a namespace ID being read as a queue ID.
var referenceableAttributes = map[string]string{
	"app_id":         resourceTypeApp,
	"queue_id":       resourceTypeQueue,
	"signing_key_id": resourceTypeKey,
}

// labeller hands out unique HCL resource labels, derived from Ably names rather
// than IDs because the output is meant to be read. Same-type collisions get a
// numeric suffix in discovery order, which is deterministic because discovery
// sorts by ID.
type labeller struct {
	used map[string]bool
}

func newLabeller() *labeller {
	return &labeller{used: map[string]bool{}}
}

// assign returns a label not yet used for resourceType. Labels only need to be
// unique per type, so an app and its rules can share a name.
func (l *labeller) assign(resourceType, base string) string {
	label := sanitiseLabel(base)
	if l.claim(resourceType, label) {
		return label
	}
	for suffix := 2; ; suffix++ {
		candidate := fmt.Sprintf("%s_%d", label, suffix)
		if l.claim(resourceType, candidate) {
			return candidate
		}
	}
}

func (l *labeller) claim(resourceType, label string) bool {
	key := resourceType + "." + label
	if l.used[key] {
		return false
	}
	l.used[key] = true
	return true
}

// sanitiseLabel turns an arbitrary Ably name into a valid HCL identifier.
func sanitiseLabel(name string) string {
	var builder strings.Builder
	previousUnderscore := false

	for _, r := range strings.ToLower(strings.TrimSpace(name)) {
		switch {
		case r >= 'a' && r <= 'z', unicode.IsDigit(r):
			builder.WriteRune(r)
			previousUnderscore = false
		default:
			if !previousUnderscore && builder.Len() > 0 {
				builder.WriteRune('_')
				previousUnderscore = true
			}
		}
	}

	label := strings.Trim(builder.String(), "_")
	if label == "" {
		return "unnamed"
	}
	// HCL identifiers can't start with a digit. Test the rune, not the first
	// byte: the filter above accepts any Unicode digit.
	if first, _ := utf8.DecodeRuneInString(label); unicode.IsDigit(first) {
		label = "r" + label
	}
	return label
}

// labelFor builds the label base for a target. Named resources use their name;
// rules have none, so they take their app's label and a numeric suffix when an
// app has several of the same type.
func labelFor(target Target, appLabel string) string {
	switch {
	case target.ResourceType == resourceTypeApp:
		if target.Name != "" {
			return target.Name
		}
		return target.ID
	case target.Name != "":
		return appLabel + "_" + sanitiseLabel(target.Name)
	default:
		return appLabel
	}
}
