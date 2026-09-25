package api

import "net/http"

// routesOps registers this domain's endpoints. See docs/API.md.
func (s *Server) routesOps(mux *http.ServeMux) {
	_ = mux
}
