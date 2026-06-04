package errors

import "testing"

func TestGenericHTTPErrorString(t *testing.T) {
	if got := NewHTTPError(200, 401, nil).Error(); got != "invalid status code returned: 200 expected but 401 returned" {
		t.Fatalf("unexpected error string: %s", got)
	}
	if got := NewHTTPError(200, 500, []byte("boom")).Error(); got != "invalid status code returned: 200 expected but 500 returned: boom" {
		t.Fatalf("unexpected error string with body: %s", got)
	}
}
