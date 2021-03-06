package jwt

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/imdario/mergo"
)

// SignOption represents the options of Sign.
type SignOption struct {
	Algorithm Algorithm
	ExpiresIn time.Duration
	Audience  string
	Issuer    string
	Subject   string
	// Header is the customized header which will be merged to token's header.
	Header Header
}

// Sign signs the given payload and serect to the JSON web token,
// when using HMAC algorithm, secretOrPrivateKey's type should be string or []
// byte , when using RSA algorithm, secretOrPrivateKey's type should be
// rsa.PrivateKey. If the opt given is nil, it will use the defualt HS256
// algorithm.
func Sign(payload Payload, secretOrPrivateKey interface{}, opt *SignOption) (token []byte, err error) {
	if payload == nil {
		return nil, ErrEmptyPayload
	}

	if opt == nil {
		opt = &SignOption{}
	}

	if secretOrPrivateKey == nil {
		return nil, ErrEmptySecretOrPrivateKey
	}

	var headerJSON, payloadJSON, signature []byte

	if headerJSON, err = marshalHeader(opt); err != nil {
		return
	}

	hBase64 := []byte(base64.StdEncoding.EncodeToString(headerJSON))

	if payloadJSON, err = marshalPayload(payload, opt); err != nil {
		return
	}

	pBase64 := []byte(base64.StdEncoding.EncodeToString(payloadJSON))

	if opt.Algorithm == "" {
		opt.Algorithm = HS256
	}

	algImp, ok := algImpMap[opt.Algorithm]

	if !ok {
		return nil, ErrInvalidAlgorithm
	}

	if signature, err = algImp.sign(bytes.Join([][]byte{hBase64, pBase64},
		periodBytes), secretOrPrivateKey); err != nil {
		return
	}

	sigBase64 := []byte(base64.StdEncoding.EncodeToString(signature))

	return bytes.Join([][]byte{hBase64, pBase64, sigBase64}, periodBytes), nil
}

func marshalHeader(opt *SignOption) ([]byte, error) {
	h := map[string]interface{}{
		"alg": opt.Algorithm,
		"typ": "JWT",
	}

	if opt.Header != nil {
		if err := mergo.Map(&h, opt.Header); err != nil {
			return nil, err
		}
	}

	return json.Marshal(h)
}

func marshalPayload(payload Payload, opt *SignOption) ([]byte, error) {
	claims := Payload{"iat": time.Now().Unix()}

	if opt.Issuer != "" {
		claims["iss"] = opt.Issuer
	}
	if opt.ExpiresIn != 0 {
		claims["exp"] = opt.ExpiresIn / 1e9
	}
	if opt.Subject != "" {
		claims["sub"] = opt.Subject
	}
	if opt.Audience != "" {
		claims["aud"] = opt.Audience
	}

	if err := mergo.Map(&claims, payload); err != nil {
		return nil, err
	}

	return json.Marshal(claims)
}
