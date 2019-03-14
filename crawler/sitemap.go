package crawler

// SiteMap is the type that holds the sitemap - made this
// a type for future proofing in case functions want to be
// called on the sitemap
type SiteMap map[string]*Page

// PathsCrawled returns a slice of the paths crawled
func (s SiteMap) PathsCrawled() []string {
	var paths []string
	for url := range s {
		paths = append(paths, url)
	}
	return paths
}