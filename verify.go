package jwt

import (
	"time"
)

type VerifyOption struct {
	Algorithm        Algorithm
	IngoreExpiration bool
	Audience         string
	Subject          string
	clockTolerance   time.Duration
}

// Verify decodes the given token and check whether the token is valid.
func Verify(token []byte, secretOrPrivateKey interface{}, opt *VerifyOption) (map[string]interface{}, map[string]interface{}, error) {
	header, payload, err := decode(string(token))

	if err != nil {
		return nil, nil, err
	}

	var (
		ok     bool
		typ    interface{}
		typStr string
		algImp algorithmImplementation
	)

	if typ, ok = header["typ"]; !ok {
		return nil, nil, ErrInvalidHeaderType
	}

	if typStr, ok = typ.(string); !ok {
		return nil, nil, ErrInvalidHeaderType
	}

	if typStr != "JWT" {
		return nil, nil, ErrInvalidHeaderType
	}

	if algImp, ok = algImpMap[opt.Algorithm]; !ok {
		return nil, nil, ErrInvalidAlgorithm
	}

	if err := algImp.verify(token, secretOrPrivateKey); err != nil {
		return nil, nil, err
	}

	return header, payload, nil
}