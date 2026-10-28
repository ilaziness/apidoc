package config

import "testing"

func TestDetectLanguage(t *testing.T) {
	cases := map[string]string{
		"":            LangEN,
		"en-US":       LangEN,
		"fr-FR":       LangEN,
		"zh":          LangZHHans,
		"zh-CN":       LangZHHans,
		"zh_CN":       LangZHHans,
		"zh-Hans":     LangZHHans,
		"zh-Hans-CN":  LangZHHans,
		"zh-SG":       LangZHHans,
		"zh-TW":       LangZHHant,
		"zh_TW.UTF-8": LangZHHant,
		"zh-HK":       LangZHHant,
		"zh-MO":       LangZHHant,
		"zh-Hant":     LangZHHant,
		"zh-Hant-TW":  LangZHHant,
	}
	for in, want := range cases {
		if got := DetectLanguage(in); got != want {
			t.Errorf("DetectLanguage(%q)=%q, want %q", in, got, want)
		}
	}
}
