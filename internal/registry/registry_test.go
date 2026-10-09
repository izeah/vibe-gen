package registry

import (
	"testing"

	"github.com/izeah01/vibe-gen/internal/generator"
)

func TestFindVariant(t *testing.T) {
	next, ok := FindFramework(generator.LanguageTypeScript, "next")
	if !ok {
		t.Fatal("typescript/next not registered")
	}
	echo, _ := FindFramework(generator.LanguageGo, "echo")

	cases := []struct {
		fw     Framework
		id     string
		want   string
		wantOK bool
	}{
		{next, "", "app", true},
		{next, "pages", "pages", true},
		{next, "nope", "", false},
		{echo, "", "", true},
		{echo, "app", "", false},
	}
	for _, c := range cases {
		v, ok := c.fw.FindVariant(c.id)
		if v.ID != c.want || ok != c.wantOK {
			t.Errorf("%s.FindVariant(%q) = (%q, %v), want (%q, %v)", c.fw.ID, c.id, v.ID, ok, c.want, c.wantOK)
		}
	}
}
