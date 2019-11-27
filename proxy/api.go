package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/foomo/neosproxy/cache/content/store"

	"github.com/cloudfoundry/bytefmt"
	"github.com/foomo/neosproxy/cache"
	"github.com/foomo/neosproxy/client/cms"
	"github.com/foomo/neosproxy/logging"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	content_cache "github.com/foomo/neosproxy/cache/content"
)

type mime string

const (
	mimeTextPlain       mime = "text/plain"
	mimeApplicationJSON mime = "application/json"
)

// ------------------------------------------------------------------------------------------------
// ~ Proxy handler methods
// ------------------------------------------------------------------------------------------------

func (p *Proxy) getContent(w http.ResponseWriter, r *http.Request) {

	// duration
	start := time.Now()

	// extract request data
	id := getRequestParameter(r, "id")
	dimension := getRequestParameter(r, "dimension")
	workspace := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("workspace")))

	// validate workspace
	if workspace == "" {
		workspace = cms.WorkspaceLive
	}

	// logger
	log := p.setupLogger(r, "getContent").WithFields(logrus.Fields{
		logging.FieldWorkspace: workspace,
		logging.FieldID:        id,
	})

	// etag cache handling
	headerEtag := r.Header.Get("ETag")
	if headerEtag != "" {
		etag, errEtag := p.contentCache.GetEtag(store.GetHash(id, dimension, workspace))
		if errEtag == nil && etag != "" && etag == headerEtag {
			w.WriteHeader(http.StatusNotModified)
			log.WithDuration(start).Debug("content not modified")
			return
		}
	}

	// try cache hit, invalidate in case of item not found
	item, errCacheGet := p.contentCache.Get(id, dimension, workspace)
	if errCacheGet != nil {

		if errCacheGet != content_cache.ErrorNotFound {
			w.WriteHeader(http.StatusInternalServerError)
			log.WithError(errCacheGet).Error("get cached content failed")
			return
		}

		// invalidate content
		startInvalidation := time.Now()
		itemInvalidated, errCacheInvalidate := p.contentCache.Load(id, dimension, workspace)
		if errCacheInvalidate != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.WithError(errCacheInvalidate).Error("serving uncached item failed")
			return
		}
		log.WithDuration(startInvalidation).WithField("len", p.contentCache.Len()).Debug("invalidated content item")

		item = itemInvalidated
	}

	// prepare response data
	data := &cms.Content{
		HTML:              item.HTML,
		CacheDependencies: item.Dependencies,
	}

	w.Header().Set("ETag", item.GetEtag())

	// stream json response
	encoder := json.NewEncoder(w)
	errEncode := encoder.Encode(data)
	if errEncode != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.WithError(errEncode).Error("json encoding failed")
		return
	}

	// done
	// log.WithDuration(start).Debug("content served")
	p.servedStatsChan <- true
	return
}

// invalidateCache will invalidate all cached contentserver export files
func (p *Proxy) invalidateCacheAll(w http.ResponseWriter, r *http.Request) {
	// extract request data
	workspace := strings.TrimSpace(
		strings.ToLower(r.URL.Query().Get("workspace")),
	)
	user := r.Header.Get("X-User")

	// validate workspace
	if workspace == "" {
		workspace = cms.WorkspaceLive
	}

	// logger
	log := p.setupLogger(r, "invalidateCacheAll").WithFields(logrus.Fields{
		logging.FieldWorkspace: workspace,
		"user":                 user,
	})

	cachedItems, err := p.contentCache.GetAll()
	if err != nil {
		log.WithError(err).Error("")
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
	}

	if len(p.config.Neos.Dimensions) == 0 {
		log.Warn("no neos dimension configured")
	}

	for _, ci := range cachedItems {
		p.contentCache.Invalidate(ci.ID, ci.Dimension, ci.Workspace)

		// load workspace worker
		workspaceCache, workspaceOK := p.workspaceCaches[ci.Workspace]
		if !workspaceOK {
			log.Warn("unknown workspace")
			http.Error(
				w,
				"cache invalidation failed: unknown workspace",
				http.StatusBadRequest,
			)
			return
		}

		// add invalidation request to queue (contentserver export)
		workspaceCache.Invalidate()
	}

	w.WriteHeader(http.StatusAccepted)
	msg := fmt.Sprintf(
		"%d cache invalidation requests accepted",
		len(cachedItems),
	)
	w.Write([]byte(msg))
	log.Debug(msg)
}

// invalidateCache will invalidate cached contentserver export file
func (p *Proxy) invalidateCache(w http.ResponseWriter, r *http.Request) {

	// extract request data
	workspace := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("workspace")))
	user := r.Header.Get("X-User")
	id := getRequestParameter(r, "id")

	// validate workspace
	if workspace == "" {
		workspace = cms.WorkspaceLive
	}

	// logger
	log := p.setupLogger(r, "invalidateCache").WithFields(logrus.Fields{
		logging.FieldWorkspace: workspace,
		logging.FieldID:        id,
		"user":                 user,
	})
	log.Debug("cache invalidation request")

	// invalidate all workspaces in case of "live" workspace
	workspaces := []string{workspace}
	if workspace == cms.WorkspaceLive {
		workspaces = []string{}
		for workspace := range p.workspaceCaches {
			workspaces = append(workspaces, workspace)
		}
	}

	if len(p.config.Neos.Dimensions) == 0 {
		log.Warn("no neos dimension configured")
	}

	for _, workspace := range workspaces {

		for _, dimension := range p.config.Neos.Dimensions {
			// add invalidation request / job / task
			p.contentCache.Invalidate(id, dimension, workspace)
		}

		// load workspace worker
		workspaceCache, workspaceOK := p.workspaceCaches[workspace]
		if !workspaceOK {
			log.Warn("unknown workspace")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("cache invalidation failed: unknown workspace"))
			return
		}

		// add invalidation request to queue (contentserver export)
		workspaceCache.Invalidate()
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("cache invalidation request accepted"))
	log.Debug("cache invalidation request accepted")
}

