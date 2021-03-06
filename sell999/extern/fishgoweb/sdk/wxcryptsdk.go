package sdk

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math/rand"

	. "github.com/milkbobo/fishgoweb/language"
)

type WxCryptSdk struct {
	AESKey string
	Token  string
	AppId  string
}

func (this *WxCryptSdk) getSignature(token string, timestamp string, nonce string, msg string) string {
	arrayInfo := []string{token, timestamp, nonce, msg}
	arrayInfo = ArraySort(arrayInfo).([]string)
	arrayInfoString := Implode(arrayInfo, "")
	return this.encodeSha1(arrayInfoString)
}

func (this *WxCryptSdk) getRandomStr(length int) []byte {
	chars := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz")
	result := make([]byte, length, length)
	for i := 0; i < length; i++ {
		key := rand.Intn(len(chars))
		result[i] = chars[key]
	}
	return result
}

func (this *WxCryptSdk) encodeSha1(data string) string {
	t := sha1.New()
	io.WriteString(t, data)
	return fmt.Sprintf("%x", t.Sum(nil))
}

func (this *WxCryptSdk) decodeXml(msg []byte, data interface{}) error {
	return xml.Unmarshal(msg, data)
}

func (this *WxCryptSdk) encodeXml(encrypt string, signature string, timestamp string, nonce string) ([]byte, error) {
	return []byte(fmt.Sprintf(`<xml>
		<Encrypt><![CDATA[%s]]></Encrypt>
		<MsgSignature><![CDATA[%s]]></MsgSignature>
		<TimeStamp>%s</TimeStamp>
		<Nonce><![CDATA[%s]]></Nonce>
		</xml>`, encrypt, signature, timestamp, nonce)), nil
}

func (this *WxCryptSdk) pkcs7Unpadding(data []byte, blockSize int) []byte {
	length := len(data)
	unPadding := int(data[length-1])
	return data[:(length - unPadding)]
}

func (this *WxCryptSdk) pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

func (this *WxCryptSdk) decryptAES(AESKey string, msg string) ([]byte, error) {
	aesKey, err := base64.StdEncoding.DecodeString(AESKey + "=")
	if err != nil {
		return nil, err
	}
	cipherText, err := base64.StdEncoding.DecodeString(msg)
	if err != nil {
		return nil, err
	}
	iv := aesKey[0:16]
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(cipherText, cipherText)
	cipherText = this.pkcs7Unpadding(cipherText, block.BlockSize())
	return cipherText, nil
}

func (this *WxCryptSdk) encryptAES(AESKey string, msg []byte) (string, error) {
	aesKey, err := base64.StdEncoding.DecodeString(AESKey + "=")
	if err != nil {
		return "", err
	}
	iv := aesKey[0:16]
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	cipherText := this.pkcs7Padding([]byte(msg), block.BlockSize())
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText, cipherText)
	cipherTextEncode := base64.StdEncoding.EncodeToString(cipherText)
	return cipherTextEncode, nil
}

func (this *WxCryptSdk) decodeMeta(packaget []byte) ([]byte, string) {
	//????????????????????????
	packaget = packaget[16:]
	//????????????
	msgLen := binary.BigEndian.Uint32(packaget[0:4])
	packaget = packaget[4:]
	//??????
	msg := packaget[0:msgLen]
	packaget = packaget[msgLen:]
	//appId
	appId := packaget
	return msg, string(appId)
}

func (this *WxCryptSdk) encodeMeta(packaget []byte, appId string) []byte {
	var buffer bytes.Buffer
	//????????????????????????
	buffer.Write(this.getRandomStr(16))
	//????????????
	lengthBuffer := make([]byte, 4, 4)
	binary.BigEndian.PutUint32(lengthBuffer, uint32(len(packaget)))
	buffer.Write(lengthBuffer)
	//??????
	buffer.Write(packaget)
	//appId
	buffer.WriteString(appId)
	return buffer.Bytes()
}

func (this *WxCryptSdk) Decrypt(msgSignature string, timestamp string, nonce string, msg []byte) (string, []byte, error) {
	//????????????xml
	var encryptMessage struct {
		ToUserName string `xml:"ToUserName"`
		Encrypt    string `xml:"Encrypt"`
	}
	err := this.decodeXml(msg, &encryptMessage)
	if err != nil {
		return "", nil, err
	}
	//????????????
	realSignature := this.getSignature(
		this.Token,
		timestamp,
		nonce,
		encryptMessage.Encrypt)
	if realSignature != msgSignature {
		return "", nil, errors.New("??????????????????")
	}
	//????????????xml
	packaget, err := this.decryptAES(
		this.AESKey,
		encryptMessage.Encrypt,
	)
	if err != nil {
		return "", nil, err
	}
	packaget, appId := this.decodeMeta(packaget)
	if appId != this.AppId {
		return "", nil, errors.New("??????appId????????????")
	}
	return encryptMessage.ToUserName, packaget, nil
}

func (this *WxCryptSdk) Encrypt(timestamp string, nonce string, msg []byte) ([]byte, error) {
	//????????????xml
	msgWithMeta := this.encodeMeta(
		msg,
		this.AppId)
	encodeMsg, err := this.encryptAES(
		this.AESKey,
		msgWithMeta)
	if err != nil {
		return nil, err
	}
	//????????????
	signature := this.getSignature(
		this.Token,
		timestamp,
		nonce,
		encodeMsg)
	//????????????xml
	return this.encodeXml(encodeMsg, signature, timestamp, nonce)
}
