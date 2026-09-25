package api

import "net/http"

// routesCore registers this domain's endpoints. See docs/API.md.
func (s *Server) routesCore(mux *http.ServeMux) {
	_ = mux
}
