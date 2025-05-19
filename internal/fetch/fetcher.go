package fetch

type Fetcher interface {
	Fetch(ref string, useCache bool) (string, error)
}
