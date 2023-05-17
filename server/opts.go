package server

// Opt represents an option for the server.
type Opt func(s *Server)

// Verbose makes the server more noisy with the request logs.
func Verbose() Opt {
	return func(s *Server) {
		s.verbose = true
	}
}
