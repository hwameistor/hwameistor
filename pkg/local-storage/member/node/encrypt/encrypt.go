package encrypt

type Encryptor interface {
	// EncryptVolume encrypts the volume with the given secret.
	EncryptVolume(volumeName string, secret string) error

	// DecryptVolume decrypts the volume with the given secret.
	DecryptVolume(volumeName string, secret string) error

	// CloseVolume closes the volume.
	CloseVolume(volumeName string) error

	// OpenVolume opens the volume with given secret and returns the decrypt volume name.
	OpenVolume(volumeName string, secret string) (string, error)
}
