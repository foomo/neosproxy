package main

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

const DefaultWorkspace = "live"

// error ...
func (p *Proxy) error(w http.ResponseWriter, r *http.Request, code int, msg string) {
	log.Println(fmt.Sprintf("%d\t%s\t%s", code, r.URL, msg))
	w.WriteHeader(code)
}

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
	return fmt.Sprintf("%s/contentserver-export-%s.json", p.CacheDir, workspace)
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

	cacheFileHash, err := p.getM5Hash(cacheFilename)
	if err != nil {
		return err
	}

	downloadFileHash, err := p.getM5Hash(downloadFilename)
	if err != nil {
		return err
	}

	log.Println("md5 new: '" + downloadFileHash + "', md5 old: '" + cacheFileHash + "'")

	if err := os.Rename(downloadFilename, cacheFilename); err != nil {
		return err
	}

	// Notify webhooks
	if len(p.CallbackUpdated) > 0 && cacheFileHash != downloadFileHash {
		p.notify("updated", p.CallbackUpdated, workspace)
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

	response, err := http.Get(p.Endpoint.String() + "/contentserver/export?workspace=" + workspace)
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
func (p *Proxy) getM5Hash(filename string) (hash string, err error) {
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

// notify notifies callbacks for the given event
func (p *Proxy) notify(event string, urls []string, workspace string) {
	log.Println(fmt.Sprintf("Notifying %d for '%s' event on workspace %s", len(urls), event, workspace))
	data, _ := json.Marshal(map[string]string{
		"type":      event,
		"workspace": workspace,
	})
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: !p.CallbackTLSVerify,
			},
		},
	}

	for _, value := range urls {
		go func() {
			// Create request
			req, err := http.NewRequest(http.MethodPost, value, bytes.NewBuffer(data))
			if err != nil {
				log.Println(fmt.Sprintf("Failed to create callback request! Got error: %s", err.Error()))
				return
			}
			// Add header
			req.Header.Set("Content-Type", "application/json")
			req.Header.Add("key", p.CallbackKey)
			// Send request
			resp, err := httpClient.Do(req)
			if err != nil {
				log.Println(fmt.Sprintf("Failed to notify a webhook! Got error: %s", err.Error()))
			} else {
				log.Println(fmt.Sprintf("Notified webhook with response code: %d", resp.StatusCode))
			}
		}()
	}
}

type Proxy struct {
	APIKey                    string
	Address                   string
	Endpoint                  *url.URL
	CacheDir                  string
	CacheInvalidationChannels map[string](chan time.Time)
	CallbackUpdated           []string
	CallbackKey               string
	CallbackTLSVerify         bool
}

func (p *Proxy) run() error {
	p.addInvalidationChannel(DefaultWorkspace)
	proxyHandler := httputil.NewSingleHostReverseProxy(p.Endpoint)
	mux := http.NewServeMux()
	mux.Handle("/contentserver/export/", proxyHandler)
	mux.HandleFunc("/contentserverproxy/cache", p.invalidateCache)
	mux.HandleFunc("/contentserver/export", p.serveCachedNeosContentServerExport)
	mux.Handle("/", proxyHandler)

	return http.ListenAndServe(p.Address, mux)
}

// addInvalidationChannel adds a new invalidation channel
func (p *Proxy) addInvalidationChannel(workspace string) chan time.Time {
	if _, ok := p.CacheInvalidationChannels[workspace]; !ok {
		channel := make(chan time.Time, 1)
		p.CacheInvalidationChannels[workspace] = channel
		go func(workspace string, channel chan time.Time) {
			for {
				sleepTime := 5 * time.Second
				time.Sleep(sleepTime)
				requestTime := <-channel
				if err := p.cacheNeosContentServerExport(workspace); err != nil {
					log.Println(err.Error())
				} else {
					log.Println(fmt.Sprintf(
						"processed cache invalidation request, which has been added at %s in %.2fs for workspace %s",
						requestTime.Format(time.RFC3339),
						time.Since(requestTime.Add(sleepTime)).Seconds(),
						workspace,
					))
				}
			}
		}(workspace, channel)
	}
	return p.CacheInvalidationChannels[workspace]
}

func main() {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("missing env variable API_KEY")
	}

	flagAddress := flag.String("address", "0.0.0.0:80", "address to listen to")
	flagNeosHostname := flag.String("neos", "http://neos/", "neos cms hostname")

	// Callback flags
	flagCallbackKey := flag.String("callback-key", "", "optinonal header key to send with each web callback")
	flagCallbackTLSVerify := flag.Bool("callback-tls-verify", true, "skip TLS verification on web callbacks")
	flagCallbackUpdated := flag.String("callback-updated", "", "comma seperated list of urls to notify on update event")

	// Update flags
	flagAutoUpdate := flag.String("auto-update", "", "duration value on which to automatically update the proxy")

	flag.Parse()

	neosURL, err := url.Parse(*flagNeosHostname)
	if err != nil {
		log.Fatal(err)
	}

	p := &Proxy{
		APIKey:                    apiKey,
		Address:                   *flagAddress,
		Endpoint:                  neosURL,
		CacheDir:                  os.TempDir(),
		CacheInvalidationChannels: make(map[string](chan time.Time)),

		CallbackKey:       *flagCallbackKey,
		CallbackTLSVerify: *flagCallbackTLSVerify,
		CallbackUpdated:   strings.Split(*flagCallbackUpdated, ","),
	}

	if *flagAutoUpdate != "" {
		autoUpdate, err := time.ParseDuration(*flagAutoUpdate)
		if err != nil {
			log.Fatal("invalid auto-update duration value: " + err.Error())
		}
		go func() {
			log.Println("starting with auto update every " + *flagAutoUpdate)
			for {
				time.Sleep(autoUpdate)
				log.Println(fmt.Sprintf("auto update: updating %d cache", len(p.CacheInvalidationChannels)))
				for workspace, channel := range p.CacheInvalidationChannels {
					select {
					case channel <- time.Now():
						log.Println(fmt.Sprintf("auto update: added cache invalidation request to queue for '%s' workspace", workspace))
					default:
						log.Println(fmt.Sprintf("auto update: ignored cache invalidation request due to pending invalidation requests for '%s' workspace", workspace))
					}
				}
			}
		}()
	}

	fmt.Println("start proxy on ", *flagAddress, "for neos", *flagNeosHostname, "with cache dir in", p.CacheDir)
	if err := p.run(); err != nil {
		log.Fatal(err)
	}
}
