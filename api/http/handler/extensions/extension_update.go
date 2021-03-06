package extensions

import (
	"errors"
	"net/http"

	"github.com/asaskevich/govalidator"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	"github.com/portainer/portainer/api"
	bolterrors "github.com/portainer/portainer/api/bolt/errors"
)

type extensionUpdatePayload struct {
	Version string
}

func (payload *extensionUpdatePayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Version) {
		return errors.New("Invalid extension version")
	}

	return nil
}

func (handler *Handler) extensionUpdate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	extensionIdentifier, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid extension identifier route variable", err}
	}
	extensionID := portainer.ExtensionID(extensionIdentifier)

	var payload extensionUpdatePayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid request payload", err}
	}

	extension, err := handler.DataStore.Extension().Extension(extensionID)
	if err == bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find a extension with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find a extension with the specified identifier inside the database", err}
	}

	err = handler.ExtensionManager.UpdateExtension(extension, payload.Version)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to update extension", err}
	}

	err = handler.DataStore.Extension().Persist(extension)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist extension status inside the database", err}
	}

	return response.Empty(w)
}
