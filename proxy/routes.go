package proxy

import "net/http"

//-----------------------------------------------------------------------------
// ~ Constants
//-----------------------------------------------------------------------------

const neosproxyPath = "/neosproxy"
const routeContentServerExport = "/contentserver/export"

//-----------------------------------------------------------------------------
// ~ Private methods
//-----------------------------------------------------------------------------

func (p *Proxy) setupRoutes() {

	// hijack content server export routes

	// content tree / sitemap
	p.router.HandleFunc(routeContentServerExport, p.streamCachedNeosContentServerExport)
	p.router.HandleFunc(routeContentServerExport, p.streamCachedNeosContentServerExport).Queries("workspace", "{workspace}")

	// etag
	p.router.HandleFunc(routeContentServerExport+"/etag/{dimension}/{id}", p.getEtagByID).Methods(http.MethodGet)
	p.router.HandleFunc(routeContentServerExport+"/etag/{dimension}/{id}", p.getEtagByID).Methods(http.MethodGet).Queries("workspace", "{workspace}")
	p.router.HandleFunc(routeContentServerExport+"/etag/{hash}", p.getEtagByHash).Methods(http.MethodGet)

	p.router.HandleFunc(routeContentServerExport+"/etags", p.getAllEtags).Methods(http.MethodGet)
	p.router.HandleFunc(routeContentServerExport+"/etags", p.getAllEtags).Methods(http.MethodGet).Queries("workspace", "{workspace}")

	// documents => /contentserver/export/de/571fd1ae-c8e4-4d91-a708-d97025fb015c?workspace=stage
	p.router.HandleFunc(routeContentServerExport+"/{dimension}/{id}", p.getContent).Methods(http.MethodGet)
	p.router.HandleFunc(routeContentServerExport+"/{dimension}/{id}", p.getContent).Methods(http.MethodGet).Queries("workspace", "{workspace}")

	p.router.HandleFunc(routeContentServerExport+"/{dimension}/{id}", p.getEtagByID).Methods(http.MethodHead)
	p.router.HandleFunc(routeContentServerExport+"/{dimension}/{id}", p.getEtagByID).Methods(http.MethodHead).Queries("workspace", "{workspace}")

	// api
	// neosproxy/cache/%s?workspace=%s
	neosproxyRouter := p.router.PathPrefix(neosproxyPath).Subrouter()
	neosproxyRouter.Use(p.middlewareTokenAuth)
	neosproxyRouter.HandleFunc("/cache/{id}", p.invalidateCache).Methods(http.MethodDelete)
	neosproxyRouter.HandleFunc("/cache/{id}", p.invalidateCache).Methods(http.MethodDelete).Queries("workspace", "{workspace}").Name("api-delete-cache")
	neosproxyRouter.HandleFunc("/status", p.streamStatus).Methods(http.MethodGet)

	// error handling
	p.router.NotFoundHandler = http.HandlerFunc(p.notFound)
	p.router.MethodNotAllowedHandler = http.HandlerFunc(p.methodNotAllowed)

	// fallback to proxy
	p.router.PathPrefix("/").Handler(p.proxyHandler)
}
