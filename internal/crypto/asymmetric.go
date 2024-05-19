// Package crypto создает публичный и приватный ключи шифрования
package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"os"
	"time"

	"github.com/dnsoftware/go-metrics/internal/constants"
)

// MakePublicKey создание публичного ключа из файла
func MakePublicKey(fullPathCert string) (*rsa.PublicKey, error) {
	publicKeyPEM, err := os.ReadFile(fullPathCert)
	if err != nil {
		return nil, err
	}
	publicKeyBlock, _ := pem.Decode(publicKeyPEM)
	cert, err := x509.ParseCertificate(publicKeyBlock.Bytes)
	if err != nil {
		return nil, err
	}

	pk, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("no rsa publickey type")
	}

	//	enc, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pk, []byte("super secret message"), nil)

	return pk, nil
}

func MakePrivateKey(fullPathPriv string) (*rsa.PrivateKey, error) {
	privateKeyPEM, err := os.ReadFile(fullPathPriv)
	if err != nil {
		return nil, err
	}
	privateKeyBlock, _ := pem.Decode(privateKeyPEM)
	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// CertFilesGenerate генерация файлов сертификата и приватного ключа
// certFilename - имя файла сетификата
// privateFilename - имя файла приватного ключа
// возвращает полные пути к файлам сертификата и приватного ключа
func CertFilesGenerate(fullPathCert, fullPathPriv string) (string, string, error) {

	// создаём шаблон сертификата
	cert := &x509.Certificate{
		// указываем уникальный номер сертификата
		SerialNumber: big.NewInt(1658),
		// заполняем базовую информацию о владельце сертификата
		Subject: pkix.Name{
			Organization: []string{"DN Software"},
			Country:      []string{"RU"},
		},
		// разрешаем использование сертификата для 127.0.0.1 и ::1
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		// сертификат верен, начиная со времени создания
		NotBefore: time.Now(),
		// время жизни сертификата — 10 лет
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		// устанавливаем использование ключа для цифровой подписи,
		// а также клиентской и серверной авторизации
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	// создаём новый приватный RSA-ключ длиной 4096 бит
	// обратите внимание, что для генерации ключа и сертификата
	// используется rand.Reader в качестве источника случайных данных
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return "", "", err
	}

	// создаём сертификат x.509
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		return "", "", err
	}

	// кодируем сертификат и ключ в формате PEM, который
	// используется для хранения и обмена криптографическими ключами
	var certPEM bytes.Buffer
	err = pem.Encode(&certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		return "", "", err
	}

	err = os.WriteFile(fullPathCert, certPEM.Bytes(), 0644)
	if err != nil {
		return "", "", err
	}

	var privateKeyPEM bytes.Buffer
	err = pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		return "", "", err
	}

	err = os.WriteFile(fullPathPriv, privateKeyPEM.Bytes(), 0644)
	if err != nil {
		return "", "", err
	}

	return fullPathCert, fullPathPriv, nil
}

// DefaultCryptoFilesName генерация путей по умолчанию к файлам асимметричных ключей
func DefaultCryptoFilesName() (string, string, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return "", "", err
	}

	fullPathCert := workDir + "/" + constants.CryptoPublicFile
	fullPathPriv := workDir + "/" + constants.CryptoPrivateFile

	return fullPathCert, fullPathPriv, nil
}
