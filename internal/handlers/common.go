package handlers

import (
    "net/http"
)

// GetUsernameFromCookie extracts username from session cookie
func GetUsernameFromCookie(r *http.Request) string {
    cookie, err := r.Cookie("spide_session")
    if err != nil {
        return ""
    }
    return cookie.Value
}




