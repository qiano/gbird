package util

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"time"
	"crypto"
)

//Md5 md5加密
func Md5(text string) string {
	hashMd5 := md5.New()
	io.WriteString(hashMd5, text)
	return fmt.Sprintf("%x", hashMd5.Sum(nil))
}

//GenUUID 产生唯一的id
func GenUUID() string {
	buf := make([]byte, 16)
	io.ReadFull(rand.Reader, buf)
	return fmt.Sprintf("%x%x", buf, time.Now().UnixNano())
}

//RsaEncrypt 加密
func RsaEncrypt(origData []byte, publicKey []byte) ([]byte, error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub := pubInterface.(*rsa.PublicKey)

	h := crypto.Hash.New(crypto.SHA1)
    h.Write(origData)
    // hashed := h.Sum(nil)

	
	return rsa.EncryptOAEP(h, rand.Reader, pub, origData, []byte(""))
	// return rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
}

//RsaDecrypt 解密
func RsaDecrypt(ciphertext []byte, privateKey []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("private key error!")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
}
