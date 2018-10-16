package filter

import (
	"bytes"
	"io"
	"log"
	"net/url"
	"regexp"
	"strings"

	"html/template"
)

// Category holds categorization data of scraped urls
// Add categorization layer?
type Category struct {
	Name        string
	Regexp      *regexp.Regexp
	Allowed     bool
	MatchedURLs []string
	Description string
}

// Results filters results based on given vars
func Results(URLs []string, categories []Category) (string, error) {

	// url decoding
	URLs, err := decodeURLs(URLs)
	if err != nil {
		return "", err
	}

	// url trimming
	URLs, err = trimURLs(URLs)
	if err != nil {
		return "", err
	}

	// remove duplicates
	URLs, err = deDup(URLs)
	if err != nil {
		return "", err
	}

	// Filter out deny
	denyPassURLs, err := deny(URLs, categories)
	if err != nil {
		return "", err
	}

	// Keep allow
	filtCategories, err := allow(denyPassURLs, categories)
	if err != nil {
		return "", err
	}

	// create template
	var buff bytes.Buffer
	if err := createHTMLAttachment(filtCategories, &buff); err != nil {
		return "", err
	}

	// return template as raw string
	return buff.String(), nil
}

// Remove duplicate URLs
func deDup(URLs []string) ([]string, error) {
	var deDupURLs []string
	encountered := map[string]bool{}

	for _, u := range URLs {
		if encountered[u] == true {
			continue
		}

		encountered[u] = true
		deDupURLs = append(deDupURLs, u)
	}

	return deDupURLs, nil
}

// trim characters before http
func trimURLs(URLs []string) ([]string, error) {
	var trimmedURLs []string
	re := regexp.MustCompile(`http.*`)
	for _, u := range URLs {

		t := re.FindAllString(u, 1)
		if t == nil {
			continue
		}
		trimmedURLs = append(trimmedURLs, t[0])
	}

	return trimmedURLs, nil
}

// decodeURLs decode's url encoding to utf-8
func decodeURLs(URLs []string) ([]string, error) {
	var decodedURLs []string

	log.Printf("urls given for decoding: %d", len(URLs))
	for _, u := range URLs {
		// Decode URL
		d, err := url.QueryUnescape(u)
		if err != nil {
			log.Println(err)
		}

		decodedURLs = append(decodedURLs, d)
	}

	return decodedURLs, nil
}

// allow given URLs based on given categories
// TODO: break up this function
func allow(URLs []string, categories []Category) ([]Category, error) {
	// Remove categories with Allowed set to false
	var allowedCategories []Category
	for _, c := range categories {
		if c.Allowed {
			allowedCategories = append(allowedCategories, c)
		}
	}

	noCategory := &Category{
		Name:        "No Category",
		Regexp:      regexp.MustCompile(`sendgrid\.net/`),
		Allowed:     true,
		Description: "a category for urls with no category",
	}

	for i, c := range allowedCategories {
		if c.Allowed {
			for _, u := range URLs {
				if c.Regexp.MatchString(u) {
					allowedCategories[i].MatchedURLs = append(allowedCategories[i].MatchedURLs, u)
					continue
				}
				if noCategory.Regexp.MatchString(u) {
					noCategory.MatchedURLs = append(noCategory.MatchedURLs, u)
				}
			}
		}
	}

	// Compare urls with no category to with already sorted urls, only include urls not in sorted
	// in no category
	var noDup []string
	encountered := map[string]bool{}
	for _, c := range allowedCategories {
		for _, u := range c.MatchedURLs {
			if encountered[u] == true {
				continue
			}

			encountered[u] = true
		}
	}
	for _, un := range noCategory.MatchedURLs {
		if encountered[un] == true {
			continue
		}

		encountered[un] = true
		noDup = append(noDup, un)
	}
	noCategory.MatchedURLs = noDup
	allowedCategories = append(allowedCategories, *noCategory)

	return allowedCategories, nil
}

// deny given URLs based on given categories
func deny(URLs []string, categories []Category) ([]string, error) {
	var denyPassURLs []string

	for _, c := range categories {
		if !c.Allowed {
			for _, url := range URLs {
				if c.Regexp.MatchString(url) {
					//log.Printf("denied:" + url)
					continue
				}
				denyPassURLs = append(denyPassURLs, url)
			}
		}
	}

	return denyPassURLs, nil
}

// Template
// generates template based on number of categories
// TODO: STATS TEMPLATE
const attachment = `
{{range .}}
<HR>
<h1>{{.Name}}</h1>
	{{range .MatchedURLs}}
		<p>{{.}}</p>
	{{end}}
{{end}}
`

// createHTMLAttachment generates text output
func createHTMLAttachment(categories []Category, w io.Writer) error {

	// parse template
	funcMap := template.FuncMap{
		"join": strings.Join,
	}
	t, err := template.New("html").Funcs(funcMap).Parse(attachment)
	if err != nil {
		return err
	}

	if err := t.Execute(w, categories); err != nil {
		return err
	}

	return nil
}
