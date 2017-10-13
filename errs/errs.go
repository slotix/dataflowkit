package errs

//400 Bad Request
//The server cannot or will not process the request due to an apparent client error (e.g., malformed request syntax, size too large, invalid request message framing, or deceptive request routing).
type BadRequest struct {
	Err error
}
func (e *BadRequest) Error() string { return e.Err.Error() }

//403 Forbidden
type ForbiddenByRobots struct {
	URL string
}
func (e *ForbiddenByRobots) Error() string { return e.URL + ": forbidden by robots.txt" }


type Forbidden struct {
	URL string
}
func (e *Forbidden) Error() string { return e.URL + ": forbidden" }


//404 Not Found
type NotFound struct {
	URL    string
}
func (e *NotFound) Error() string {
	return e.URL + ": resource not found" 	
}

//400 Invalid Host
type InvalidHost struct {
	URL    string
}
func (e *InvalidHost) Error() string {
	return e.URL + ": Invalid Host" 	
}

//504 Gateway Time-out
type GatewayTimeout struct {
}
func (e *GatewayTimeout) Error() string {
	return "Timeout exceeded rendering page" 	
}

//All the rest. Unspecified errors 
type Error struct {
	Err string
}
func (e *Error) Error() string {
	return e.Err
}
