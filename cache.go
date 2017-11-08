package neosproxy

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

const DefaultWorkspace = "live"

// invalidateCache ...
func (p *Proxy) invalidateCache(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		p.error(w, r, http.StatusMethodNotAllowed, "cached contentserver export: invalidate cache failed - method not allowed")
		return
	}

	if r.Header.Get("Authorization") != p.APIKey {
		p.error(w, r, http.StatusUnauthorized, "cached contentserver export: invalidate cache failed - authorization required")
		return
	}

	log.Println(fmt.Sprintf("%s\t%s", r.URL, "cache invalidation request"))

	workspace := p.getRequestedWorkspace(r.URL)
	channel := p.addInvalidationChannel(workspace)

	select {
	case channel <- time.Now():
		w.WriteHeader(http.StatusOK)
		log.Println(fmt.Sprintf("added cache invalidation request to queue for workspace %s", workspace))
	default:
		w.WriteHeader(http.StatusTooManyRequests)
		log.Println(fmt.Sprintf("ignored cache invalidation request due to pending invalidation requests for workspace %s", workspace))
	}
}

// serveCachedNeosContentServerExport ...
func (p *Proxy) serveCachedNeosContentServerExport(w http.ResponseWriter, r *http.Request) {
	workspace := p.getRequestedWorkspace(r.URL)
	cacheFilename := p.getCacheFilename(workspace)

	log.Println(fmt.Sprintf("%s\t%s", r.URL, "serve cached neos content server export request"))

	if _, err := os.Stat(cacheFilename); os.IsNotExist(err) {
		log.Println(fmt.Sprintf("cached contentserver export: not yet cached for workspace %s", workspace))
		if err = p.cacheNeosContentServerExport(workspace); err != nil {
			log.Println(err.Error())
			p.error(w, r, http.StatusInternalServerError, "cached contentserver export: unable to load export from neos")
			return
		}
	}
	p.streamCachedNeosContentServerExport(w, r)
}

// getRequestedWorkspace returns the requested workspace or default
func (p *Proxy) getRequestedWorkspace(url *url.URL) string {
	workspace := url.Query().Get("workspace")
	if workspace == "" {
		workspace = DefaultWorkspace
	}
	return workspace
}

// getCacheFilename returns the cache filename for the given workspace
func (p *Proxy) getCacheFilename(workspace string) string {
	return fmt.Sprintf("%s/contentserver-export-%s.json", p.Config.Cache.Directory, workspace)
}

// streamCachedNeosContentServerExport ...
func (p *Proxy) streamCachedNeosContentServerExport(w http.ResponseWriter, r *http.Request) {
	log.Println("cached contentserver export: stream file start")

	workspace := p.getRequestedWorkspace(r.URL)
	cacheFilename := p.getCacheFilename(workspace)

	if _, err := os.Stat(cacheFilename); os.IsNotExist(err) {
		p.error(w, r, http.StatusNotFound, "cached contentserver export: file not found")
		return
	}

	fileInfo, err := os.Stat(cacheFilename)
	if err != nil {
		p.error(w, r, http.StatusInternalServerError, "cached contentserver export: read file info failed")
		return
	}

	bytes, err := ioutil.ReadFile(cacheFilename)
	if err != nil {
		p.error(w, r, http.StatusInternalServerError, "cached contentserver export: read file failed")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Last-Modified", fileInfo.ModTime().Format(http.TimeFormat))
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
	log.Println("cached contentserver export: stream file done")
	return
}

// cacheNeosContentServerExport ...
func (p *Proxy) cacheNeosContentServerExport(workspace string) error {
	log.Println(fmt.Sprintf("caching new neos contentserver export for workspace %s", workspace))

	cacheFilename := p.getCacheFilename(workspace)
	downloadFilename := cacheFilename + ".download"

	if err := p.downloadNeosContentServerExport(downloadFilename, workspace); err != nil {
		return err
	}

	cacheFileHash, err := p.getMD5Hash(cacheFilename)
	if err != nil {
		return err
	}

	downloadFileHash, err := p.getMD5Hash(downloadFilename)
	if err != nil {
		return err
	}

	log.Println("md5 new: '" + downloadFileHash + "', md5 old: '" + cacheFileHash + "'")

	if err := os.Rename(downloadFilename, cacheFilename); err != nil {
		return err
	}

	// notify webhooks
	if cacheFileHash != downloadFileHash {
		p.NotifyOnUpdate(workspace)
	} else {
		log.Println("skipping 'updated' notifications since nothing changed")
	}

	log.Println(fmt.Sprintf("cached new contentserver export from neos for workspace %s", workspace))

	return nil
}

// downloadNeosContentServerExport ...
func (p *Proxy) downloadNeosContentServerExport(filename string, workspace string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	response, err := http.Get(p.Config.Neos.URL.String() + "/contentserver/export?workspace=" + workspace)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintln("unexpected status code from site contentserver export", response.StatusCode, response.Status))
	}
	defer response.Body.Close()

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if _, err = file.Write(bodyBytes); err != nil {
		return err
	}

	return nil
}

// getM5Hash returns a file's md5 hash
func (p *Proxy) getMD5Hash(filename string) (hash string, err error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return "", nil
	}
	file, err := os.Open(filename)
	if err != nil {
		return hash, err
	}
	defer file.Close()
	hasher := md5.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return hash, err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
