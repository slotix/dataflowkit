package errs

// BadRequest 400
//
// The server cannot or will not process the request due to an apparent client error (e.g., malformed request syntax, size too large, invalid request message framing, or deceptive request routing).
type BadRequest struct {
	Err error
}

func (e *BadRequest) Error() string { return e.Err.Error() }

// ForbiddenByRobots 403
//
// Client does not have access rights to the content caused by robots.txt restrictions.
type ForbiddenByRobots struct {
	URL string
}

func (e *ForbiddenByRobots) Error() string { return e.URL + ": forbidden by robots.txt" }

// Forbidden 403
//
// Client does not have access rights to the content so server is rejecting to give proper response.
type Forbidden struct {
	URL string
}

func (e *Forbidden) Error() string { return e.URL + ": forbidden" }

// NotFound 404
//
// Server can not find requested resource. This response code probably is most famous one due to its frequency to occur in web.
type NotFound struct {
	URL string
}

func (e *NotFound) Error() string {
	return e.URL + ": Not found"
}

// BadGateway 502
//
// This error response means that the server, while working as a gateway to get a response needed to handle the request, got an invalid response.
type BadGateway struct {
}

func (e *BadGateway) Error() string {
	return "Invalid response from server"
}

// GatewayTimeout Gateway Time-out 504
//
// This error response is given when the server is acting as a gateway and cannot get a response in time.
type GatewayTimeout struct {
}

func (e *GatewayTimeout) Error() string {
	return "Timeout exceeded rendering page"
}

// Error represents all the rest (unspecified errors).
type Error struct {
	Err string
}

func (e *Error) Error() string {
	return e.Err
}
