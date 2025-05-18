package util

type Fetcher interface {
	Fetch(ref string, useCache bool) (string, error)
}
