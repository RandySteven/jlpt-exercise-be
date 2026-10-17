package middlewares

type ClientMiddleware struct {
	middlewares *Middlewares
}

func RegisterClientMiddleware(middlewares *Middlewares) *ClientMiddleware {
	return &ClientMiddleware{
		middlewares: middlewares,
	}
}
