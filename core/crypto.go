package core

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"github.com/ugorji/go/codec"
	"gopcr/config"
	"io"
	"math/big"
	"strconv"
)

// pcrCrypto PCR Msgpack
type pcrCrypto struct {
	mh codec.MsgpackHandle // MessagePack处理器
}

// newPCRCrypto 创建一个新的PCR加密器
func newPCRCrypto() *pcrCrypto {
	var mh codec.MsgpackHandle
	// 设置handle选项，使其行为与标准msgpack一致
	mh.WriteExt = true
	mh.RawToString = true

	return &pcrCrypto{
		mh: mh,
	}
}

// calcSID  计算SID
func calcSID(sid string) string {
	h := md5.New()
	_, err := io.WriteString(h, sid+"c!SID!n")
	if err != nil {
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))
}

// createKey 生成一个随机32字节的密钥
func createKey() ([]byte, error) {
	key := make([]byte, 32)
	for i := range key {
		n, err := rand.Int(rand.Reader, big.NewInt(16))
		if err != nil {
			return nil, err
		}
		key[i] = "0123456789abcdef"[n.Int64()]
	}
	return key, nil
}

// padData PKCS#7 Padding
func padData(data []byte) []byte {
	padding := 16 - len(data)%16
	return append(data, bytes.Repeat([]byte{byte(padding)}, padding)...)
}

// unpadData 改进版 PKCS#7 Unpadding
func unpadData(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("数据长度为0")
	}

	padding := int(data[len(data)-1])
	if padding > 16 || padding == 0 {
		return nil, errors.New("无效的填充值")
	}

	// 验证所有填充字节是否一致
	for i := 0; i < padding; i++ {
		if data[len(data)-1-i] != byte(padding) {
			return nil, errors.New("无效的填充")
		}
	}

	return data[:len(data)-padding], nil
}

// encrypt 加密数据
func (c *pcrCrypto) encrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	paddedData := padData(data)

	encrypted := make([]byte, len(paddedData))
	// 执行加密
	cipher.NewCBCEncrypter(block, []byte(config.PcrAesIV)).CryptBlocks(encrypted, paddedData)

	// 尾部添加key
	return append(encrypted, key...), nil
}

// decrypt 解密数据
func (c *pcrCrypto) decrypt(data []byte) ([]byte, error) {
	if len(data) < 32 {
		return nil, errors.New("密文太短")
	}

	// 从数据末尾提取密钥
	encryptedData, key := data[:len(data)-32], data[len(data)-32:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	decrypted := make([]byte, len(encryptedData))
	cipher.NewCBCDecrypter(block, []byte(config.PcrAesIV)).CryptBlocks(decrypted, encryptedData)
	return decrypted, nil
}

// encodeToMsgpack 将对象编码为MessagePack格式
func (c *pcrCrypto) encodeToMsgpack(v any) ([]byte, error) {
	var buf bytes.Buffer
	if err := codec.NewEncoder(&buf, &c.mh).Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// decodeFromMsgpack 从MessagePack格式解码为对象
func (c *pcrCrypto) decodeFromMsgpack(data []byte, v any) error {
	err := codec.NewDecoderBytes(data, &c.mh).Decode(v)
	if err != nil {
		return err
	}
	return nil
}

// EncryptData 加密对象数据（先编码为msgpack，再加密）
func (c *pcrCrypto) EncryptData(v any) ([]byte, error) {
	// 生成随机密钥
	key, err := createKey()
	if err != nil {
		return nil, err
	}

	// 先编码为msgpack
	encoded, err := c.encodeToMsgpack(v)
	if err != nil {
		return nil, err
	}

	// 加密
	encrypted, err := c.encrypt(encoded, key)
	if err != nil {
		return nil, err
	}

	// Base64编码
	return encrypted, nil
}

func (c *pcrCrypto) EncryptViewerId(id uint64) (string, error) {
	// 生成随机密钥
	key, err := createKey()
	if err != nil {
		return "", err
	}
	// 加密
	encrypted, err := c.encrypt([]byte(strconv.FormatUint(id, 10)), key)
	if err != nil {
		return "", err
	}
	//	Base64编码
	b64encoded := base64.StdEncoding.EncodeToString(encrypted)
	return b64encoded, nil

}

// DecryptData 解密对象数据（先解密，再从msgpack解码）。result形参为解密结构的指针
func (c *pcrCrypto) DecryptData(encodedData string, result any) error {
	// Base64解码
	data, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return err
	}
	// 解密
	decrypted, err := c.decrypt(data)
	if err != nil {
		return err
	}
	// 去除填充
	unpaddedData, err := unpadData(decrypted)
	if err != nil {
		return err
	}
	// 解码
	err = c.decodeFromMsgpack(unpaddedData, result)
	if err != nil {
		return err
	}
	return nil
}
