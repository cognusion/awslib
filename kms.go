package awslib

import (
	"encoding/base64"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
)

// KMSSession is a helper to easily [en|de]crypt strings using KMS
type KMSSession struct {
	client *kms.KMS
	keyId  string
}

// NewKMSSessions takes a keyID and returns a KMSSession.
// Assumes InitAWS has been called.
func NewKMSSession(keyId string) (ksession *KMSSession) {

	ksession = &KMSSession{
		client: kms.New(AWSSession),
		keyId:  keyId,
	}

	return

}

// Decrypt does just that on the provided ciphertext string
func (k *KMSSession) Decrypt(ciphertext string) (decrypted string, err error) {

	decoded, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return
	}

	resp, err := k.client.Decrypt(&kms.DecryptInput{
		CiphertextBlob: decoded,
	})
	if err != nil {
		return
	}

	decrypted = string(resp.Plaintext)
	return

}

// Encrypt does just that on the provided plaintext string
func (k *KMSSession) Encrypt(plaintext string) (encoded string, err error) {

	resp, err := k.client.Encrypt(&kms.EncryptInput{
		Plaintext: []byte(plaintext),
		KeyId:     aws.String(k.keyId),
	})
	if err != nil {
		return
	}

	encoded = base64.StdEncoding.EncodeToString(resp.CiphertextBlob)
	return

}
