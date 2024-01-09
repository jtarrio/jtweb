package web

import (
	"net/http"

	auth "github.com/abbot/go-http-auth"
)

const adminUser = "admin"

type AdminChecker struct {
	digest *auth.DigestAuth
}

func NewAdminChecker(secret string) *AdminChecker {
	return &AdminChecker{
		digest: auth.NewDigestAuthenticator("Comments admin", func(user, realm string) string {
			if user == adminUser {
				return secret
			}
			return ""
		}),
	}
}

func (c *AdminChecker) RequiringAdmin(handler http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if c.RequireAdmin(rw, req) {
			handler(rw, req)
		}
	}
}

func (c *AdminChecker) HasAdmin(req *http.Request) bool {
	user, _ := c.digest.CheckAuth(req)
	return user == adminUser
}

func (c *AdminChecker) RequireAdmin(rw http.ResponseWriter, req *http.Request) bool {
	if c.HasAdmin(req) {
		return true
	}
	c.digest.RequireAuth(rw, req)
	return false
}
