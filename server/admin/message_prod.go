// +build production

package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"git.nkagami.me/natsukagami/kjudge"
	"github.com/pkg/errors"
)

type jsonRelease struct {
	TagName string `json:"tag_name"`
}

type version struct {
	TagName    string
	LastUpdate time.Time
}

var currentVersion version

// NewVersionMessageGet checks if there is a new version of kjudge
func NewVersionMessageGet() (string, error) {
	if currentVersion.TagName == "" || time.Now().After(currentVersion.LastUpdate.Add(time.Hour)) {
		response, err := http.Get("https://git.nkagami.me/api/v1/repos/natsukagami/kjudge/releases?page=1&per_page=1")
		if err != nil {
			return "", errors.WithStack(err)
		}
		defer response.Body.Close()
		var x []jsonRelease
		decode := json.NewDecoder(response.Body)
		if err := decode.Decode(&x); err != nil {
			return "", errors.WithStack(err)
		}
		if len(x) == 0 {
			return "", errors.New("Found no relase")
		}
		currentVersion.TagName = x[0].TagName[1:]
		currentVersion.LastUpdate = time.Now()
	}
	if kjudge.Version != currentVersion.TagName {
		return fmt.Sprintf("Please update kjudge to the newest version (%s)", currentVersion.TagName), nil
	} else {
		return "", nil
	}
}
