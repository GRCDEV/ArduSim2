package ports

// LoggerClient abstracts the HTTP call to the logger container.
// remoteHost: "" or "localhost" -> http://localhost:8080
// remoteHost: "192.168.1.x"   -> http://192.168.1.x:8080
type LoggerClient interface {
	DownloadZip(remoteHost string) ([]byte, error)
}
