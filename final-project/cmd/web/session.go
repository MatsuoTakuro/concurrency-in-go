package main

import (
	"encoding/gob"
	"net/http"
	"os"
	"time"

	"github.com/MatsuoTakuro/final-project/data"
	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"
)

func initSession() *scs.SessionManager {
	// WARN: need to register the User struct for session to work because it's a custom type
	gob.Register(data.User{}) // TODO: check if it's really necessary to initialize the redis every time new custom type is added or edited to data.

	session := scs.New()
	session.Store = redisstore.New(initRedis())    // set redis as the session store
	session.Lifetime = 24 * time.Hour              // set the session lifetime to 24 hours
	session.Cookie.Persist = true                  // set the session cookie to persist across browser sessions, even if the browser is closed
	session.Cookie.SameSite = http.SameSiteLaxMode // set the session cookie to be sent for same-site requests
	/*
		The SameSite attribute is a security feature introduced in cookies to help prevent Cross-Site Request Forgery (CSRF) attacks and Cross-Site Script Inclusion (XSSI) attacks.
		It controls the behavior of cookies and defines whether a browser should send the cookie along with cross-site requests.
		Here's a breakdown of the different SameSite modes and their use cases:

		1. SameSite=Strict:
		Cookies are sent only if the request originated from the same site as the target domain.
		Use Case: This mode is suitable for highly sensitive operations like banking websites where strict isolation from other sites is required.

		2. SameSite=Lax (Default in many contexts):
		Cookies are sent with same-site requests and with cross-site top-level navigations that are initiated by a link click, such as navigating to a different website.
		Cookies are not sent with cross-site subrequests (like images or frames) or AJAX/XHR requests.
		Use Case: This mode provides a balance between security and usability, making it suitable for most websites.
		It helps prevent CSRF attacks while still allowing some cross-site requests.

		3. SameSite=None:
		Cookies are sent with all requests, regardless of whether they are same-site or cross-site.
		This mode must be used in conjunction with the Secure attribute, meaning the cookie will only be sent over HTTPS.
		Use Case: This mode is suitable for third-party services that need to embed content across different websites and require cookies to be sent with those requests.

		Why Control SameSite?
		Controlling the SameSite attribute allows developers to fine-tune the behavior of cookies based on the specific needs and security requirements of their application.
		By setting the appropriate SameSite mode, developers can strike the right balance between security and functionality.
		For example, in a third-party authentication service that needs to work across various domains, setting SameSite=None would be necessary.
		Conversely, for a banking application where strict isolation is required, SameSite=Strict would be more appropriate.
		In the context of the code snippet you provided, setting SameSite to http.SameSiteLaxMode ensures that the session cookie adheres to the Lax behavior, providing a good balance between security and usability for most web applications.
	*/
	session.Cookie.Secure = true // set the session cookie to be sent only over HTTPS connections

	return session
}

func initRedis() *redis.Pool {
	redisPool := &redis.Pool{
		MaxIdle: 10, // maximum number of idle connections in the pool
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", os.Getenv("REDIS"))
		},
	}

	return redisPool
}
