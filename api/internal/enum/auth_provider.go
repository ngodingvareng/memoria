package enum

type AuthProvider string

const (
	AuthProviderGoogle     AuthProvider = "google"
	AuthProviderCredential AuthProvider = "credential"
)
