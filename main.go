package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

type Response struct {
	status int
	data   []byte
}

type RouteEntry struct {
	method   string
	path     string
	response Response
}

type Router struct {
	entries []RouteEntry
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

	r := regexp.MustCompile(`:(\w+)`)
	path = r.ReplaceAllString(path, "(?P<$1>\\w+)")

	entry := RouteEntry{
		method:   method,
		path:     path,
		response: response,
	}

	router.entries = append(router.entries, entry)
}

func (e RouteEntry) match(r *http.Request) map[string]string {
	if e.method != r.Method {
		return nil
	}

	path := regexp.MustCompile("^" + e.path + "$")

	matches := path.FindStringSubmatch(r.URL.Path)

	if matches == nil {
		return nil
	}

	params := make(map[string]string)

	groupNames := path.SubexpNames()

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

func main() {
	raw := fileToBytes("api.json")

	var options Options

	err := json.Unmarshal(raw, &options)

	if err != nil {
		fmt.Println("Invalid JSON format.")
	}

	r := &Router{}

	for _, endpoint := range options.Endpoints {
		response := Response{
			status: endpoint.Status,
			data:   fileToBytes(endpoint.JsonPath),
		}

		r.Route(endpoint.Method, endpoint.Path, response)
	}

	port := strconv.Itoa(options.Port)
	fmt.Println("Listening at port " + port)
	_ = http.ListenAndServe(":"+port, r)
}

func fileToBytes(path string) []byte {
	file, err := os.Open(path)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}(file)

	buf := new(bytes.Buffer)

	_, _ = buf.ReadFrom(file)

	s := buf.String()

	return []byte(s)
}
