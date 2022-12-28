package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

type Router struct {
	entries []RouteEntry
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, e := range router.entries {
		params := e.match(r)
		if params == nil {
			continue
		}

		ctx := context.WithValue(r.Context(), "params", params)
		controller(w, r.WithContext(ctx), e.response)
		return
	}

	http.NotFound(w, r)
}

func (router *Router) Route(method string, path string, response Response) {
	entry := RouteEntry{
		method:   method,
		path:     regexp.MustCompile("^" + path + "$"),
		response: response,
	}

	router.entries = append(router.entries, entry)
}

type Response struct {
	status int
	data   []byte
}

type RouteEntry struct {
	method   string
	path     *regexp.Regexp
	response Response
}

func (e RouteEntry) match(r *http.Request) map[string]string {
	if e.method != r.Method {
		return nil
	}

	matches := e.path.FindStringSubmatch(r.URL.Path)

	if matches == nil {
		return nil
	}

	params := make(map[string]string)

	groupNames := e.path.SubexpNames()

	for i, match := range matches {
		params[groupNames[i]] = match
	}

	return params
}

func controller(w http.ResponseWriter, r *http.Request, response Response) {
	w.Header().Set("Accept", "application/json")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.status)
	_, _ = w.Write(response.data)
}

type Options struct {
	Port      int        `json:"port"`
	Endpoints []Endpoint `json:"endpoints"`
}

type Endpoint struct {
	Method   string `json:"method"`
	Status   int    `json:"status"`
	Path     string `json:"path"`
	JsonPath string `json:"jsonPath"`
}

func main() {
	// bytes, err := ioutil.ReadFile("./api.json")
	raw, err := os.Open("api.json")

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var options Options

	rawBytes, err := ioutil.ReadAll(raw)

	err = json.Unmarshal(rawBytes, &options)

	r := &Router{}

	for _, endpoint := range options.Endpoints {
		file, err := os.Open(endpoint.JsonPath)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		//_ = file.Close()

		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(file)

		s := buf.String()

		b := []byte(s)

		response := Response{
			status: endpoint.Status,
			data:   b,
		}
		r.Route(endpoint.Method, endpoint.Path, response)
	}

	port := strconv.Itoa(options.Port)
	fmt.Println("Listening at port " + port)
	_ = http.ListenAndServe(":"+port, r)
}

func URLParam(r *http.Request, name string) string {
	ctx := r.Context()
	params := ctx.Value("params").(map[string]string)
	return params[name]
}