// streamCachedNeosContentServerExport will stream contentserver export
func (p *Proxy) streamCachedNeosContentServerExport(w http.ResponseWriter, r *http.Request) {

	// duration
	start := time.Now()

	// extract request data
	workspace := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("workspace")))

	// validate workspace
	if workspace == "" {
		workspace = cms.WorkspaceLive
	}

	// logger
	log := p.setupLogger(r, "streamCachedNeosContentServerExport").WithField(logging.FieldWorkspace, workspace)

	workspaceCache, workspaceWorkerOK := p.workspaceCaches[workspace]
	if !workspaceWorkerOK {
		log.Error("workspace worker not found")
		p.error(w, r, http.StatusBadRequest, "workspace worker not found")
		return
	}

	workspaceCache.FileLock.RLock()
	defer workspaceCache.FileLock.RUnlock()

	// open file
	file, fileInfo, errFile := workspaceCache.GetContentServerExport()
	if errFile != nil {

		if errFile == cache.ErrorFileNotExists {
			workspaceCache.Invalidate()
			log.WithError(errFile).Error("cached contentserver export: cache empty, invalidation triggered")
			p.error(w, r, http.StatusConflict, "cache empty; cache invalidation triggered; please try again later")
			return
		}

		log.WithError(errFile).Error("cached contentserver export: read file failed")
		p.error(w, r, http.StatusInternalServerError, "cached contentserver export: read file failed")
		return
	}
	defer file.Close()

	// set header
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Last-Modified", fileInfo.ModTime().Format(http.TimeFormat))
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// stream file
	written, errFileStreaming := io.Copy(w, file)
	if errFileStreaming != nil {
		log.WithError(errFileStreaming).WithField("size", bytefmt.ByteSize(uint64(written))).Error("cached contentserver export: file stream failed")
		p.error(w, r, http.StatusInternalServerError, "cached contentserver export: file stream failed")
		return
	}

	// log stats
	log.WithDuration(start).WithField("size", bytefmt.ByteSize(uint64(written))).Info("streamed file")
}

func (p *Proxy) streamStatus(w http.ResponseWriter, r *http.Request) {

	// logger
	log := p.setupLogger(r, "status")

	// stream
	var errEncode error
	contentNegotioation := parseAcceptHeader(r.Header.Get("accept"))
	switch contentNegotioation {
	case mimeApplicationJSON:
		w.Header().Set("Content-Type", string(mimeApplicationJSON))
		encoder := json.NewEncoder(w)
		errEncode = encoder.Encode(p.status)
	case mimeTextPlain:
		w.Header().Set("Content-Type", "application/x-yaml")
		encoder := yaml.NewEncoder(w)
		errEncode = encoder.Encode(p.status)
	}

	// error handling
	if errEncode != nil {
		log.WithError(errEncode).WithField("content-negotiation", contentNegotioation).Error("failed streaming status")
		http.Error(w, "failed streaming status", http.StatusInternalServerError)
		return
	}

}

func (p *Proxy) getAllEtags(w http.ResponseWriter, r *http.Request) {

	// extract request data
	workspace := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("workspace")))

	// validate workspace
	if workspace == "" {
		workspace = cms.WorkspaceLive
	}

	// logger
	log := p.setupLogger(r, "getAllEtags").WithField(logging.FieldWorkspace, workspace)

	etags := p.contentCache.GetAllEtags(workspace)

	w.Header().Set("Content-Type", string(mimeApplicationJSON))
	encoder := json.NewEncoder(w)
	errEncode := encoder.Encode(etags)

	// error handling
	if errEncode != nil {
		log.WithError(errEncode).Error("failed encoding etags")
		http.Error(w, "failed encoding etags", http.StatusInternalServerError)
		return
	}

	return
}

func (p *Proxy) getEtag(w http.ResponseWriter, r *http.Request, hash string) {
	// logger
	log := p.setupLogger(r, "getEtag").WithField(logging.FieldID, hash)

	etag, errEtag := p.contentCache.GetEtag(hash)

	// error handling
	if errEtag != nil {
		log.WithError(errEtag).Error("failed getting etag")
		http.Error(w, "failed getting etag", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", string(mimeTextPlain))
	w.Header().Set("ETag", etag)
	w.Write([]byte(etag))

	return
}

func (p *Proxy) getEtagByID(w http.ResponseWriter, r *http.Request) {
	// extract request data
	workspace := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("workspace")))

	// validate workspace
	if workspace == "" {
		workspace = cms.WorkspaceLive
	}

	// extract request data
	id := getRequestParameter(r, "id")
	dimension := getRequestParameter(r, "dimension")

	hash := store.GetHash(id, dimension, workspace)

	p.getEtag(w, r, hash)
}

func (p *Proxy) getEtagByHash(w http.ResponseWriter, r *http.Request) {

	// extract request data
	hash := getRequestParameter(r, "hash")

	p.getEtag(w, r, hash)
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func getRequestParameter(r *http.Request, parameter string) string {
	return getParameter(mux.Vars(r), parameter)
}

func getParameter(m map[string]string, key string) string {
	if val, ok := m[key]; ok {
		return val
	}
	return ""
}

func parseAcceptHeader(accept string) mime {
	mimes := strings.Split(accept, ",")
	for _, mime := range mimes {
		values := strings.Split(mime, ";")

		switch values[0] {
		case string(mimeApplicationJSON):
			return mimeApplicationJSON
		case string(mimeTextPlain):
			return mimeTextPlain
		}
	}
	return mimeApplicationJSON
}
