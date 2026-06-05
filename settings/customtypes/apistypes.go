package customtypes

// RouterNames is the key type used in GetApisMap() to identify which
// middleware tier a set of routes belongs to.
//
// A "tier" is a named Gin router group created in GINServer.Setup().
// Every key you use here must match exactly one of the string keys passed
// to GINServer.Setup()'s middlewares map.  For example, if Setup() is
// called with:
//
//	server.Setup(map[string][]Middleware{
//	    "open": {rateLimitMiddleware},
//	    "auth": {jwtMiddleware},
//	})
//
// then valid RouterNames values are "open" and "auth".  Using any other
// value causes the route to be silently skipped at registration time.
//
// Convention (not enforced by the library):
//
//	"open"  — public endpoints, no authentication required
//	"auth"  — endpoints that require a valid JWT / session token
//	"base"  — internal / health-check endpoints
type RouterNames string

// ApiTypes represents the HTTP method for a single route.
type ApiTypes string

const (
	POST   ApiTypes = "POST"
	GET    ApiTypes = "GET"
	PUT    ApiTypes = "PUT"
	DELETE ApiTypes = "DELETE"
)
