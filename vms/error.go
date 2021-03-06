// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package vms

import (
	"net/http"

	"github.com/c4milo/osx-builder/apperror"
)

var ErrInternal = apperror.Error{
	Code:       "internal-error",
	Message:    "Whops! Our team is currently looking into this. Apologies for the inconvenience",
	HTTPStatus: http.StatusInternalServerError,
}

var ErrVMNotFound = apperror.Error{
	Code:       "vm-not-found",
	Message:    "The requested virtual machine ID was not found",
	HTTPStatus: http.StatusNotFound,
}

var ErrReadingReqBody = apperror.Error{
	Code:       "request-io-error",
	Message:    "There was an IO error while reading request's body. Please try again.",
	HTTPStatus: http.StatusBadRequest,
}

var ErrParsingJSON = apperror.Error{
	Code:       "invalid-json",
	Message:    "There was an error parsing the provided JSON message. Please try again.",
	HTTPStatus: http.StatusUnsupportedMediaType,
}

var ErrCreatingVM = apperror.Error{
	Code:       "vm-create-error",
	Message:    "There was an unexpected error trying to create the virtual machine. We are looking into it.",
	HTTPStatus: http.StatusInternalServerError,
}

var ErrOpeningVM = apperror.Error{
	Code: "vm-open-error",
	Message: "The VM was found but we were unable to open its configuration file. " +
		"Caused, most likely, by a corrupt VMX file or a stalled lock.",
	HTTPStatus: http.StatusConflict,
}

var ErrCbURL = apperror.Error{
	Code:    "err-marshalling-response",
	Message: "There was an error marshaling the response. Please try again creating your virtual machine.",
}
