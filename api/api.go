package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"

	"github.com/golang-jwt/jwt/v5"

	"server/model/blacklist"
	"server/model/geo"
	"server/model/users"
)

type Api struct {
	*http.ServeMux
	secret    []byte
	users     *users.Users
	geo       *geo.Geo
	blSubnets *blacklist.Subnets
}

func New(u *users.Users, geo *geo.Geo, blSubnets *blacklist.Subnets) *Api {
	secret, err := base64.StdEncoding.DecodeString("CGWpjarkRIXzCIIw5vXKc+uESy5ebrbOyVMZvftj19k=")
	if err != nil {
		log.Fatal(err)
	}

	asApi := &Api{
		ServeMux:  http.NewServeMux(),
		secret:    secret,
		users:     u,
		geo:       geo,
		blSubnets: blSubnets,
	}

	asApi.HandleFunc("POST /auth", asApi.mwCheckIP(http.HandlerFunc(asApi.auth)))
	asApi.HandleFunc("PUT /user", asApi.mwCheckIP(http.HandlerFunc(asApi.register)))
	asApi.HandleFunc("PATCH /user", asApi.mwCheckIP(asApi.mwAuth(asApi.edit)))
	asApi.HandleFunc("GET /user", asApi.mwCheckIP(asApi.mwAuth(asApi.getUser)))
	asApi.HandleFunc("PUT /blacklist/subnet/{ip}/{mask}", asApi.mwCheckIP(asApi.mwAuth(asApi.blocklistSubnetAdd)))
	asApi.HandleFunc("DELETE /blacklist/subnet/{ip}/{mask}", asApi.mwCheckIP(asApi.mwAuth(asApi.blocklistSubnetDelete)))
	asApi.HandleFunc("PUT /blacklist/user/{login}", asApi.mwCheckIP(asApi.mwAuth(asApi.blocklistUserAdd)))
	asApi.HandleFunc("DELETE /blacklist/user/{login}", asApi.mwCheckIP(asApi.mwAuth(asApi.blocklistUserDelete)))

	return asApi
}

func (a *Api) auth(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
		Nonce    string `json:"nonce"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Print(err)
		return
	}

	user := a.users.Get(req.Login)
	if user == nil || user.Password != req.Password || user.IsBlocked {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	ipCity := a.geo.GetCityByIp(net.ParseIP(r.Header.Get("X-Forwarded-For")))
	if ipCity != nil && ipCity.Country != user.Country {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	claims := struct {
		jwt.RegisteredClaims
		Login string `json:"login"`
		Nonce string `json:"nonce"`
	}{
		Login: user.Login,
		Nonce: req.Nonce,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	strToken, err := token.SignedString(a.secret)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(strToken); err != nil {
		log.Print(err)
	}
}

func (a *Api) getUser(user *users.User, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(struct {
		Login   string `json:"login"`
		Name    string `json:"name"`
		Phone   string `json:"phone"`
		Country string `json:"country"`
		IsAdmin bool   `json:"is_admin,omitempty"`
	}{
		user.Login, user.Name, user.Phone, user.Country, user.IsAdmin,
	}); err != nil {
		log.Print(err)
	}
}

func (a *Api) register(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
		Name     string `json:"name"`
		Phone    string `json:"phone"`
		Country  string `json:"country"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Print(err)
		return
	}

	ipCity := a.geo.GetCityByIp(net.ParseIP(r.Header.Get("X-Forwarded-For")))
	if ipCity != nil && ipCity.Country != req.Country {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if err := a.users.Register(users.User{
		Login:    req.Login,
		Password: req.Password,
		Name:     req.Name,
		Phone:    req.Phone,
		Country:  req.Country,
	}); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Print(err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (a *Api) edit(user *users.User, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req struct {
		Password *string `json:"password"`
		Name     *string `json:"name"`
		Phone    *string `json:"phone"`
		Country  *string `json:"country"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Print(err)
		return
	}

	if err := a.users.Edit(user.Login, req.Password, req.Name, req.Phone, req.Country); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (a *Api) blocklistSubnetAdd(user *users.User, w http.ResponseWriter, r *http.Request) {
	if !user.IsAdmin {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	_, ipNet, err := net.ParseCIDR(r.PathValue("ip") + "/" + r.PathValue("mask"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := a.blSubnets.Add(ipNet); err != nil {
		w.WriteHeader(http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (a *Api) blocklistSubnetDelete(user *users.User, w http.ResponseWriter, r *http.Request) {
	if !user.IsAdmin {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	_, ipNet, err := net.ParseCIDR(r.PathValue("ip") + "/" + r.PathValue("mask"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := a.blSubnets.Delete(ipNet); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *Api) blocklistUserAdd(user *users.User, w http.ResponseWriter, r *http.Request) {
	if !user.IsAdmin {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if err := a.users.Block(r.PathValue("login")); err != nil {
		switch {
		case errors.Is(err, users.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)
		case errors.Is(err, users.ErrAlreadyBlocked):
			w.WriteHeader(http.StatusConflict)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (a *Api) blocklistUserDelete(user *users.User, w http.ResponseWriter, r *http.Request) {
	if !user.IsAdmin {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if err := a.users.Unblock(r.PathValue("login")); err != nil {
		switch {
		case errors.Is(err, users.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
