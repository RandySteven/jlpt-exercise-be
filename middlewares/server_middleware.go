package middlewares

type ServerMiddleware struct {
	middlewares *Middlewares
}

func RegisterServerMiddleware(middlewares *Middlewares) *ServerMiddleware {
	return &ServerMiddleware{middlewares: middlewares}
}
