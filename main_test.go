package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/eirka/eirka-libs/config"
	"github.com/eirka/eirka-libs/csrf"
	"github.com/eirka/eirka-libs/db"
	"github.com/eirka/eirka-libs/user"
)

// publicAllowlist is the set of "METHOD /path" routes that are intentionally
// reachable without authentication or a moderator role. Every other route built
// by setupRouter() must be denied to anonymous and to non-moderator callers.
//
// Adding an entry here is a deliberate, reviewable decision: a new route that is
// NOT behind the admin group will otherwise fail TestRouteSecurity below.
var publicAllowlist = map[string]bool{
	"GET /status": true,
}

// isAuthorizedQuery matches the role lookup run by user.Protect -> IsAuthorized
// (eirka-libs/user/user.go). go-sqlmock matches the expected SQL as a regexp.
const isAuthorizedQuery = `SELECT COALESCE\(\(SELECT MAX\(role_id\) FROM user_ib_role_map`

// TestRouteSecurity is a regression guard for the two invariants of this service:
// every endpoint requires authentication, and every endpoint requires a moderator
// (role 3) or admin (role 4). It enumerates the real router and asserts that each
// route outside publicAllowlist denies both an anonymous request and an
// authenticated regular-user (role 2) request with 403.
func TestRouteSecurity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Make the JWT layer usable: a short test secret + relaxed length validation.
	user.SetTestMode(true)
	config.Settings.Session.NewSecret = "testsecret"

	router := setupRouter()

	// A valid CSRF token pair, minted through the real csrf.Cookie middleware, so
	// state-changing (POST/DELETE) requests get past csrf.Verify and actually
	// exercise the auth/role gates rather than being rejected by CSRF.
	csrfCookie, csrfToken := csrfPair(t)

	seenAllowlisted := map[string]bool{}

	for _, rt := range router.Routes() {
		key := rt.Method + " " + rt.Path

		if publicAllowlist[key] {
			seenAllowlisted[key] = true
			continue
		}

		path := concretePath(rt.Path)

		// 1) Anonymous (no session cookie) must be denied by user.Auth(true).
		t.Run("anonymous denied: "+key, func(t *testing.T) {
			req, _ := http.NewRequest(rt.Method, path, nil)
			attachCSRF(req, rt.Method, csrfCookie, csrfToken)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusForbidden, w.Code,
				"anonymous request to %q must be rejected with 403 (auth required)", key)
		})

		// 2) Authenticated regular user (role 2) must be denied by user.Protect.
		t.Run("non-moderator denied: "+key, func(t *testing.T) {
			mock, err := db.NewTestDb()
			require.NoError(t, err)
			defer db.CloseDb()

			// IsAuthorized looks up the caller's role for board (ib) = 1.
			mock.ExpectQuery(isAuthorizedQuery).
				WithArgs(1, 2).
				WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow(2))

			req, _ := http.NewRequest(rt.Method, path, nil)
			req.AddCookie(jwtCookie(t, 2))
			attachCSRF(req, rt.Method, csrfCookie, csrfToken)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusForbidden, w.Code,
				"role-2 request to %q must be rejected with 403 (moderator role required)", key)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}

	// Every allowlisted route must actually be registered, so the allowlist can't
	// silently rot and start excusing a route that no longer exists.
	for key := range publicAllowlist {
		assert.True(t, seenAllowlisted[key],
			"allowlisted public route %q is not registered by setupRouter()", key)
	}
}

// TestModeratorReachesEndpoint is the positive control: it proves the role gate
// admits a moderator (role 3) rather than denying everyone. It uses GET
// /statistics/:ib, whose controller touches only the database. With the role
// lookup returning 3, user.Protect passes and the request reaches the controller
// (which then errors on the unmocked downstream query) — the key point is that the
// response is NOT a 403/401 from the auth/role gates.
func TestModeratorReachesEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	user.SetTestMode(true)
	config.Settings.Session.NewSecret = "testsecret"

	router := setupRouter()

	mock, err := db.NewTestDb()
	require.NoError(t, err)
	defer db.CloseDb()

	mock.ExpectQuery(isAuthorizedQuery).
		WithArgs(1, 3).
		WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow(3))

	req, _ := http.NewRequest("GET", "/statistics/1", nil)
	req.AddCookie(jwtCookie(t, 3))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusForbidden, w.Code,
		"a moderator (role 3) must not be blocked by the auth/role gates")
	assert.NotEqual(t, http.StatusUnauthorized, w.Code,
		"a moderator (role 3) must not be blocked by the auth/role gates")
}

// concretePath replaces every ":param" path segment with "1" so a route template
// such as /post/:ib/:thread/:id becomes a requestable /post/1/1/1.
func concretePath(p string) string {
	parts := strings.Split(p, "/")
	for i, seg := range parts {
		if strings.HasPrefix(seg, ":") || strings.HasPrefix(seg, "*") {
			parts[i] = "1"
		}
	}
	return strings.Join(parts, "/")
}

// attachCSRF adds a valid CSRF cookie + header to state-changing requests so they
// survive csrf.Verify; GET requests are skipped by the CSRF middleware.
func attachCSRF(req *http.Request, method string, cookie *http.Cookie, token string) {
	if method == http.MethodGet {
		return
	}
	req.AddCookie(cookie)
	req.Header.Set(csrf.HeaderName, token)
}

// jwtCookie returns a session cookie holding a valid JWT for the given user id.
func jwtCookie(t *testing.T, uid uint) *http.Cookie {
	t.Helper()
	tok, err := user.MakeToken(uid)
	require.NoError(t, err)
	return &http.Cookie{Name: user.CookieName, Value: tok}
}

// csrfPair mints a matching (cookie, sent-token) pair using the real csrf.Cookie
// middleware, which is the same mechanism that issues tokens in production.
func csrfPair(t *testing.T) (*http.Cookie, string) {
	t.Helper()

	g := gin.New()
	g.Use(csrf.Cookie())
	g.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, c.MustGet("csrf_token").(string))
	})

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, req)

	var cookie *http.Cookie
	for _, ck := range w.Result().Cookies() {
		if ck.Name == csrf.CookieName {
			cookie = ck
		}
	}
	require.NotNil(t, cookie, "csrf.Cookie should set the %q cookie", csrf.CookieName)

	return cookie, w.Body.String()
}
