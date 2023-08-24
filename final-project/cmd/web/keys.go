package main

// XXX_CTX is both a contexts key and a session key.
const (
	FLASH_CTX   = "flash"
	WARNING_CTX = "warning"
	ERROR_CTX   = "error"
	USER_ID_CTX = "user_id"
	USER_CTX    = "user"
	PLAN_ID_CTX = "id"
)

// XXX_ATTR is an attribute or element's name embedded in html.
const (
	EMAIL_ATTR      = "email"
	PASSWORD_ATTR   = "password"
	FIRST_NAME_ATTR = "first-name"
	LAST_NAME_ATTR  = "last-name"
	PLANS_ATTR      = "plans"
)
