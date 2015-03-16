package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
)

// AES是对称加密算法
// AES-128。key长度：16, 24, 32 bytes 对应 AES-128, AES-192, AES-256
// 记住每次加密解密前都要设置iv.

// 该包默认的密匙
const defaultAesKey = "12345abcdef67890"

func AesEncrypt(plaintext []byte) ([]byte, error) {
	key := []byte(defaultAesKey)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	plaintext = PKCS5Padding(plaintext, blockSize)

	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	ciphertext := make([]byte, len(plaintext))

	blockMode.CryptBlocks(ciphertext, plaintext)
	return ciphertext, nil
}

func AesDecrypt(ciphertext []byte) ([]byte, error) {
	key := []byte(defaultAesKey)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	plaintext := make([]byte, len(ciphertext))

	blockMode.CryptBlocks(plaintext, ciphertext)
	plaintext = PKCS5UnPadding(plaintext)
	return plaintext, nil
}

func PKCS5Padding(plaintext []byte, blockSize int) []byte {
	padding := blockSize - len(plaintext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plaintext, padtext...)
}

func PKCS5UnPadding(plaintext []byte) []byte {
	length := len(plaintext)
	// 去掉最后一个字节 unpadding 次
	unpadding := int(plaintext[length-1])
	return plaintext[:(length - unpadding)]
}

func ToMd5(b []byte) string {
	h := md5.New()
	h.Write(b)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func ToSha1(b []byte) string {
	h := sha1.New()
	h.Write(b)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func ToSha2(b []byte) string {
	h := sha256.New()
	h.Write(b)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func CryptPass(str string) string {
	ismd5 := true
	if len(str) == 32 {
		for _, ch := range str {
			if !(ch >= '0' && ch <= '9' || ch >= 'a' && ch <= 'f' || ch >= 'A' && ch <= 'Z') {
				ismd5 = false
				break
			}
		}
	} else {
		ismd5 = false
	}
	if !ismd5 {
		str = ToMd5([]byte(str))
	}
	return str
}

func CryptBankPass(str string) string {
	ismd5 := true
	if len(str) == 32 {
		for _, ch := range str {
			if !(ch >= '0' && ch <= '9' || ch >= 'a' && ch <= 'f' || ch >= 'A' && ch <= 'Z') {
				ismd5 = false
				break
			}
		}
	} else {
		ismd5 = false
	}
	if !ismd5 {
		str = ToMd5([]byte(str))
	}
	buff := []byte(str)
	for i, b := range buff {
		buff[i] = ^b
	}
	return ToSha1(buff)
}
