package cms

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {

	id := "a839f683-dc58-47aa-8000-72d5b6fdeb85"
	dimension := "de"
	workspace := "stage"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		path := fmt.Sprintf(pathContent, dimension, id, workspace)

		if r.RequestURI != path {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("error: invalid request uri: " + r.RequestURI))
			return
		}

		data := &Content{
			HTML: "<h1>Test</h1>",
		}

		encoder := json.NewEncoder(w)
		err := encoder.Encode(data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error: " + err.Error()))
			return
		}
	}))
	defer ts.Close()

	client, clientErr := New(ts.URL)
	assert.NoError(t, clientErr, "client must be initialised without errors")

	// context
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()

	content, contentErr := client.CMS.GetContent(id, dimension, workspace, ctx)
	assert.NoError(t, contentErr)
	assert.NotEmpty(t, content)
	assert.Equal(t, "<h1>Test</h1>", content.HTML)
}

func TestClientTimeout(t *testing.T) {

	id := "a839f683-dc58-47aa-8000-72d5b6fdeb85"
	dimension := "de"
	workspace := "stage"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		path := fmt.Sprintf(pathContent, dimension, id, workspace)

		if r.RequestURI != path {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("error: invalid request uri: " + r.RequestURI))
			return
		}

		time.Sleep(time.Second * 1)

		data := &Content{
			HTML: "<h1>Test</h1>",
		}

		encoder := json.NewEncoder(w)
		err := encoder.Encode(data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error: " + err.Error()))
			return
		}
	}))
	defer ts.Close()

	client, clientErr := New(ts.URL)
	assert.NoError(t, clientErr, "client must be initialised without errors")

	// context
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()

	content, contentErr := client.CMS.GetContent(id, dimension, workspace, ctx)
	assert.Error(t, contentErr)
	assert.Empty(t, content)
	assert.Equal(t, ErrorResponseTimeout, contentErr)
}
