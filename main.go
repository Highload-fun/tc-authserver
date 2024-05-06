package main

import (
	"log"
	"net/http"
	"path/filepath"
	"time"

	"server/api"
	"server/model/blacklist"
	"server/model/geo"
	"server/model/users"
)

func main() {
	start := time.Now()
	dataPath := "/storage/data"

	mUsers := users.Load(filepath.Join(dataPath, "users.jsonl"))

	if err := mUsers.Register(users.User{
		Login:    "user_login",
		Password: "password",
	}); err != nil {
		log.Fatal(err)
	}
	log.Printf("Users loaded in %v", time.Now().Sub(start))

	mGeo := geo.New(filepath.Join(dataPath, "/GeoLite2-City-CSV"))
	mBlacklistSubnets := blacklist.NewSubnets()
	log.Printf("Ready in %v", time.Now().Sub(start))
	if err := http.ListenAndServe(":8080", CORSMiddleware(api.New(mUsers, mGeo, mBlacklistSubnets))); err != nil {
		log.Fatal(err)
	}
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With, X-API-Key")

		//for i := 0; i < 100; i++ {
		//	w.Header().Set(fmt.Sprintf("X-Test-Header-%d", i), fmt.Sprintf("Value-%d", i))
		//}

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
