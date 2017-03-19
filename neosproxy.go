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

var md5Sum string

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

	select {
	case p.CacheInvalidationChannel <- time.Now():
		w.WriteHeader(http.StatusOK)
		log.Println(fmt.Sprintf("%d\t%s\t%s", http.StatusOK, r.URL, "added cache invalidation request to queue"))
	default:
		w.WriteHeader(http.StatusTooManyRequests)
		log.Println(fmt.Sprintf("%d\t%s\t%s", http.StatusTooManyRequests, r.URL, "ignored cache invalidation request due to pending invalidation requests"))
	}
}

// serveCachedNeosContentServerExport ...
func (p *Proxy) serveCachedNeosContentServerExport(w http.ResponseWriter, r *http.Request) {
	if _, err := os.Stat(p.FilenameCachedContentServerExport); os.IsNotExist(err) {
		log.Println("cached contentserver export: not yet cached")
		err = p.cacheNeosContentServerExport()
		if err != nil {
			p.error(w, r, http.StatusInternalServerError, "cached contentserver export: unable to load export from neos")
			return
		}
	}
	p.streamCachedNeosContentServerExport(w, r)
}

// streamCachedNeosContentServerExport ...
func (p *Proxy) streamCachedNeosContentServerExport(w http.ResponseWriter, r *http.Request) {
	log.Println("cached contentserver export: stream file start")
	if _, err := os.Stat(p.FilenameCachedContentServerExport); os.IsNotExist(err) {
		p.error(w, r, http.StatusNotFound, "cached contentserver export: file not found")
		return
	}

	bytes, err := ioutil.ReadFile(p.FilenameCachedContentServerExport)
	if err != nil {
		p.error(w, r, http.StatusInternalServerError, "cached contentserver export: read file failed")
		return
	}

	fileInfo, err := os.Stat(p.FilenameCachedContentServerExport)
	if err != nil {
		p.error(w, r, http.StatusInternalServerError, "cached contentserver export: read file info failed")
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
func (p *Proxy) cacheNeosContentServerExport() (err error) {
	log.Println(fmt.Sprintf("%d\t%s\t%s", http.StatusProcessing, "/contentserverproxy/cache", "get new contentserver export from neos"))
	cacheFile, err := os.Create(p.FilenameCachedContentServerExport + ".download")
	if err != nil {
		return
	}
	defer cacheFile.Close()

	response, err := http.Get(p.Endpoint.String() + "/contentserver/export")
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		err = errors.New(fmt.Sprintln("unexpected status code from site contentserver export", response.StatusCode, response.Status))
		return
	}
	defer response.Body.Close()

	if _, err = io.Copy(cacheFile, response.Body); err != nil {
		return
	}

	err = os.Rename(cacheFile.Name(), p.FilenameCachedContentServerExport)
	if err != nil {
		return
	}

	log.Println(fmt.Sprintf("%d\t%s\t%s", http.StatusOK, "/contentserverproxy/cache", "got new contentserver export from neos"))

	return

	hasher := md5.New()
	if _, err = io.Copy(hasher, response.Body); err != nil {
		return
	}
	newMD5Sum := hex.EncodeToString(hasher.Sum(nil))
	log.Println("md5 new: " + newMD5Sum + ", md5 old: " + md5Sum)

	// Notify webhooks
	if len(p.CallbackUpdated) > 0 && md5Sum != newMD5Sum {
		p.notify("updated", p.CallbackUpdated)
		md5Sum = newMD5Sum
	} else {
		log.Println("skipping 'updated' notifications since nothing changed")
	}

	return nil
}

// notify notifies callbacks for the given event
func (p *Proxy) notify(event string, urls []string) {
	log.Println(fmt.Sprintf("Notifying %d for '%s' event", len(urls), event))
	data, _ := json.Marshal(map[string]string{
		"type": event,
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
	APIKey                            string
	Address                           string
	Endpoint                          *url.URL
	FilenameCachedContentServerExport string
	CacheInvalidationChannel          chan time.Time
	CallbackUpdated                   []string
	CallbackKey                       string
	CallbackTLSVerify                 bool
}

func (p Proxy) run() error {
	go func(channel chan time.Time) {
		for {
			sleepTime := 5 * time.Second
			time.Sleep(sleepTime)
			requestTime := <-channel
			p.cacheNeosContentServerExport()
			duration := time.Since(requestTime.Add(sleepTime))
			log.Println(fmt.Sprintf("%d\t%s\t%s %s %s %s", http.StatusCreated, "/contentserverproxy/cache", "processed cache invalidation request, which has been added at", requestTime.Format(time.RFC3339), "in", duration))
		}
	}(p.CacheInvalidationChannel)

	proxyHandler := httputil.NewSingleHostReverseProxy(p.Endpoint)

	mux := http.NewServeMux()
	mux.Handle("/contentserver/export/", proxyHandler)
	mux.HandleFunc("/contentserverproxy/cache", p.invalidateCache)
	mux.HandleFunc("/contentserver/export", p.serveCachedNeosContentServerExport)
	mux.Handle("/", proxyHandler)

	return http.ListenAndServe(p.Address, mux)
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
		APIKey:                            apiKey,
		Address:                           *flagAddress,
		Endpoint:                          neosURL,
		CacheInvalidationChannel:          make(chan time.Time, 1),
		FilenameCachedContentServerExport: os.TempDir() + "neos-contentserverexport.json",
		CallbackKey:                       *flagCallbackKey,
		CallbackTLSVerify:                 *flagCallbackTLSVerify,
		CallbackUpdated:                   strings.Split(*flagCallbackUpdated, ","),
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
				select {
				case p.CacheInvalidationChannel <- time.Now():
					log.Println("auto update: added cache invalidation request to queue")
				default:
					log.Println("auto update: ignored cache invalidation request due to pending invalidation requests")
				}
			}
		}()
	}

	fmt.Println("start proxy on ", *flagAddress, "for neos", *flagNeosHostname, "with cache file in", p.FilenameCachedContentServerExport)
	if err := p.run(); err != nil {
		log.Fatal(err)
	}
}
