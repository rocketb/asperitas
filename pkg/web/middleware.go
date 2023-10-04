package web

// Middleware is a function designed to run before and/or after some Handler.
type Middleware func(Handler) Handler

// wrapMiddleware creates new handler by wrapping middleware around a final Handler.
func wrapMiddleware(mw []Middleware, handler Handler) Handler {
	for i := len(mw) - 1; i >= 0; i-- {
		h := mw[i]
		if h != nil {
			handler = h(handler)
		}
	}
	return handler
}
