/*
Egdegrid - Akamai REST API client

Use Akamai edgegrid API client authentication to make request to REST API

Usage:

	edgegrid <METHOD> <:PATH> [HEADER:VALUE]...


make an http request using METHOD to the specified PATH (must be prefixed by a colon char ':')
Optional additional headers could be specified by specifying colon separated pairs.

In case of a POST, PUT or PATCH request, the body should be filled via piping the content to the program

Egdegrid credential file could be specified via RC_PATH or default "~/.edgerc" will be used,
default section will always be used

Exit code is zero (0) if HTTP response status is 200, otherwise exitcode is one (1)

*/
package main

// use https://github.com/akamai/AkamaiOPEN-edgegrid-golang/tree/v1
import (
	"fmt"
	"github.com/akamai/AkamaiOPEN-edgegrid-golang/client-v1"
	"github.com/akamai/AkamaiOPEN-edgegrid-golang/edgegrid"
    "github.com/pkg/errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// GetEdgercPath find which edgerc file to use.
func GetEdgercPath() string {
	rcPath := os.Getenv("RC_PATH") // either use the one pointed by RC_PATH environment variable.
	if rcPath != "" {
		return rcPath
	}

	return "~/.edgerc" // or use the default one.
}


func getReqBody(verb string) ([]byte, error) {
	if !slices.Contains([]string{"POST", "PUT", "PATCH"}, verb) {
        return nil, nil
    }
	body, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, errors.Wrap(err, "Error reading request body")
	}
    if len(body) == 0 {
        return nil, nil
    }

    return body, nil
}
// will include request body if verb requires it (assuming json body).
func getReq(verb string, path string, config *edgegrid.Config) (*http.Request, error) {
    body, err := getReqBody(verb)
    if err != nil {
        return nil, errors.Wrap(err, "Error creating request")
    }

	return client.NewJSONRequest(*config, verb, path[1:], body)
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage:\n  %s <METHOD> <:PATH> [HEADER:VALUE]...", filepath.Base(os.Args[0]))
		os.Exit(1)
	}
	verb := os.Args[1]
	path := os.Args[2]
	// limit to usable verbs with akamai APis.
	if !slices.Contains([]string{"GET", "POST", "PUT", "PATCH", "DELETE"}, verb) {
		panic("unsupported method")
	}

	// path must be in :/path format (consistent with httpie-edgegrid).
	if path[0] != ':' {
		panic("path must start with a colon char (':')")
	}

	config, err := edgegrid.Init(GetEdgercPath(), "default")
	if err != nil {
		panic(err)
	}

	req, err := getReq(verb, path, &config)
	if err != nil {
		panic(err)
	}

	if len(os.Args) > 3 {
		for _, element := range os.Args[3:] {
			pair := strings.SplitN(element, ":", 2)
			if len(pair) == 2 {
				req.Header.Set(pair[0], pair[1])
			}
		}
	}

	resp, err := client.Do(config, req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	_, err = io.Copy(os.Stdout, resp.Body)
    if err != nil {
        panic(err)
    }
	if resp.StatusCode != http.StatusContinue {
		defer os.Exit(1)
	}
}
