package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/skeswa/gophr/common"
)

const (
	packageRequestRegexIndexAuthor                         = 1
	packageRequestRegexIndexRepo                           = 2
	barePackageRequestRegexIndexSubpath                    = 3
	packageRefRequestRegexIndexRef                         = 3
	packageRefRequestRegexIndexSubpath                     = 4
	packageVersionRequestRegexIndexSemverPrefix            = 3
	packageVersionRequestRegexIndexSemverMajorVersion      = 4
	packageVersionRequestRegexIndexSemverMinorVersion      = 5
	packageVersionRequestRegexIndexSemverPatchVersion      = 6
	packageVersionRequestRegexIndexSemverPrereleaseLabel   = 7
	packageVersionRequestRegexIndexSemverPrereleaseVersion = 8
	packageVersionRequestRegexIndexSemverSuffix            = 9
	packageVersionRequestRegexIndexSubpath                 = 10
)

const (
	formKeyGoGet                    = "go-get"
	formValueGoGet                  = "1"
	contentTypeHTML                 = "text/html"
	subPathRegexStr                 = `((?:\/[a-zA-Z0-9][-.a-zA-Z0-9]*)*)`
	userRepoRegexStr                = `^\/([a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\/([a-zA-Z0-9\.\-_]+)`
	masterGitRefLabel               = "master"
	someFakeGitTagRef               = "refs/tags/thisisnotathinginanyrepo"
	gitRefsInfoSubPath              = "/info/refs"
	httpLocationHeader              = "Location"
	refSelectorRegexStr             = "([a-fA-F0-9]{40})"
	gitUploadPackSubPath            = "/git-upload-pack"
	httpContentTypeHeader           = "Content-Type"
	packagePageURLTemplate          = "https://%s/#/packages/%s/%s"
	contentTypeGitUploadPack        = "application/x-git-upload-pack-advertisement"
	githubUploadPackURLTemplate     = "https://github.com/%s/%s/git-upload-pack"
	packageRequestRegexTemplate     = `%s(?:@%s)%s$`
	versionSelectorRegexTemplate    = `([\%c\%c]?)([0-9]+)(?:\.([0-9]+|%c))?(?:\.([0-9]+|%c))?(?:\-([a-zA-Z0-9\-_]+[a-zA-Z0-9])(?:\.([0-9]+|%c))?)?([\%c\%c]?)`
	barePackageRequestRegexTemplate = `%s%s$`
)

var (
	goGetMetadataTemplate   = `<html><head><meta name="go-import" content="%s git %s://%s"><meta name="go-source" content="%s _ https://%s/tree/%s{/dir} https://%s/blob/%s{/dir}/{file}#L{line}"></head><body>go get %s</body></html>`
	versionSelectorRegexStr = fmt.Sprintf(
		versionSelectorRegexTemplate,
		common.SemverSelectorTildeChar,
		common.SemverSelectorCaratChar,
		common.SemverSelectorWildcardChar,
		common.SemverSelectorWildcardChar,
		common.SemverSelectorWildcardChar,
		common.SemverSelectorLessThanChar,
		common.SemverSelectorGreaterThanChar,
	)
	packageRefRequestRegexStr = fmt.Sprintf(
		packageRequestRegexTemplate,
		userRepoRegexStr,
		refSelectorRegexStr,
		subPathRegexStr,
	)
	barePackageRequestRegexStr = fmt.Sprintf(
		barePackageRequestRegexTemplate,
		userRepoRegexStr,
		subPathRegexStr,
	)
	packageVersionRequestRegexStr = fmt.Sprintf(
		packageRequestRegexTemplate,
		userRepoRegexStr,
		versionSelectorRegexStr,
		subPathRegexStr,
	)

	packageRefRequestRegex     = regexp.MustCompile(packageRefRequestRegexStr)
	barePackageRequestRegex    = regexp.MustCompile(barePackageRequestRegexStr)
	packageVersionRequestRegex = regexp.MustCompile(packageVersionRequestRegexStr)
)

