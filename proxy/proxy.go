package proxy

import (
	"crypto/subtle"
	"net/http"

	"github.com/auth0/go-jwt-middleware"
	"github.com/foomo/neosproxy/logging"
)

//-----------------------------------------------------------------------------
// ~ Constants
//-----------------------------------------------------------------------------

const neosproxyPath = "/neosproxy"
const routeContentServerExport = "/contentserver/export"

//-----------------------------------------------------------------------------
// ~ Public methods
//-----------------------------------------------------------------------------

// Run a proxy
func (p *Proxy) Run() error {
	return http.ListenAndServe(p.config.Proxy.Address, p.router)
}

//-----------------------------------------------------------------------------
// ~ Private methods
//-----------------------------------------------------------------------------

func (p *Proxy) setupRoutes() {

	// hijack content server export routes
	p.router.HandleFunc(routeContentServerExport, p.streamCachedNeosContentServerExport)
	p.router.HandleFunc(routeContentServerExport, p.streamCachedNeosContentServerExport).Queries("workspace", "{workspace}")

	// /contentserver/export/de/571fd1ae-c8e4-4d91-a708-d97025fb015c?workspace=stage
	p.router.HandleFunc(routeContentServerExport+"/{dimension}/{id}", p.getContent)
	p.router.HandleFunc(routeContentServerExport+"/{dimension}/{id}", p.getContent).Queries("workspace", "{workspace}")

	// api
	// neosproxy/cache/%s?workspace=%s
	neosproxyRouter := p.router.PathPrefix(neosproxyPath).Subrouter()
	neosproxyRouter.Use(p.middlewareTokenAuth)
	neosproxyRouter.HandleFunc("/cache/{id}", p.invalidateCache).Methods(http.MethodDelete)
	neosproxyRouter.HandleFunc("/cache/{id}", p.invalidateCache).Methods(http.MethodDelete).Queries("workspace", "{workspace}").Name("api-delete-cache")

	// middlewares
	p.router.Use(p.middlewareServiceUnavailable)

	// error handling
	p.router.NotFoundHandler = http.HandlerFunc(p.notFound)
	p.router.MethodNotAllowedHandler = http.HandlerFunc(p.methodNotAllowed)

	// fallback to proxy
	p.router.PathPrefix("/").Handler(p.proxyHandler)
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

func (p *Proxy) middlewareServiceUnavailable(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// service unavailable
		if p.maintenance {
			p.serviceNotAvailable(w, r)
			return
		}

		// call next handler
		next.ServeHTTP(w, r)
	})
}

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
