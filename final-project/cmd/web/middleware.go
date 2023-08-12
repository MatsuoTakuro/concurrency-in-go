package main

import "net/http"

func (s *Server) SessionLoad(next http.Handler) http.Handler {

	/* Let me explain what the LoadAndSave method does in more detail here.
	1, load a sesson cookie for the current request.
	2, retrieve context values with the given session token.
	3, call the next handler in the chain.
	4, check the session status in the context
	5, save the session cookie to the response if the session has been modified or destroyed.
	*/
	return s.Session.LoadAndSave(next)
}
