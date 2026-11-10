package i18n

import (
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

func TestCatalogIsComplete(t *testing.T) {
	for _, k := range Keys() {
		txt, _ := Lookup(k)
		if !txt.Complete() {
			t.Errorf("catalog key %q lacks a translation", k)
		}
		if strings.Count(txt[RU], "%") != strings.Count(txt[EN], "%") {
			t.Errorf("catalog key %q: ru and en have different format verbs", k)
		}
	}
	if S(RU, "no.such.key") != "no.such.key" {
		t.Error("unknown keys must be returned verbatim")
	}
}

func TestParseAndDetect(t *testing.T) {
	for in, want := range map[string]Lang{"ru": RU, "ru_RU.UTF-8": RU, "EN": EN, "en_GB.UTF-8": EN} {
		if got, ok := Parse(in); !ok || got != want {
			t.Errorf("Parse(%q) = %v, %v", in, got, ok)
		}
	}
	if _, ok := Parse("de"); ok {
		t.Error("de is not supported")
	}
	t.Setenv("SURVIVE85_LANG", "")
	t.Setenv("LC_ALL", "")
	t.Setenv("LC_MESSAGES", "")
	t.Setenv("LANG", "ru_RU.UTF-8")
	if Detect() != RU {
		t.Error("LANG=ru_RU must select Russian")
	}
	t.Setenv("LANG", "de_DE.UTF-8")
	if Detect() != EN {
		t.Error("other locales fall back to English")
	}
	t.Setenv("SURVIVE85_LANG", "ru")
	if Detect() != RU {
		t.Error("SURVIVE85_LANG wins")
	}
	if RU.Next() != EN || EN.Next() != RU {
		t.Error("Next must toggle")
	}
}

func TestTextYAML(t *testing.T) {
	var v struct {
		A Text `yaml:"a"`
		B Text `yaml:"b"`
	}
	if err := yaml.Unmarshal([]byte("a: {ru: привет, en: hello}\nb: same\n"), &v); err != nil {
		t.Fatal(err)
	}
	if v.A.Get(RU) != "привет" || v.A.Get(EN) != "hello" || v.B.Get(RU) != "same" || !v.B.Complete() {
		t.Fatalf("decoded %+v", v)
	}
	if err := yaml.Unmarshal([]byte("a: {de: hallo}\n"), &v); err == nil {
		t.Fatal("unknown language keys must be rejected")
	}
}

func TestMoneyAndNumbers(t *testing.T) {
	cases := []struct {
		l    Lang
		v    int64
		want string
	}{
		{RU, 154000, "154 000 ₽"},
		{RU, -1500, "−1 500 ₽"},
		{RU, 0, "0 ₽"},
		{EN, 154000, "$6,160"},
		{EN, -2500000, "−$100,000"},
	}
	for _, c := range cases {
		if got := Money(c.l, c.v); got != c.want {
			t.Errorf("Money(%s, %d) = %q, want %q", c.l, c.v, got, c.want)
		}
	}
	if SignedMoney(EN, 2500) != "+$100" {
		t.Errorf("SignedMoney = %q", SignedMoney(EN, 2500))
	}
	if Percent(0.00042) != "0.042%" || Percent(0.153) != "15.3%" || Percent(0.05) != "5.00%" {
		t.Errorf("Percent: %s %s %s", Percent(0.00042), Percent(0.153), Percent(0.05))
	}
}
