package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/skeswa/gophr/common/depot"
	"github.com/skeswa/gophr/common/errors"
)

const (
	blobHandlerURLVarAuthor = "author"
	blobHandlerURLVarRepo   = "repo"
	blobHandlerURLVarSHA    = "sha"
	blobHandlerURLVarPath   = "path"
)

type blobRequestArgs struct {
	author string
	repo   string
	sha    string
	path   string
}

// BlobHandler creates an HTTP request handler that responds to filepath lookups.
func BlobHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		args, err := extractBlobRequestArgs(r)
		if err != nil {
			errors.RespondWithError(w, err)
			return
		}

		// Request the filepath from depot gitweb.
		hashedRepoName := depot.BuildHashedRepoName(args.author, args.repo, args.sha)
		depotBlobURL := fmt.Sprintf("http://%s/?p=%s;a=blob_plain;f=%s;hb=refs/heads/master", depot.DepotInternalServiceAddress, hashedRepoName, args.path)
		depotBlobResp, err := http.Get(depotBlobURL)
		if err != nil {
			errors.RespondWithError(w, err)
			return
		}

		// If path was not found return 404.
		if depotBlobResp.StatusCode == 404 {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte{})
		}

		body, err := ioutil.ReadAll(depotBlobResp.Body)
		if err != nil {
			errors.RespondWithError(w, err)
			return
		}
		depotBlobResp.Body.Close()

		if len(body) > 0 {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(body))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte{})
		}
	}
}

func extractBlobRequestArgs(r *http.Request) (blobRequestArgs, error) {
	vars := mux.Vars(r)
	args := blobRequestArgs{}

	args.author = vars[blobHandlerURLVarAuthor]
	if len(args.author) < 0 {
		return args, NewInvalidURLParameterError(blobHandlerURLVarAuthor, args.author)
	}

	args.repo = vars[blobHandlerURLVarRepo]
	if len(args.repo) < 0 {
		return args, NewInvalidURLParameterError(blobHandlerURLVarRepo, args.repo)
	}

	args.sha = vars[blobHandlerURLVarSHA]
	if len(args.sha) < 0 {
		return args, NewInvalidURLParameterError(blobHandlerURLVarSHA, args.sha)
	}

	args.path = vars[blobHandlerURLVarPath]
	if len(args.path) < 0 {
		return args, NewInvalidURLParameterError(blobHandlerURLVarPath, args.path)
	}

	return args, nil
}
