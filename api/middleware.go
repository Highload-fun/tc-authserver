package api

import (
	"log"
	"net"
	"net/http"

	"github.com/golang-jwt/jwt/v5"

	"server/model/users"
)

type AuthenticatedHandler func(user *users.User, w http.ResponseWriter, r *http.Request)

func (a *Api) mwCheckIP(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := net.ParseIP(r.Header.Get("X-Forwarded-For"))
		if a.blSubnets.CheckIp(ip) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		h.ServeHTTP(w, r)
	}
}

func (a *Api) mwAuth(h AuthenticatedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var claims struct {
			jwt.RegisteredClaims
			Login string `json:"login"`
		}
		if _, err := jwt.ParseWithClaims(r.Header.Get("X-API-Key"), &claims, func(token *jwt.Token) (interface{}, error) {
			return a.secret, nil
		}); err != nil {
			log.Printf("%v", err)
			w.WriteHeader(http.StatusForbidden)
			return
		}

		user := a.users.Get(claims.Login)
		if user == nil || user.IsBlocked {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		ipCity := a.geo.GetCityByIp(net.ParseIP(r.Header.Get("X-Forwarded-For")))
		if ipCity != nil && ipCity.Country != user.Country {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		h(user, w, r)
	}
}

func (a *Api) mwCheckRegion(h AuthenticatedHandler) AuthenticatedHandler {
	return func(user *users.User, w http.ResponseWriter, r *http.Request) {

	}
}
