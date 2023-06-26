package server

import "os"

// Opt represents an option for the server.
type Opt func(s *Server)

// Verbose makes the server more noisy with the request logs.
func Verbose() Opt {
	return func(s *Server) {
		s.verbose = true
	}
}

// Favicon makes the server serves given file at /favicon.ico
func Favicon(path string) Opt {
	if _, err := os.Stat(path); err != nil {
		panic(err)
	}
	return func(s *Server) {
		s.faviconPath = path
	}
}
