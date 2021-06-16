package connect

func isRegistered() bool {
	return fileExists(defaulCredPath)
}
