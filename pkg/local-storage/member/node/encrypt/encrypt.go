package encrypt

type Encryptor interface {
	// EncryptVolume encrypts the volume with the given secret.
	EncryptVolume(volumePath string, secret string) error

	// DecryptVolume decrypts the volume with the given secret.
	DecryptVolume(volumePath string, secret string) error

	// IsVolumeEncrypted checks if the volume is encrypted.
	IsVolumeEncrypted(volumePath string) (bool, error)

	// CloseVolume closes the volume.
	CloseVolume(volumePath string) error

	// OpenVolume opens the volume with given secret and returns the decrypt volume name.
	OpenVolume(volumePath string, secret string) (string, error)
}
