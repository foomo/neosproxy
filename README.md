NEOSPROXY
======

Neosproxy could be used to proxy a NEOS CMS.
It will cache requests to the Foomo.Neos.Contentserver package.

application
----

The application requires an API_KEY as an environment variable.
This API_KEY is used for authorization on cache invalidation requests.

Furthermore you need to define on which address the proxy should listen.
The NEOS endpoint must be configured by setting the "neos" parameter as well.

```
neosproxy --address=":80" --neos="https://www.myneosendpoint.com"
```

web hooks
----

Neosproxy can call given comma seperated urls to notify you on events:


* `--callback-updated`: called after update

If you want to skip tls verification you can pass the following option:

* `--callback-tls-verify=false`: skip tls verification

You can optionally add a `--callback-key` to be sent as header with each call.  

Example:

```
neosproxy --address=":80" --neos="https://www.myneosendpoint.com" --callback-tls-verify="false" --callback-updated="http://foo.com,https://bar.com" --callback-key="secret"
```

docker
----

For docker pass in the API_KEY as an environment variable as well.
Run the container in the same network as your neos container. 

```bash
docker run --rm -it -p="8080:80" -e="API_KEY=ZPNFYsXouqeRYPZ34cV4962KaZdU2Lp29LwbftMDeFBae3wcWX" foomo/neosproxy:latest -neos https://www.myneosendpoint.com 
```

routes
----

```
/contentserver/export

	GET: get the whole navigation tree as a contentserver json dump


/contentserverproxy/cache

	DELETE     : invalidate the cache

```

curl
----

Some curl examples for local development.

start proxy

```bash
export API_KEY=ZPNFYsXouqeRYPZ34cV4962KaZdU2Lp29LwbftMDeFBae3wcWX && \
go run neosproxy.go -address "127.0.0.1:8080" -neos "http://neos.localhost"
```

invalidate cache

```bash
curl -X "DELETE" -H "Authorization: ZPNFYsXouqeRYPZ34cV4962KaZdU2Lp29LwbftMDeFBae3wcWX" 127.0.0.1:8080/contentserverproxy/cache
```

get the contentserver export

```bash
curl -k 127.0.0.1:8080/contentserver/export
```
