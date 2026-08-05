package httpapi

import (
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	r:=httptest.NewRequest("GET","/healthz",nil); w:=httptest.NewRecorder(); (&Handler{}).ServeHTTP(w,r)
	if w.Code!=200 || w.Body.String()!="{\"status\":\"ok\"}\n" { t.Fatalf("unexpected response: %d %s",w.Code,w.Body.String()) }
}
