package validation

import "testing"

func TestValidateModule(t *testing.T) {
	for _, ok := range []string{"github.com/acme/app", "app", "example.com/x/v2", "my-proj/my_app"} {
		if err := ValidateModule(ok); err != nil {
			t.Errorf("ValidateModule(%q) = %v, want nil", ok, err)
		}
	}
	for _, bad := range []string{"", "a b", "x\nreplace y => z", "/abs", "a//b", "a/", "../x", "a/.hidden"} {
		if ValidateModule(bad) == nil {
			t.Errorf("ValidateModule(%q) = nil, want error", bad)
		}
	}
}
