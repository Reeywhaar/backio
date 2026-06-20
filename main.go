package main

import (
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

func backupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 32MB in memory; larger spills to OS temp files automatically
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	name         := strings.TrimSpace(r.FormValue("name"))
	subdirectory := strings.TrimSpace(r.FormValue("subdirectory"))
	provider     := strings.TrimSpace(r.FormValue("provider"))

	file, _, err := r.FormFile("backup")
	if err != nil {
		http.Error(w, "backup file is required", http.StatusBadRequest)
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
		http.Error(w, strings.Join(errs, "\n"), http.StatusBadRequest)
		return
	}

	tmp, err := os.CreateTemp("", "backup-*.tar")
	if err != nil {
		http.Error(w, "failed to create temp file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmp.Name())

	if _, err := io.Copy(tmp, file); err != nil {
		tmp.Close()
		http.Error(w, "failed to write upload: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmp.Close()

	destination := provider + ":" + filepath.Join(subdirectory, name)
	out, err := exec.Command("rclone", "copyto", tmp.Name(), destination).CombinedOutput()
	if err != nil {
		http.Error(w, fmt.Sprintf("rclone failed:\n%s", out), http.StatusInternalServerError)
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
