package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestCreateKeyFile(t *testing.T) {

	certFilename, privateFilename, err := DefaultCryptoFilesName()
	require.NoError(t, err)

	fullPathCert, fullPathPriv, err := CertFilesGenerate(certFilename, privateFilename)
	require.NoError(t, err)

	_, err = os.Stat(fullPathCert)
	assert.NoError(t, err, fullPathCert+" file does not exist")

	_, err = os.Stat(fullPathPriv)
	assert.NoError(t, err, fullPathPriv+" file does not exist")
}

func TestMakePublicKey(t *testing.T) {
	fullPathCert, _, err := DefaultCryptoFilesName()
	require.NoError(t, err)

	_, err = MakePublicKey(fullPathCert)
	assert.NoError(t, err)

}

func TestMakePrivateKey(t *testing.T) {
	_, fullPathPriv, err := DefaultCryptoFilesName()
	require.NoError(t, err)

	_, err = MakePrivateKey(fullPathPriv)
	assert.NoError(t, err)
}

func TestEncryptDescript(t *testing.T) {
	fullPathCert, fullPathPriv, _ := DefaultCryptoFilesName()
	publicKey, _ := MakePublicKey(fullPathCert)
	privateKey, _ := MakePrivateKey(fullPathPriv)

	testMessage := "Golang forever))"

	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, []byte(testMessage), nil)
	require.NoError(t, err)

	decryptedBytes, err := privateKey.Decrypt(nil, ciphertext, &rsa.OAEPOptions{Hash: crypto.SHA256})

	require.NoError(t, err)
	assert.Equal(t, testMessage, string(decryptedBytes))
}