// PackageRequest is stuct that standardizes the output of all the scenarios
// through which a package may be requested. PackageRequest is essentially a
// helper struct to move data between the sub-functions of
// RespondToPackageRequest and RespondToPackageRequest itself.
type PackageRequest struct {
	Repo       string
	Author     string
	Subpath    string
	RefsData   []byte
	Selector   string
	GithubTree string
}

// RespondToPackageRequest processes an incoming request, evaluates whether is a
// correctly formatted request for package-related data, and either responds
// appropriately or returns an error indicating what went wrong.
func RespondToPackageRequest(
	context common.RequestContext,
	req *http.Request,
	res http.ResponseWriter,
) error {
	var (
		err            error
		packageRequest PackageRequest
	)

	// Attempt every request parsing strategy in order or popularity
	packageRequest, err = processPackageVersionRequest(context, req)
	if err != nil {
		refReqErr := err
		packageRequest, err = processPackageRefRequest(context, req)
		if err != nil {
			verReqErr := err
			packageRequest, err = processBarePackageRequest(context, req)
			if err != nil {
				return NewInvalidPackageRequestError(
					req.URL.Path,
					refReqErr,
					verReqErr,
					err,
				)
			}
		}
	}

	switch packageRequest.Subpath {
	case gitUploadPackSubPath:
		log.Printf(
			"[%s] Responding with the Github upload pack permanent redirect\n",
			context.RequestID,
		)

		res.Header().Set(
			httpLocationHeader,
			fmt.Sprintf(
				githubUploadPackURLTemplate,
				packageRequest.Author,
				packageRequest.Repo,
			),
		)
		res.WriteHeader(http.StatusMovedPermanently)
	case gitRefsInfoSubPath:
		log.Printf(
			"[%s] Responding with the git refs pulled from Github\n",
			context.RequestID,
		)

		res.Header().Set(httpContentTypeHeader, contentTypeGitUploadPack)
		res.Write(packageRequest.RefsData)
	default:
		if req.FormValue(formKeyGoGet) == formValueGoGet {
			log.Printf(
				"[%s] Responding with html formatted for go get\n",
				context.RequestID,
			)

			res.Header().Set(httpContentTypeHeader, contentTypeHTML)
			res.Write([]byte(generateGoGetMetadata(
				packageRequest.Author,
				packageRequest.Repo,
				packageRequest.Selector,
				packageRequest.Subpath,
				packageRequest.GithubTree,
			)))
		} else {
			log.Printf(
				"[%s] Responding with a permanent redirect to the gophr package webpage\n",
				context.RequestID,
			)

			res.Header().Set(
				httpLocationHeader,
				fmt.Sprintf(
					packagePageURLTemplate,
					req.URL.Host,
					packageRequest.Author,
					packageRequest.Repo,
				),
			)

			res.WriteHeader(http.StatusMovedPermanently)
		}
	}

	return nil
}

// processPackageRefRequest is a sub-function of RespondToPackageRequest that
// parses and simplifies the information in a package ref request into an
// instance of PackageRequest.
func processPackageRefRequest(
	context common.RequestContext,
	req *http.Request,
) (PackageRequest, error) {
	var (
		matches    []string
		requestURL string
	)

	requestURL = req.URL.Path
	matches = packageRefRequestRegex.FindStringSubmatch(requestURL)
	if matches == nil {
		return PackageRequest{},
			NewInvalidPackageRefRequestURLError(requestURL)
	}

	var (
		packageRef     = matches[packageRefRequestRegexIndexRef]
		packageRepo    = matches[packageRequestRegexIndexRepo]
		packageAuthor  = matches[packageRequestRegexIndexAuthor]
		packageSubpath = matches[packageRefRequestRegexIndexSubpath]

		packageRefsData []byte
	)

	// Only go out to fetch refs if they're going to get used
	if packageSubpath == gitRefsInfoSubPath {
		refs, err := common.FetchRefs(packageAuthor, packageRepo)
		if err != nil {
			return PackageRequest{}, err
		}
		// Reserialize the refs data with everything pointing at the specified ref.
		// The ref hash is obviously packageRef, but the name is empty needs to be a
		// made up tag.
		packageRefsData = refs.Reserialize(someFakeGitTagRef, packageRef)
	}

	return PackageRequest{
		Repo:       packageRepo,
		Author:     packageAuthor,
		Subpath:    packageSubpath,
		RefsData:   packageRefsData,
		Selector:   packageRef,
		GithubTree: packageRef,
	}, nil
}

