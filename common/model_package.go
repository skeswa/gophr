package common

import (
	"fmt"
	"regexp"
	"time"

	"github.com/gocql/gocql"
)

// Constants directly related to interacting with the packages model in the
// cassandra database.
const (
	// TableNamePackages is the name of the table containing the packages model
	TableNamePackages = "packages"
	// IndexNamePackages is the name of the lucene index
	IndexNamePackages             = "packages_index"
	ColumnNamePackagesRepo        = "repo"
	ColumnNamePackagesExists      = "exists"
	ColumnNamePackagesAuthor      = "author"
	ColumnNamePackagesVersions    = "versions"
	ColumnNamePackagesGodocURL    = "godoc_url"
	ColumnNamePackagesIndexTime   = "index_time"
	ColumnNamePackagesAwesomeGo   = "awesome_go"
	ColumnNamePackagesSearchBlob  = "search_blob"
	ColumnNamePackagesDescription = "description"
)

const (
	packagesSearchBlobTemplate = "%s %s %s"
)

var (
	cqlQueryFuzzySearchPackagesTemplate = fmt.Sprintf(
		`SELECT %s,%s,%s FROM %s WHERE expr(%s,'{query:{type:"fuzzy",field:"%s",value:"%s"}}') LIMIT 10`,
		ColumnNamePackagesRepo,
		ColumnNamePackagesAuthor,
		ColumnNamePackagesDescription,
		TableNamePackages,
		IndexNamePackages,
		ColumnNamePackagesSearchBlob,
		"%s",
	)

	cqlQuerySelectPackageVersions = fmt.Sprintf(
		`SELECT %s FROM %s WHERE %s = ? AND %s = ? LIMIT 1`,
		ColumnNamePackagesVersions,
		TableNamePackages,
		ColumnNamePackagesAuthor,
		ColumnNamePackagesRepo,
	)

	cqlQueryInsertSearchPackage = fmt.Sprintf(
		`INSERT INTO %s (%s,%s,%s,%s,%s,%s,%s,%s,%s) VALUES (?,?,?,?,?,?,?,?,?)`,
		TableNamePackages,
		ColumnNamePackagesRepo,
		ColumnNamePackagesExists,
		ColumnNamePackagesAuthor,
		ColumnNamePackagesVersions,
		ColumnNamePackagesGodocURL,
		ColumnNamePackagesIndexTime,
		ColumnNamePackagesAwesomeGo,
		ColumnNamePackagesSearchBlob,
		ColumnNamePackagesDescription,
	)
)

var (
	alphanumericFilterRegex = regexp.MustCompile(`[^\sa-zA-Z0-9\-_]+`)
)

// PackageModel is a struct representing one individual package in the database.
type PackageModel struct {
	Repo        *string
	Exists      *bool
	Author      *string
	Versions    []string
	GodocURL    *string
	IndexTime   *time.Time
	AwesomeGo   *bool
	SearchBlob  *string
	Description *string
}

// NewPackageModelForInsert creates an instance of PackageModel that is
// optimized and validated for the insert operation in the database.
func NewPackageModelForInsert(
	repo string,
	exists bool,
	author string,
	versions []string,
	godocURL string,
	indexTime time.Time,
	awesomeGo bool,
	description string,
) (*PackageModel, error) {
	if len(repo) < 1 {
		return nil, NewInvalidParameterError("repo", repo)
	}
	if len(author) < 1 {
		return nil, NewInvalidParameterError("author", author)
	}
	if len(godocURL) < 1 {
		return nil, NewInvalidParameterError("godocURL", godocURL)
	}

	searchBlob := fmt.Sprintf(
		packagesSearchBlobTemplate,
		author,
		repo,
		description,
	)

	return &PackageModel{
		Repo:        &repo,
		Exists:      &exists,
		Author:      &author,
		Versions:    versions,
		GodocURL:    &godocURL,
		IndexTime:   &indexTime,
		AwesomeGo:   &awesomeGo,
		SearchBlob:  &searchBlob,
		Description: &description,
	}, nil
}

// NewPackageModelFromBulkSelect creates an instance of PackageModel that is
// optimized and validated for a select operation designed to get data about
// multiple packages from the database.
func NewPackageModelFromBulkSelect(
	repo string,
	author string,
	description string,
) (*PackageModel, error) {
	if len(repo) < 1 {
		return nil, NewInvalidParameterError("repo", repo)
	}
	if len(author) < 1 {
		return nil, NewInvalidParameterError("author", author)
	}

	return &PackageModel{
		Repo:        &repo,
		Author:      &author,
		Description: &description,
	}, nil
}

