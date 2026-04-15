package secureerr

import (
	"errors"
	"net/http"
	"testing"
)

func TestErrorReturnsSafeMessage(t *testing.T) {
	internal := errors.New("sql connection timeout: secret-host")
	appErr := Internal("USER_CREATE_FAILED", "failed to create user", internal, nil)

	if got := appErr.Error(); got != "failed to create user" {
		t.Fatalf("expected safe message, got %q", got)
	}
	if !errors.Is(appErr, internal) {
		t.Fatalf("expected wrapped internal error to remain accessible")
	}
}

func TestFromErrorWrapsUnknownErrors(t *testing.T) {
	internal := errors.New("raw database error")
	appErr := FromError(internal)

	if appErr.StatusCode() != http.StatusInternalServerError {
		t.Fatalf("expected internal status, got %d", appErr.StatusCode())
	}
	if appErr.Error() != http.StatusText(http.StatusInternalServerError) {
		t.Fatalf("expected generic safe message, got %q", appErr.Error())
	}
	if !errors.Is(appErr, internal) {
		t.Fatalf("expected unknown error to remain wrapped")
	}
}