// processBarePackageRequest is a sub-function of RespondToPackageRequest that
// parses and simplifies the information in a base package request into an
// instance of PackageRequest.
func processBarePackageRequest(
	context common.RequestContext,
	req *http.Request,
) (PackageRequest, error) {
	var (
		matches    []string
		requestURL string
	)

	requestURL = req.URL.Path
	matches = barePackageRequestRegex.FindStringSubmatch(requestURL)
	if matches == nil {
		return PackageRequest{},
			NewInvalidBarePackageRequestURLError(requestURL)
	}

	var (
		packageRepo    = matches[packageRequestRegexIndexRepo]
		packageAuthor  = matches[packageRequestRegexIndexAuthor]
		packageSubpath = matches[barePackageRequestRegexIndexSubpath]

		packageRefsData []byte
	)

	// Only go out to fetch refs if they're going to get used
	if packageSubpath == gitRefsInfoSubPath {
		refs, err := common.FetchRefs(packageAuthor, packageRepo)
		if err != nil {
			return PackageRequest{}, err
		}
		// Just pass the refs along
		// TODO(skeswa): come up with a way to skip candidate matching here
		packageRefsData = refs.Data
	}

	return PackageRequest{
		Repo:       packageRepo,
		Author:     packageAuthor,
		Subpath:    packageSubpath,
		RefsData:   packageRefsData,
		Selector:   "",
		GithubTree: masterGitRefLabel,
	}, nil
}

