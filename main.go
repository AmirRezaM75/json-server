package main

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
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
		e.handlerFunc(w, r.WithContext(ctx))
		return
	}

	http.NotFound(w, r)
}

func (router *Router) Route(method string, path string, handlerFunc http.HandlerFunc) {
	entry := RouteEntry{
		method:      method,
		path:        regexp.MustCompile("^" + path + "$"),
		handlerFunc: handlerFunc,
	}

	router.entries = append(router.entries, entry)
}

type RouteEntry struct {
	method      string
	path        *regexp.Regexp
	handlerFunc http.HandlerFunc
}

func (e RouteEntry) match(r *http.Request) map[string]string {
	matches := e.path.FindStringSubmatch(r.URL.Path)

	if matches == nil {
		return nil
	}

	params := make(map[string]string)

	groupNames := e.path.SubexpNames()

	for i, match := range matches {
		params[groupNames[i]] = match
	}
	/*
	   if e.method != r.Method {
	   		return nil
	   	}*/
	return params
}

func indexUserHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintln(w, "IndexUserHandler")
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	userId := URLParam(r, "UserId")

	_, _ = fmt.Fprintln(w, "GetUserHandler "+userId)
}

func main() {
	r := &Router{}
	r.Route(http.MethodGet, "/users", indexUserHandler)
	r.Route(http.MethodGet, `/users/(?P<UserId>\d+)`, getUserHandler)
	fmt.Println("Listening at port 3000...")
	_ = http.ListenAndServe(":3000", r)
}

// URLParam extracts a parameter from the URL by name
func URLParam(r *http.Request, name string) string {
	ctx := r.Context()

	// ctx.Value returns an `interface{}` type, so we
	// also have to cast it to a map, which is the
	// type we'll be using to store our parameters.
	params := ctx.Value("params").(map[string]string)
	return params[name]
}
