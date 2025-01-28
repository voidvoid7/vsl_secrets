package requesthandler

import (
	"bytes"
	"embed"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"

	"github.com/voidvoid7/vsl_secrets/internal/secretmap"
)

//go:embed static/*
var staticFiles embed.FS

const cssPath = "static/css/main.css"
const faviconPath = "static/favicon.ico"
const secretPathVariableName = "secret"
const decryptionSecretFormFieldName = "base64MapKey"

type templateCache struct {
	home       *template.Template
	getSecret  *template.Template
	showSecret *template.Template
}

type staticCache struct {
	css     []byte
	favicon []byte
}

type requestHandlerImpl struct {
	templateCache   *templateCache
	staticCache     *staticCache
	staticFS        *embed.FS
	secretMapHolder secretmap.SecretMapHolder
}

type RequestHandler interface {
	RegisterHandlers()
}

type SecretWrapper struct {
	Base64MapKey string
}

type ShowSecretWrapper struct {
	Secret string
}

func createTemplateCache(staticFS fs.FS) *templateCache {
	slog.Info("Parsing templates")
	homeTemplate := template.Must(template.ParseFS(staticFS, "static/index.html"))
	getSecretTemplate := template.Must(template.ParseFS(staticFS, "static/get-secret.html"))
	showSecret := template.Must(template.ParseFS(staticFS, "static/show-secret.html"))

	return &templateCache{home: homeTemplate, getSecret: getSecretTemplate, showSecret: showSecret}
}

func createStaticCache() (*staticCache, error) {
	css, err := staticFiles.ReadFile(cssPath)
	if err != nil {
		slog.Error("Error reading css file", "error", err)
		return nil, err
	}
	facicon, err := staticFiles.ReadFile(faviconPath)
	if err != nil {
		slog.Error("Error reading favicon file", "error", err)
		return nil, err
	}
	return &staticCache{css: css, favicon: facicon}, nil
}

func NewRequestHandler(secretMapHolder secretmap.SecretMapHolder) (RequestHandler, error) {
	staticCache, err := createStaticCache()
	if err != nil {
		return nil, err
	}
	return &requestHandlerImpl{
		templateCache:   createTemplateCache(&staticFiles),
		staticCache:     staticCache,
		staticFS:        &staticFiles,
		secretMapHolder: secretMapHolder,
	}, nil
}

func (rh *requestHandlerImpl) writeNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("<h1>Whatever you're looking for is not here.<br/> <a href=\"/\">Go back Home</a></h1>"))
}

func (rh *requestHandlerImpl) writeSecretDoesNotExist(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("<h1>Secret does not exist or it was already read. <br/> <a href=\"/\">Go back Home</a></h1>"))
}

func (rh *requestHandlerImpl) cssHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")
	w.Write(rh.staticCache.css)
}

func (rh *requestHandlerImpl) faviconHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Write(rh.staticCache.favicon)
}

func (rh *requestHandlerImpl) getSecretHandler(w http.ResponseWriter, r *http.Request) {
	secret := r.PathValue(secretPathVariableName)
	if secret == "" {
		slog.Error("No secret provided", "path", r.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err := rh.templateCache.getSecret.Execute(w, SecretWrapper{Base64MapKey: secret})
	if err != nil {
		slog.Error("Error executing template", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (rh *requestHandlerImpl) generalHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		rh.templateCache.home.Execute(w, nil)
		return
	}
	rh.writeNotFound(w)
}

func (rh *requestHandlerImpl) createSecretHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	secretValue := r.Form.Get("secretValue")
	base64MapKey, err := rh.secretMapHolder.Set([]byte(secretValue))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/get-secret/"+base64MapKey, http.StatusFound)
}

func (rh *requestHandlerImpl) decryptSecretHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	base64Mapkey := r.Form.Get(decryptionSecretFormFieldName)
	base64Mapkey = strings.TrimSpace(base64Mapkey)
	if base64Mapkey == "" {
		slog.Error("No secret provided")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	usersSecret, err := rh.secretMapHolder.Get(base64Mapkey)
	if err != nil {
		slog.Error("Error getting secret", "error", err)
		rh.writeSecretDoesNotExist(w)
		return
	}
	w.WriteHeader(http.StatusOK)
	usersSecret = bytes.TrimRight(usersSecret, "\x00")
	err = rh.templateCache.showSecret.Execute(w, ShowSecretWrapper{Secret: string(usersSecret)})
	if err != nil {
		slog.Error("Error executing show secret template", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (rh *requestHandlerImpl) RegisterHandlers() {
	http.HandleFunc("GET /"+cssPath, rh.cssHandler)
	http.HandleFunc("GET /"+faviconPath, rh.faviconHandler)
	http.HandleFunc("GET /get-secret/{"+secretPathVariableName+"...}", rh.getSecretHandler)
	http.HandleFunc("GET /", rh.generalHandler)

	http.HandleFunc("POST /create", rh.createSecretHandler)
	http.HandleFunc("POST /decrypt-secret", rh.decryptSecretHandler)
}
