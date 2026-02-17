Middleware — это функция, которая:

делает что-то ДО handler
делает что-то ПОСЛЕ handler


например:

Request
 ↓
middleware (log start)
 ↓
handler
 ↓
middleware (log latency)
 ↓
Response

Middleware — это просто wrapper:

func wrapper(handler) handler {
    return newHandler {
        before
        handler()
        after
    }
}

Это pattern называется: decorator pattern