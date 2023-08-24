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

func (s *Server) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.Session.Exists(r.Context(), USER_ID_CTX) {
			s.Session.Put(r.Context(), ERROR_CTX, LOGIN_FIRST_MSG)
			http.Redirect(w, r, LOGIN_PATH, http.StatusTemporaryRedirect)
			/* NOTE: status code 307 (vs 303)
				In the case of a 307 Temporary Redirect, the client should continue to use the original URL for future requests.
			This means that the client should use the same URL that was used in the original request, including the same HTTP method, headers, and body.

				In the example code you provided, the original URL for future requests is the URL that the client used to access the listOfPlans handler.
			When the client is redirected to the login page, the http.Redirect function is called with the original request (r) and a 307 Temporary Redirect status code.
			This ensures that the client continues to use the original URL for future requests, including any query parameters, headers, and request body.
			*/
			return
		}
		next.ServeHTTP(w, r)
	})
}
