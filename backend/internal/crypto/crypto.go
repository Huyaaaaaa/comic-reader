package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"regexp"
	"strings"
)

// ExtractKeys 从HTML中提取AES密钥
func ExtractKeys(html string) (iv, key string, err error) {
	// 提取 IV (aei)
	ivRe := regexp.MustCompile(`var\s+aei\s*=\s*'([^']+)'`)
	ivMatch := ivRe.FindStringSubmatch(html)
	if len(ivMatch) < 2 {
		return "", "", errors.New("无法提取 IV (aei)")
	}
	iv = ivMatch[1]

	// 提取 Key (aek)
	keyRe := regexp.MustCompile(`var\s+aek\s*=\s*'([^']+)'`)
	keyMatch := keyRe.FindStringSubmatch(html)
	if len(keyMatch) < 2 {
		return "", "", errors.New("无法提取 Key (aek)")
	}
	key = keyMatch[1]

	return iv, key, nil
}

// Decrypt AES-CBC解密
func Decrypt(ciphertextB64, key, iv string) (string, error) {
	ciphertextB64 = strings.TrimSpace(ciphertextB64)
	if ciphertextB64 == "" {
		return "", nil
	}

	// Base64解码
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		// 解密失败返回原文
		return ciphertextB64, nil
	}

	// 创建AES cipher
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return ciphertextB64, nil
	}

	// 检查密文长度
	if len(ciphertext) < aes.BlockSize {
		return ciphertextB64, nil
	}

	// CBC模式解密
	mode := cipher.NewCBCDecrypter(block, []byte(iv))
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// 去除PKCS7填充
	plaintext, err = pkcs7Unpad(plaintext)
	if err != nil {
		return ciphertextB64, nil
	}

	return string(plaintext), nil
}

// pkcs7Unpad 去除PKCS7填充
func pkcs7Unpad(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("数据为空")
	}

	padding := int(data[length-1])
	if padding > length || padding > aes.BlockSize {
		return nil, errors.New("无效的填充")
	}

	// 验证填充
	for i := 0; i < padding; i++ {
		if data[length-1-i] != byte(padding) {
			return nil, errors.New("无效的填充")
		}
	}

	return data[:length-padding], nil
}
