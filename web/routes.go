package web

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

// used to render index.html
type serverInfo struct {
	ApiPollSecs    int
	ExecutablePath string
}

//  Route Handlers

// forward to secrets handler if secrets aren't loaded. Otherwise render index
func (w *WebApp) handleRoot(c echo.Context) error {
	if !w.Sc.SecretsAreLoaded() {
		return c.Redirect(http.StatusSeeOther, "/secrets")
	}
	executablePath, err := os.Executable()
	if err != nil {
		executablePath = "File path not determined"
	}
	serverInfo := serverInfo{
		ApiPollSecs:    w.serverParams.pollRate,
		ExecutablePath: executablePath,
	}
	return c.Render(http.StatusOK, "index.html", serverInfo)
}

// directs user to unlock secrets if encrypted secrets are resent, otherwise directs user to submit secrets
func (w *WebApp) handleSecrets(c echo.Context) error {

	if !w.Sc.SecretsAreLoaded() {
		if w.Sc.EncFilePresent() {
			return c.Render(http.StatusOK, "unlockSecrets.html", nil)
		}
		return c.Render(http.StatusOK, "enterSecrets.html", nil)
	}
	return c.Redirect(http.StatusSeeOther, "/")
}

// handle reception of user-supplied secrets
func (w *WebApp) handleReceiveSecrets(c echo.Context) error {
	var submission submittedSecrets
	if err := c.Bind(&submission); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
	}

	// If secrets file is present, only password is required (unlock mode)
	if w.Sc.EncFilePresent() {
		if submission.Password == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Password is required"})
		}
		err := w.Sc.DecryptSecrets([]byte(submission.Password), 1024)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to decrypt API Key"})
		}
		go w.pollApi()
		return c.Redirect(http.StatusSeeOther, "/")
	}

	// If secrets file is not present, require all secrets and password (entry mode)
	if submission.Username == "" || submission.IntegrationCode == "" || submission.Secret == "" || submission.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "API Key and Password are required"})
	}
	w.Sc.SetSecrets(submission.IntegrationCode, submission.Secret, submission.Username)
	go w.pollApi()
	err := w.Sc.EncryptToDisk([]byte(submission.Password))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save secrets"})
	}
	return c.Redirect(http.StatusSeeOther, "/")
}
