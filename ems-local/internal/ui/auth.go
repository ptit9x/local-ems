package ui

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

// AuthConfig holds authentication credentials.
type AuthConfig struct {
	Username string
	Password string
}

// DefaultAuthConfig returns the default admin credentials.
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		Username: "admin",
		Password: "ems@2025",
	}
}

// sessionStore manages active sessions.
type sessionStore struct {
	mu       sync.RWMutex
	sessions map[string]sessionEntry
	secret   []byte
}

type sessionEntry struct {
	username  string
	expiresAt time.Time
}

func newSessionStore() *sessionStore {
	secret := make([]byte, 32)
	rand.Read(secret)
	return &sessionStore{
		sessions: make(map[string]sessionEntry),
		secret:   secret,
	}
}

func (s *sessionStore) create(username string) string {
	raw := make([]byte, 32)
	rand.Read(raw)
	mac := hmac.New(sha256.New, s.secret)
	mac.Write(raw)
	token := hex.EncodeToString(mac.Sum(nil))

	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[token] = sessionEntry{
		username:  username,
		expiresAt: time.Now().Add(24 * time.Hour),
	}
	return token
}

func (s *sessionStore) validate(token string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.sessions[token]
	if !ok {
		return false
	}
	if time.Now().After(entry.expiresAt) {
		delete(s.sessions, token)
		return false
	}
	return true
}

func (s *sessionStore) destroy(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, token)
}

// authMiddleware wraps an http.Handler with session-based authentication.
func authMiddleware(next http.Handler, cfg AuthConfig, sessions *sessionStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow login page and POST without auth
		if r.URL.Path == "/login" {
			if r.Method == http.MethodPost {
				handleLoginPost(w, r, cfg, sessions)
				return
			}
			serveLoginPage(w, r, "")
			return
		}
		if r.URL.Path == "/logout" {
			handleLogout(w, r, sessions)
			return
		}

		// Check session cookie
		cookie, err := r.Cookie("ems_session")
		if err != nil || !sessions.validate(cookie.Value) {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func handleLoginPost(w http.ResponseWriter, r *http.Request, cfg AuthConfig, sessions *sessionStore) {
	if err := r.ParseForm(); err != nil {
		serveLoginPage(w, r, "Invalid request")
		return
	}

	user := r.FormValue("username")
	pass := r.FormValue("password")

	userHash := sha256.Sum256([]byte(user))
	passHash := sha256.Sum256([]byte(pass))
	expectedUserHash := sha256.Sum256([]byte(cfg.Username))
	expectedPassHash := sha256.Sum256([]byte(cfg.Password))

	userOK := subtle.ConstantTimeCompare(userHash[:], expectedUserHash[:]) == 1
	passOK := subtle.ConstantTimeCompare(passHash[:], expectedPassHash[:]) == 1

	if !userOK || !passOK {
		serveLoginPage(w, r, "Invalid username or password")
		return
	}

	token := sessions.create(user)
	http.SetCookie(w, &http.Cookie{
		Name:     "ems_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400,
	})

	http.Redirect(w, r, "/", http.StatusFound)
}

func handleLogout(w http.ResponseWriter, r *http.Request, sessions *sessionStore) {
	cookie, err := r.Cookie("ems_session")
	if err == nil {
		sessions.destroy(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "ems_session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/login", http.StatusFound)
}

func serveLoginPage(w http.ResponseWriter, _ *http.Request, errMsg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(loginHTML(errMsg)))
}

func loginHTML(errMsg string) string {
	errBlock := ""
	if errMsg != "" {
		errBlock = `<div class="error">` + errMsg + `</div>`
	}

	return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Local EMS — Sign In</title>
<style>
  * { margin:0; padding:0; box-sizing:border-box; }
  body { font-family:-apple-system,'Segoe UI',Roboto,sans-serif; background:#0f172a; color:#e2e8f0;
    min-height:100vh; display:flex; align-items:center; justify-content:center; }

  .login-container { width:100%; max-width:400px; padding:24px; }

  .login-card { background:#1e293b; border:1px solid #334155; border-radius:16px; padding:40px 32px; }

  .logo { text-align:center; margin-bottom:32px; }
  .logo .icon { font-size:48px; margin-bottom:8px; }
  .logo h1 { font-size:20px; font-weight:600; color:#e2e8f0; }
  .logo p { font-size:13px; color:#64748b; margin-top:4px; }

  .form-group { margin-bottom:20px; }
  .form-group label { display:block; font-size:13px; font-weight:500; color:#94a3b8; margin-bottom:6px; }
  .form-group input { width:100%; padding:12px 14px; background:#0f172a; border:1px solid #334155;
    border-radius:8px; color:#e2e8f0; font-size:14px; outline:none; transition:border-color .2s; }
  .form-group input:focus { border-color:#facc15; }
  .form-group input::placeholder { color:#475569; }

  .submit-btn { width:100%; padding:12px; background:#facc15; color:#0f172a; border:none;
    border-radius:8px; font-size:14px; font-weight:600; cursor:pointer; transition:opacity .2s; margin-top:4px; }
  .submit-btn:hover { opacity:.9; }
  .submit-btn:active { transform:scale(.98); }

  .error { background:#dc262622; border:1px solid #dc262644; color:#f87171;
    padding:10px 14px; border-radius:8px; font-size:13px; margin-bottom:20px; text-align:center; }

  .footer { text-align:center; margin-top:24px; font-size:11px; color:#475569; }
</style>
</head>
<body>
<div class="login-container">
  <div class="login-card">
    <div class="logo">
      <div class="icon">⚡</div>
      <h1>Local EMS</h1>
      <p>Energy Management System — 5MWh BESS</p>
    </div>

    ` + errBlock + `

    <form method="POST" action="/login">
      <div class="form-group">
        <label for="username">Username</label>
        <input type="text" id="username" name="username" placeholder="Enter username" autocomplete="username" required autofocus>
      </div>
      <div class="form-group">
        <label for="password">Password</label>
        <input type="password" id="password" name="password" placeholder="Enter password" autocomplete="current-password" required>
      </div>
      <button type="submit" class="submit-btn">Sign In</button>
    </form>
  </div>
  <div class="footer">Local EMS v0.2.0 • Simulation Mode</div>
</div>
</body>
</html>`
}