// processPackageVersionRequest is a sub-function of RespondToPackageRequest
// that parses and simplifies the information in a package version request into
// an instance of PackageRequest.
func processPackageVersionRequest(
	context common.RequestContext,
	req *http.Request,
) (PackageRequest, error) {
	var (
		matches    []string
		requestURL string
	)

	requestURL = req.URL.Path
	matches = packageVersionRequestRegex.FindStringSubmatch(requestURL)
	if matches == nil {
		return PackageRequest{},
			NewInvalidPackageVersionRequestURLError(requestURL)
	}

	var (
		packageRepo          = matches[packageRequestRegexIndexRepo]
		packageAuthor        = matches[packageRequestRegexIndexAuthor]
		packageSubpath       = matches[packageVersionRequestRegexIndexSubpath]
		hasMatchedCandidate  = false
		semverSelectorExists = false

		semverSelector        common.SemverSelector
		packageRefsData       []byte
		matchedCandidate      common.SemverCandidate
		matchedCandidateLabel string
	)

	selector, err := common.NewSemverSelector(
		matches[packageVersionRequestRegexIndexSemverPrefix],
		matches[packageVersionRequestRegexIndexSemverMajorVersion],
		matches[packageVersionRequestRegexIndexSemverMinorVersion],
		matches[packageVersionRequestRegexIndexSemverPatchVersion],
		matches[packageVersionRequestRegexIndexSemverPrereleaseLabel],
		matches[packageVersionRequestRegexIndexSemverPrereleaseVersion],
		matches[packageVersionRequestRegexIndexSemverSuffix],
	)

	if err != nil {
		return PackageRequest{},
			NewInvalidPackageVersionRequestURLError(requestURL, err)
	}

	semverSelector = selector
	semverSelectorExists = true

	log.Printf(
		"[%s] Found a version selector in the request URL\n",
		context.RequestID,
	)

	// Only go out to fetch refs if they're going to get used
	if req.FormValue(formKeyGoGet) == formValueGoGet ||
		packageSubpath == gitRefsInfoSubPath {
		log.Printf(
			"[%s] Fetching Github refs since this request is either from a go get or has an info path\n",
			context.RequestID,
		)

		refs, err := common.FetchRefs(packageAuthor, packageRepo)

		if err != nil {
			return PackageRequest{}, err
		}

		if semverSelectorExists &&
			refs.Candidates != nil &&
			len(refs.Candidates) > 0 {
			// Get the list of candidates that match the selector
			matchedCandidates := refs.Candidates.Match(semverSelector)
			log.Printf(
				"[%s] Matched candidates to the version selector\n",
				context.RequestID,
			)
			// Only proceed if there is at least one matched candidate
			if matchedCandidates != nil && len(matchedCandidates) > 0 {
				if len(matchedCandidates) == 1 {
					matchedCandidate = matchedCandidates[0]
					hasMatchedCandidate = true
				} else {
					selectorHasLessThan :=
						semverSelector.Suffix == common.SemverSelectorSuffixLessThan
					selectorHasWildcards :=
						semverSelector.MinorVersion.Type == common.SemverSegmentTypeWildcard ||
							semverSelector.PatchVersion.Type == common.SemverSegmentTypeWildcard ||
							semverSelector.PrereleaseVersion.Type == common.SemverSegmentTypeWildcard

					var matchedCandidateReference *common.SemverCandidate
					if selectorHasWildcards || selectorHasLessThan {
						matchedCandidateReference = matchedCandidates.Highest()
					} else {
						matchedCandidateReference = matchedCandidates.Lowest()
					}

					matchedCandidate = *matchedCandidateReference
					hasMatchedCandidate = true
				}

				log.Printf(
					"[%s] There was at least one candidate matched to the version selector\n",
					context.RequestID,
				)
			}
		}

		if hasMatchedCandidate {
			log.Printf(
				"[%s] Tweaked the refs to focus on the matched candidate\n",
				context.RequestID,
			)
			packageRefsData = refs.Reserialize(
				matchedCandidate.GitRefName,
				matchedCandidate.GitRefHash,
			)
			matchedCandidateLabel = matchedCandidate.GitRefLabel
		} else {
			if !semverSelectorExists {
				// Since there was no selector, we are fine with the fact that we didn't
				// find a match. Now, return the original refs that we downloaded from
				// github that point to master by default.
				packageRefsData = refs.Data
			} else {
				log.Printf(
					"[%s] Couldn't find any candidates to match to the version selector \"%s\"\n",
					context.RequestID,
					semverSelector.String(),
				)

				return PackageRequest{}, NewNoSuchPackageVersionError(
					packageAuthor,
					packageRepo,
					semverSelector.String(),
				)
			}
		}
	}

	// If there is no label as of yet, just default to master
	if len(matchedCandidateLabel) < 1 {
		matchedCandidateLabel = masterGitRefLabel
	}

	return PackageRequest{
		Repo:       packageRepo,
		Author:     packageAuthor,
		Subpath:    packageSubpath,
		RefsData:   packageRefsData,
		Selector:   semverSelector.String(),
		GithubTree: matchedCandidateLabel,
	}, nil
}

// generateGoGetMetadata generates the format of metadata that the go get tool
// expects to receive from unknown repository domains before its starts pulling
// down source code.
func generateGoGetMetadata(
	user string,
	repo string,
	selector string,
	subpath string,
	githubTree string,
) string {
	var (
		buffer bytes.Buffer

		config     = getConfig()
		protocol   string
		gophrRoot  string
		gophrPath  string
		githubRoot string
	)

	if config.dev {
		protocol = "http"
	} else {
		protocol = "https"
	}

	buffer.WriteString(config.domain)
	buffer.WriteByte('/')
	buffer.WriteString(user)
	buffer.WriteByte('/')
	buffer.WriteString(repo)
	if len(selector) > 0 {
		buffer.WriteByte('@')
		buffer.WriteString(selector)
	}
	gophrRoot = buffer.String()

	buffer.WriteString(subpath)
	gophrPath = buffer.String()

	buffer.Reset()
	buffer.WriteString("github.com")
	buffer.WriteByte('/')
	buffer.WriteString(user)
	buffer.WriteByte('/')
	buffer.WriteString(repo)
	githubRoot = buffer.String()

	if len(githubTree) < 1 {
		githubTree = masterGitRefLabel
	}

	return fmt.Sprintf(
		goGetMetadataTemplate,
		gophrRoot,
		protocol,
		gophrRoot,
		gophrRoot,
		githubRoot,
		githubTree,
		githubRoot,
		githubTree,
		gophrPath,
	)
}