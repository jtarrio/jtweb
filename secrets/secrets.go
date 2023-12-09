package secrets

type SecretSupplier interface {
	GetSecret(key string) (string, error)
}
