package frontend

import (
	"fmt"
	"net/http"

	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
)

// HandleError serves failed requests
func HandleError(w http.ResponseWriter, r *http.Request, status int, errs ...error) {
	d := errorModel{
		Status:        status,
		Title:         fmt.Sprintf("Error %d", status),
		InvertedTheme: isInverted(r),
		Errors:        errs,
	}
	w.WriteHeader(status)

	for _, err := range errs {
		if err != nil {
			Logger.WithFields(log.Fields{"trace": errors.ErrorStack(err)}).Errorf("Err: %s", err)
		}
	}

	w.Header().Set("Cache-Control", " no-store, must-revalidate")
	w.Header().Set("Pragma", " no-cache")
	w.Header().Set("Expires", " 0")
	RenderTemplate(r, w, "error", d)
}
