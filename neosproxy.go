package main

import (
	"bytes"
	"crypto/tls"
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

func (p *Proxy) error(w http.ResponseWriter, r *http.Request, code int, msg string) {
	log.Println(fmt.Sprintf("%d\t%s\t%s", code, r.URL, msg))
	w.WriteHeader(code)
}

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

	_, err = io.Copy(cacheFile, response.Body)
	if err != nil {
		return
	}

	err = os.Rename(cacheFile.Name(), p.FilenameCachedContentServerExport)
	if err != nil {
		return
	}

	log.Println(fmt.Sprintf("%d\t%s\t%s", http.StatusOK, "/contentserverproxy/cache", "got new contentserver export from neos"))

	// Notify webhooks
	if len(p.UpdatedHooks) > 0 {
		p.notify("updated", p.UpdatedHooks)
	}

	return nil
}

func (p *Proxy) notify(event string, urls []string) {
	log.Println(fmt.Sprintf("Notifying %d for '%s' event", len(urls), event))
	for _, url := range p.UpdatedHooks {
		data, _ := json.Marshal(map[string]string{
			"type": event,
		})
		go func() {
			var client *http.Client
			if p.TLSHooks {
				client = &http.Client{}
			} else {
				client = &http.Client{Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: p.TLSHooksSkipVerify},
				},
				}
			}
			resp, err := client.Post(url, "application/json", bytes.NewBuffer(data))
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
	UpdatedHooks                      []string
	TLSHooks                          bool
	TLSHooksSkipVerify                bool
}

func (p Proxy) run() {
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

	log.Fatal(http.ListenAndServe(p.Address, mux))
}

func main() {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal(fmt.Errorf("missing env variable API_KEY"))
		return
	}

	flagAddress := flag.String("address", "0.0.0.0:80", "address to listen to")
	flagNeosHostname := flag.String("neos", "http://neos/", "neos cms hostname")
	flagTLSHooks := flag.Bool("tls-hooks", false, "enable TLS on webhook notifications")
	flagTLSHooksSkipVerify := flag.Bool("tls-hooks-skip-verify", false, "skip TLS verification on webhook notifications")
	flagUpdatedHooks := flag.String("updated-hooks", "", "comma seperated list of urls to notify on update event")
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
		UpdatedHooks:                      strings.Split(*flagUpdatedHooks, ","),
		TLSHooks:                          *flagTLSHooks,
		TLSHooksSkipVerify:                *flagTLSHooksSkipVerify,
	}

	fmt.Println("start proxy on ", *flagAddress, "for neos", *flagNeosHostname, "with cache file in", p.FilenameCachedContentServerExport)
	p.run()
}
