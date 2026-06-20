package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var safePathRE = regexp.MustCompile(`^[A-Za-z0-9._\-][A-Za-z0-9._\-/]*$`)

func jsonError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	b, _ := json.Marshal(map[string]any{"error": true, "message": message})
	w.Write(b)
}

func validateField(value, field string) string {
	if value == "" {
		return field + " is required"
	}
	if strings.Contains(value, "..") {
		return field + " must not contain '..'"
	}
	if !safePathRE.MatchString(value) {
		return field + " contains invalid characters"
	}
	return ""
}

func listBackupsHandler(w http.ResponseWriter, r *http.Request) {
	subdirectory := strings.TrimSpace(r.URL.Query().Get("subdirectory"))
	provider     := strings.TrimSpace(r.URL.Query().Get("provider"))

	var errs []string
	for _, check := range [][2]string{{"subdirectory", subdirectory}, {"provider", provider}} {
		if msg := validateField(check[1], check[0]); msg != "" {
			errs = append(errs, msg)
		}
	}
	if len(errs) > 0 {
		jsonError(w, strings.Join(errs, "\n"), http.StatusBadRequest)
		return
	}

	target := provider + ":" + subdirectory
	out, err := exec.Command("rclone", "lsjson", target).CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		jsonError(w, "rclone failed: "+msg, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func deleteBackupHandler(w http.ResponseWriter, r *http.Request) {
	name         := strings.TrimSpace(r.URL.Query().Get("name"))
	subdirectory := strings.TrimSpace(r.URL.Query().Get("subdirectory"))
	provider     := strings.TrimSpace(r.URL.Query().Get("provider"))

	var errs []string
	for _, check := range [][2]string{{"name", name}, {"subdirectory", subdirectory}, {"provider", provider}} {
		if msg := validateField(check[1], check[0]); msg != "" {
			errs = append(errs, msg)
		}
	}
	if len(errs) > 0 {
		jsonError(w, strings.Join(errs, "\n"), http.StatusBadRequest)
		return
	}

	target := provider + ":" + filepath.Join(subdirectory, name)
	out, err := exec.Command("rclone", "deletefile", target).CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		jsonError(w, "rclone failed: "+msg, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"ok","deleted":%q}`, target)
}

func backupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		listBackupsHandler(w, r)
		return
	}
	if r.Method == http.MethodDelete {
		deleteBackupHandler(w, r)
		return
	}
	if r.Method != http.MethodPost {
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 32MB in memory; larger spills to OS temp files automatically
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		jsonError(w, "failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	name         := strings.TrimSpace(r.FormValue("name"))
	subdirectory := strings.TrimSpace(r.FormValue("subdirectory"))
	provider     := strings.TrimSpace(r.FormValue("provider"))

	file, _, err := r.FormFile("backup")
	if err != nil {
		jsonError(w, "backup file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	var errs []string
	for _, check := range [][2]string{{"name", name}, {"subdirectory", subdirectory}, {"provider", provider}} {
		if msg := validateField(check[1], check[0]); msg != "" {
			errs = append(errs, msg)
		}
	}
	if len(errs) > 0 {
		jsonError(w, strings.Join(errs, "\n"), http.StatusBadRequest)
		return
	}

	tmp, err := os.CreateTemp("", "backup-*.tar")
	if err != nil {
		jsonError(w, "failed to create temp file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmp.Name())

	if _, err := io.Copy(tmp, file); err != nil {
		tmp.Close()
		jsonError(w, "failed to write upload: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmp.Close()

	destination := provider + ":" + filepath.Join(subdirectory, name)
	out, err := exec.Command("rclone", "copyto", tmp.Name(), destination).CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		jsonError(w, "rclone failed: "+msg, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"ok","destination":%q}`, destination)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/backup", backupHandler)
	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
