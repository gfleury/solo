/*
 *
 * solo Server API
 *
 */
package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}

func NewP2PRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range p2pRoutes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
}

var p2pRoutes = Routes{
	Route{
		"RegisterNode",
		"POST",
		"/api/v1/node/register",
		RegisterNode,
	},

	Route{
		"GetConnectionConfigurationChallenge",
		"POST",
		"/api/v1/node/connnection_configuration",
		GetConnectionConfigurationChallenge,
	},

	Route{
		"GetConnectionConfiguration",
		"PUT",
		"/api/v1/node/connnection_configuration",
		GetConnectionConfiguration,
	},
}

var routes = Routes{
	Route{
		"Index",
		"GET",
		"/api/v1/",
		Index,
	},

	Route{
		"AddNetwork",
		strings.ToUpper("Post"),
		"/api/v1/network",
		AddNetwork,
	},

	Route{
		"DeleteNetwork",
		strings.ToUpper("Delete"),
		"/api/v1/network/{networkId}",
		DeleteNetwork,
	},

	Route{
		"GetNetworkById",
		strings.ToUpper("Get"),
		"/api/v1/network/{networkId}",
		GetNetworkById,
	},

	Route{
		"GetNextFreeIPAddress",
		strings.ToUpper("Get"),
		"/api/v1/network/{networkId}/nextip",
		GetNextFreeIPAddress,
	},

	Route{
		"NetworkAssignNodeFromRegistrationCode",
		strings.ToUpper("Put"),
		"/api/v1/network/{networkId}/register/{code}",
		NetworkAssignNodeFromRegistrationCode,
	},

	Route{
		"GetNetworks",
		strings.ToUpper("Get"),
		"/api/v1/networks",
		GetNetworks,
	},

	Route{
		"UpdateNetwork",
		strings.ToUpper("Put"),
		"/api/v1/network",
		UpdateNetwork,
	},

	Route{
		"CreateUser",
		strings.ToUpper("Post"),
		"/api/v1/user",
		CreateUser,
	},

	Route{
		"DeleteUser",
		strings.ToUpper("Delete"),
		"/api/v1/user/{username}",
		DeleteUser,
	},

	Route{
		"LoginUser",
		strings.ToUpper("Get"),
		"/api/v1/user/login",
		LoginUser,
	},

	Route{
		"GetUserByName",
		strings.ToUpper("Get"),
		"/api/v1/user/{username}",
		GetUserByName,
	},

	Route{
		"LogoutUser",
		strings.ToUpper("Get"),
		"/api/v1/user/logout",
		LogoutUser,
	},

	Route{
		"UpdateUser",
		strings.ToUpper("Put"),
		"/api/v1/user/{username}",
		UpdateUser,
	},

	Route{
		"ImageProxy",
		strings.ToUpper("Get"),
		"/api/v1/image",
		ImageProxy,
	},

	Route{
		"GetNodeRegistration",
		strings.ToUpper("Get"),
		"/api/v1/node/register",
		GetNodeRegistration,
	},

	Route{
		"UpdateNode",
		strings.ToUpper("Put"),
		"/api/v1/node",
		UpdateNode,
	},

	Route{
		"UpdateNode",
		strings.ToUpper("Get"),
		"/api/v1/nodes",
		GetNodes,
	},

	Route{
		"DeleteNode",
		strings.ToUpper("Delete"),
		"/api/v1/node/{nodeId}",
		DeleteNode,
	},
}
