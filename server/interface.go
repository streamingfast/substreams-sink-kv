package server

type Serveable interface {
	Serve(listenAddr string) error
	Shutdown()
}
