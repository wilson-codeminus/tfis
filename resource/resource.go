package resource

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const TerraformBaseUrl = "https://www.terraform.io"

type docNotFoundError struct {
	resType string
}

type importSyntaxNotFoundError struct {
	String string
}

type TerraformResource struct {
	Type     string
	Provider string
	Name     string
}

func (e *docNotFoundError) Error() string {
	return fmt.Sprintf("Unable to find documentation for %v", e.resType)
}

func (e *importSyntaxNotFoundError) Error() string {
	return "Unable to find import syntax in documentation"
}

func properties(resourceType string) []string {
	return strings.SplitN(resourceType, "_", 2)
}

// New creates a new TerraformResource
func New(resourceType string) *TerraformResource {
	tr := &TerraformResource{}
	tr.Type = resourceType

	props := properties(tr.Type)
	tr.Provider = props[0]
	tr.Name = props[1]

	return tr
}

// DocURL returns the resource documentation URL based on expected patterns
func (r *TerraformResource) DocURL() (string, error) {

	possibleUrls := []string{
		TerraformBaseUrl + "/docs/providers/" + r.Provider + "/r/" + r.Name + ".html",
		TerraformBaseUrl + "/docs/providers/" + r.Provider + "/r/" + r.Type + ".html",
	}

	for _, url := range possibleUrls {
		resp, _ := http.Get(url)
		if resp.StatusCode == 200 {
			return url, nil
		}
	}

	return "", &docNotFoundError{r.Type}

}

// func (r *TerraformResource) Docs(reader *io.Reader) (string, error) {

// }

// ImportSyntaxes extracts the resource import syntax from the docs
func (r *TerraformResource) ImportSyntaxes(reader io.Reader) ([]string, error) {
	url, err := r.DocURL()
	if err != nil {
		return nil, err
	}

	if reader == nil {
		req, err := http.NewRequest("GET", url, nil)

		if err != nil {
			return nil, err
		}

		client := &http.Client{}
		res, err := client.Do(req)

		if err != nil {
			return nil, err
		}

		reader = res.Body

	}

	doc, err := goquery.NewDocumentFromReader(reader)
	// doc, err := goquery.NewDocument(url)
	if err != nil {
		return nil, err
	}

	syntaxes := []string{}
	doc.Find("pre").Each(func(i int, item *goquery.Selection) {
		if strings.Contains(item.Text(), "terraform import "+r.Type) {
			for _, i := range strings.Split(strings.TrimSpace(item.Text()), "\n") {
				syntaxes = append(syntaxes, strings.TrimSpace(i)[2:])
			}
		}
	})

	if len(syntaxes) == 0 {
		return nil, &importSyntaxNotFoundError{}
	}
	return syntaxes, nil
}