// NewPackageModelFromSingleSelect creates an instance of PackageModel that is
// optimized and validated for a select operation designed to get data about
// a single package from the database.
func NewPackageModelFromSingleSelect(
	repo string,
	exists bool,
	author string,
	versions []string,
	godocURL string,
	awesomeGo bool,
	description string,
) (*PackageModel, error) {
	if len(repo) < 1 {
		return nil, NewInvalidParameterError("repo", repo)
	}
	if len(author) < 1 {
		return nil, NewInvalidParameterError("author", author)
	}
	if len(godocURL) < 1 {
		return nil, NewInvalidParameterError("godocURL", godocURL)
	}

	return &PackageModel{
		Repo:        &repo,
		Exists:      &exists,
		Author:      &author,
		Versions:    versions,
		GodocURL:    &godocURL,
		AwesomeGo:   &awesomeGo,
		Description: &description,
	}, nil
}

// FindPackageVersions gets the versions of a package from the database. If
// no such package exists, or there were no versions for said package, then nil
// is returned.
func FindPackageVersions(session *gocql.Session, author string, repo string) ([]string, error) {
	var (
		err      error
		versions []string
	)

	iter := session.Query(cqlQuerySelectPackageVersions, author, repo).Iter()

	if !iter.Scan(&versions) {
		return nil, nil
	}

	if err = iter.Close(); err != nil {
		return nil, NewQueryScanError(nil, err)
	}

	return versions, nil
}

// FuzzySearchPackages finds a list of packages relevant to a query phrase
// string. The search takes author, package and description into account.
func FuzzySearchPackages(
	session *gocql.Session,
	searchText string,
) ([]*PackageModel, error) {
	// First, remove all non-essential characters
	searchText = alphanumericFilterRegex.ReplaceAllString(searchText, "")
	// Next put the search text into a query string
	query := fmt.Sprintf(cqlQueryFuzzySearchPackagesTemplate, searchText)
	// Return the processed results of the query
	return scanPackageModels(session.Query(query))
}

// InsertPackage inserts an individual package into the database.
func InsertPackage(
	session *gocql.Session,
	packageModel *PackageModel,
) error {
	err := session.Query(cqlQueryInsertSearchPackage,
		*packageModel.Repo,
		*packageModel.Exists,
		*packageModel.Author,
		packageModel.Versions,
		*packageModel.GodocURL,
		*packageModel.IndexTime,
		*packageModel.AwesomeGo,
		*packageModel.SearchBlob,
		*packageModel.Description,
	).Exec()

	return err
}

// InsertPackages inserts a slice of package models into the database.
func InsertPackages(
	session *gocql.Session,
	packageModels []*PackageModel,
) error {
	batch := gocql.NewBatch(gocql.LoggedBatch)

	if packageModels == nil || len(packageModels) == 0 {
		return NewInvalidParameterError("packageModels", packageModels)
	}

	for _, packageModel := range packageModels {
		if packageModel != nil &&
			packageModel.Repo != nil &&
			packageModel.Exists != nil &&
			packageModel.Author != nil &&
			packageModel.GodocURL != nil &&
			packageModel.IndexTime != nil &&
			packageModel.AwesomeGo != nil &&
			packageModel.SearchBlob != nil &&
			packageModel.Description != nil {
			batch.Query(
				cqlQueryInsertSearchPackage,
				*packageModel.Repo,
				*packageModel.Exists,
				*packageModel.Author,
				packageModel.Versions,
				*packageModel.GodocURL,
				*packageModel.IndexTime,
				*packageModel.AwesomeGo,
				*packageModel.SearchBlob,
				*packageModel.Description,
			)
		} else {
			return NewInvalidParameterError(
				"packageModels",
				fmt.Sprintf("[ ..., %v, ... ]", packageModel),
			)
		}
	}

	err := session.ExecuteBatch(batch)
	if err != nil {
		return err
	}

	return nil
}

/********************************** HELPERS ***********************************/

// TODO(skeswa): implement this for querying single packages
func scanPackageModel(query *gocql.Query) ([]*PackageModel, error) {
	return nil, nil
}

func scanPackageModels(query *gocql.Query) ([]*PackageModel, error) {
	var (
		err          error
		scanError    error
		closeError   error
		packageModel *PackageModel

		repo        string
		author      string
		description string

		iter          = query.Iter()
		packageModels = make([]*PackageModel, 0)
	)

	for iter.Scan(&repo, &author, &description) {
		packageModel, err = NewPackageModelFromBulkSelect(repo, author, description)
		if err != nil {
			scanError = err
			break
		} else {
			packageModels = append(packageModels, packageModel)
		}
	}

	if err = iter.Close(); err != nil {
		closeError = err
	}

	if scanError != nil || closeError != nil {
		return nil, NewQueryScanError(scanError, closeError)
	}

	return packageModels, nil
}
