package splash

//400 Bad Request
//The server cannot or will not process the request due to an apparent client error (e.g., malformed request syntax, size too large, invalid request message framing, or deceptive request routing).
type ErrorBadRequest struct {
	Err error
}
func (e *ErrorBadRequest) Error() string { return e.Err.Error() }

//403 Forbidden
type ErrorForbiddenByRobots struct {
	URL string
}
func (e *ErrorForbiddenByRobots) Error() string { return e.URL + ": forbidden by robots.txt" }

//404 Not Found
type ErrorNotFound struct {
	URL    string
}
func (e *ErrorNotFound) Error() string {
	return e.URL + ": resource not found" 	
}

//400 Invalid Host
type ErrorInvalidHost struct {
	URL    string
}
func (e *ErrorInvalidHost) Error() string {
	return e.URL + ": Invalid Host" 	
}

//All the rest. Unspecified errors 
type Error struct {
	Err string
}
func (e *Error) Error() string {
	return e.Err
}


