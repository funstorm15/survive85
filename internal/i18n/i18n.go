// Package i18n holds the language model shared by game data and the UI:
// a Lang enum, a bilingual Text value that YAML can decode from either a
// plain string or a {ru, en} mapping, and locale detection.
package i18n

import (
	"fmt"
	"os"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Lang is a supported interface language.
type Lang string

const (
	RU Lang = "ru"
	EN Lang = "en"
)

// All lists supported languages in toggle order.
var All = []Lang{RU, EN}

// Parse converts a user-supplied code ("ru", "en_US.UTF-8", "RU") to a Lang.
func Parse(s string) (Lang, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch {
	case strings.HasPrefix(s, "ru"):
		return RU, true
	case strings.HasPrefix(s, "en"):
		return EN, true
	}
	return "", false
}

// Next returns the language that follows l in the toggle order.
func (l Lang) Next() Lang {
	for i, x := range All {
		if x == l {
			return All[(i+1)%len(All)]
		}
	}
	return EN
}

// Detect picks a language from the environment the same way gettext does:
// SURVIVE85_LANG wins, then LC_ALL, LC_MESSAGES and LANG. Anything that is
// not Russian falls back to English.
func Detect() Lang {
	for _, key := range []string{"SURVIVE85_LANG", "LC_ALL", "LC_MESSAGES", "LANG"} {
		if v := os.Getenv(key); v != "" && v != "C" && v != "POSIX" {
			if l, ok := Parse(v); ok {
				return l
			}
			// An explicit non-Russian, non-English locale still means English.
			return EN
		}
	}
	return EN
}

// Text is a localized string. In YAML it is written either as a mapping
// ({ru: "...", en: "..."}) or, for language-neutral strings, as a scalar.
type Text map[Lang]string

// T builds a Text from Russian and English variants.
func T(ru, en string) Text { return Text{RU: ru, EN: en} }

// Get returns the text in l, falling back to English and then Russian.
func (t Text) Get(l Lang) string {
	if s := t[l]; s != "" {
		return s
	}
	if s := t[EN]; s != "" {
		return s
	}
	return t[RU]
}

// Complete reports whether every supported language has a non-empty value.
func (t Text) Complete() bool {
	for _, l := range All {
		if strings.TrimSpace(t[l]) == "" {
			return false
		}
	}
	return true
}

// IsZero reports whether no language has a value.
func (t Text) IsZero() bool {
	for _, s := range t {
		if s != "" {
			return false
		}
	}
	return true
}

// UnmarshalYAML accepts a scalar (same text for every language) or a mapping
// keyed by language code.
func (t *Text) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		*t = Text{}
		for _, l := range All {
			(*t)[l] = node.Value
		}
		return nil
	case yaml.MappingNode:
		raw := map[string]string{}
		if err := node.Decode(&raw); err != nil {
			return err
		}
		*t = Text{}
		for k, v := range raw {
			l, ok := Parse(k)
			if !ok || string(l) != strings.ToLower(k) {
				return fmt.Errorf("line %d: unknown language key %q", node.Line, k)
			}
			(*t)[l] = strings.TrimRight(v, "\n")
		}
		return nil
	default:
		return fmt.Errorf("line %d: localized text must be a string or a {ru, en} mapping", node.Line)
	}
}
