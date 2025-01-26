package main

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

type app struct {
	templateCache   *templateCache
	staticFS        *embed.FS
	secretMapHolder secretmap.SecretMapHolder
}

type SecretWrapper struct {
	Base64MapKey string
}

type ShowSecretWrapper struct {
	Secret string
}

func createTemplateCache(staticFS fs.FS) *templateCache {
	homeTemplate := template.Must(template.ParseFS(staticFS, "static/index.html"))
	getSecretTemplate := template.Must(template.ParseFS(staticFS, "static/get-secret.html"))
	showSecret := template.Must(template.ParseFS(staticFS, "static/show-secret.html"))
	return &templateCache{home: homeTemplate, getSecret: getSecretTemplate, showSecret: showSecret}
}

func newApp(staticFilesFs *embed.FS) *app {
	return &app{templateCache: createTemplateCache(staticFilesFs), staticFS: staticFilesFs, secretMapHolder: secretmap.NewSecretMapHolder()}
}

func (a *app) writeNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("<h1>Whatever you're looking for is not here.<br/> <a href=\"/\">Go back Home</a></h1>"))
}

func (a *app) writeSecretDoesNotExist(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("<h1>Secret does not exist or it was already read. <br/> <a href=\"/\">Go back Home</a></h1>"))
}

func (a *app) generalHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		a.templateCache.home.Execute(w, nil)
		return
	}
	slog.Error("Responding with not found as no handler matches request")
	a.writeNotFound(w)
}

func (a *app) cssHandler(w http.ResponseWriter, r *http.Request) {
	b, err := staticFiles.ReadFile(cssPath)
	if err != nil {
		slog.Error("Error reading css file", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/css")
	w.Write(b)
}

func (a *app) faviconHandler(w http.ResponseWriter, r *http.Request) {
	b, err := staticFiles.ReadFile(faviconPath)
	if err != nil {
		slog.Error("Error reading favicon file", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Write(b)
}

func (a *app) createSecretHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	secretValue := r.Form.Get("secretValue")
	base64MapKey, err := a.secretMapHolder.Set([]byte(secretValue))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/get-secret/"+base64MapKey, http.StatusFound)
}
func (a *app) getSecretHandler(w http.ResponseWriter, r *http.Request) {
	secret := r.PathValue(secretPathVariableName)
	if secret == "" {
		slog.Error("No secret provided")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err := a.templateCache.getSecret.Execute(w, SecretWrapper{Base64MapKey: secret})
	if err != nil {
		slog.Error("Error executing template", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (a *app) decryptSecretHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	base64Mapkey := r.Form.Get(decryptionSecretFormFieldName)
	base64Mapkey = strings.TrimSpace(base64Mapkey)
	if base64Mapkey == "" {
		slog.Error("No secret provided")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	usersSecret, err := a.secretMapHolder.Get(base64Mapkey)
	if err != nil {
		slog.Error("Error getting secret", "error", err)
		a.writeSecretDoesNotExist(w)
		return
	}
	w.WriteHeader(http.StatusOK)
	usersSecret = bytes.TrimRight(usersSecret, "\x00")
	err = a.templateCache.showSecret.Execute(w, ShowSecretWrapper{Secret: string(usersSecret)})
	if err != nil {
		slog.Error("Error executing show secret template", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func main() {
	slog.Info("Parsing templates")
	a := newApp(&staticFiles)
	slog.Info("Application starting on port :8080")
	http.HandleFunc("GET /"+cssPath, a.cssHandler)
	http.HandleFunc("GET /"+faviconPath, a.faviconHandler)
	http.HandleFunc("GET /get-secret/{"+secretPathVariableName+"...}", a.getSecretHandler)
	http.HandleFunc("GET /", a.generalHandler)
	http.HandleFunc("POST /create", a.createSecretHandler)
	http.HandleFunc("POST /decrypt-secret", a.decryptSecretHandler)
	slog.Error("Server crashed", "error", http.ListenAndServe(":8080", nil))
}
