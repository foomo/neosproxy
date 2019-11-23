package proxy

import (
	"crypto/subtle"
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/foomo/neosproxy/logging"
)

//-----------------------------------------------------------------------------
// ~ Public methods
//-----------------------------------------------------------------------------

// Run a proxy
func (p *Proxy) Run() error {
	return http.ListenAndServe(p.config.Proxy.Address, p.router)
}

//-----------------------------------------------------------------------------
// ~ Error handler
//-----------------------------------------------------------------------------

func (p *Proxy) error(w http.ResponseWriter, r *http.Request, statusCode int, msg string) {
	p.log.WithField(logging.FieldURI, r.RequestURI).WithField(logging.FieldHTTPStatusCode, statusCode).Warn(msg)
	w.WriteHeader(statusCode)
	w.Write([]byte(msg + "\n"))
}

func (p *Proxy) serviceNotAvailable(w http.ResponseWriter, r *http.Request) {
	p.error(w, r, http.StatusServiceUnavailable, "service not available")
}

func (p *Proxy) notFound(w http.ResponseWriter, r *http.Request) {
	p.error(w, r, http.StatusNotFound, "not found")
}

func (p *Proxy) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	p.error(w, r, http.StatusMethodNotAllowed, "method not allowed")
}

//-----------------------------------------------------------------------------
// ~ Middleware
//-----------------------------------------------------------------------------

func (p *Proxy) setupLogger(r *http.Request, method string) logging.Entry {
	return p.log.WithField(logging.FieldURI, r.RequestURI).WithField(logging.FieldFunction, method)
}

func (p *Proxy) middlewareTokenAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		log := p.setupLogger(r, "middlewareTokenAuth")
		realm := "API secured with bearer auth"

		// unauthorised helper func
		unauthorised := func() {
			w.Header().Set("WWW-Authenticate", `Bearer realm="`+realm+`"`)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("unauthorised\n"))
		}

		// bearer authentication
		bearer, errBearer := jwtmiddleware.FromAuthHeader(r)
		if errBearer != nil {
			log.WithError(errBearer).Error("bearer authentication error")
			unauthorised()
			return
		}
		if bearer == "" {
			log.Info("bearer authentication must be set")
			unauthorised()
			return
		}
		if bearer != p.config.Proxy.Token {
			log.Warn("bearer authentication mismatch")
			unauthorised()
			return
		}

		// call next handler
		next.ServeHTTP(w, r)
	})
}

func (p *Proxy) middlewareBasicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		realm := "API secured with basic auth"

		// unauthorised helper func
		unauthorised := func() {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("unauthorised\n"))
		}

		// basic auth
		user, pass, ok := r.BasicAuth()

		// no basic auth provided
		if !ok {
			unauthorised()
			return
		}

		// check all available user/password combinations
		match := false
		for _, auth := range p.basicAuth {
			if subtle.ConstantTimeCompare([]byte(user), []byte(auth.user)) == 1 && subtle.ConstantTimeCompare([]byte(pass), []byte(auth.password)) == 1 {
				match = true
				break
			}
		}

		// invalid basic auth credentials
		if !match {
			unauthorised()
			return
		}

		// call next handler
		next.ServeHTTP(w, r)
	})
}
