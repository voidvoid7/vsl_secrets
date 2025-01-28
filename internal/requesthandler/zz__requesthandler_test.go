package requesthandler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/voidvoid7/vsl_secrets/internal/secretmap"
)

func makeRequestHandler(t *testing.T) *requestHandlerImpl {
	secretMapHolder := secretmap.NewSecretMapHolder()
	staticCache, err := createStaticCache()
	if err != nil {
		t.Fatal(err)
	}
	return &requestHandlerImpl{
		templateCache:   createTemplateCache(&staticFiles),
		staticCache:     staticCache,
		staticFS:        &staticFiles,
		secretMapHolder: secretMapHolder,
	}
}

func TestFavicon(t *testing.T) {
	handler := makeRequestHandler(t)
	r, err := http.NewRequest(http.MethodGet, "/"+faviconPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	handler.faviconHandler(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestCss(t *testing.T) {
	handler := makeRequestHandler(t)
	r, err := http.NewRequest(http.MethodGet, "/"+cssPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	handler.cssHandler(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHome(t *testing.T) {
	handler := makeRequestHandler(t)
	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	handler.generalHandler(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestCreateSecret(t *testing.T) {
	handler := makeRequestHandler(t)
	r, err := http.NewRequest(http.MethodPost, "/create", nil)
	if err != nil {
		t.Fatal(err)
	}
	secret := "test me _ćčdaddy1*的"
	r.Form = map[string][]string{"secretValue": {secret}}
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	handler.createSecretHandler(w, r)
	if w.Code != http.StatusFound {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
	redirectLocation := w.Header().Get("Location")
	t.Log("redirectLocation", redirectLocation)
	expectedRedirectLocationStart := "/get-secret/"
	if !strings.HasPrefix(redirectLocation, expectedRedirectLocationStart) {
		t.Fatalf("Expected redirect to /get-secret/, got %s", redirectLocation)
	}
	base64MapKey := redirectLocation[len(expectedRedirectLocationStart):]
	if len(base64MapKey) != (1032/3)*4 {
		t.Fatalf("Invalid base64MapKey length %d", len(base64MapKey))
	}
	// at this point secret should exist in secret map and we will attempt to read it
	r, err = http.NewRequest(http.MethodGet, expectedRedirectLocationStart+base64MapKey, nil)
	if err != nil {
		t.Fatal(err)
	}
	r.SetPathValue(secretPathVariableName, base64MapKey)
	w = httptest.NewRecorder()
	handler.getSecretHandler(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// finally fetch it with post
	r, err = http.NewRequest(http.MethodPost, "decrypt-secret", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.PostForm = map[string][]string{decryptionSecretFormFieldName: {base64MapKey}}
	w = httptest.NewRecorder()
	handler.decryptSecretHandler(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
	respHtml := w.Body.String()
	preSecretTag := "<pre class=\"secret-showcase\">"
	postSecretTag := "</pre>"
	start := strings.Index(respHtml, preSecretTag)
	end := strings.Index(respHtml, postSecretTag)
	if start == -1 || end == -1 {
		t.Fatalf("Expected secret to be wrapped in pre tag, got %s", respHtml)
	}
	fetchedSecret := respHtml[start+len(preSecretTag) : end]
	if fetchedSecret != secret {
		t.Fatalf("Expected secret %s, got %s", secret, fetchedSecret)
	}

	// attempt to read again
	r, err = http.NewRequest(http.MethodPost, "decrypt-secret", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.PostForm = map[string][]string{decryptionSecretFormFieldName: {base64MapKey}}
	w = httptest.NewRecorder()
	handler.decryptSecretHandler(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}
