package main

import (
	"encoding/pem"
	"github.com/insolar/x-crypto/x509"
	"github.com/pkg/errors"
	"testing"

	"github.com/insolar/insolar/api/requester"
	"github.com/insolar/insolar/platformpolicy"
	xecdsa "github.com/insolar/x-crypto/ecdsa"
	xelliptic "github.com/insolar/x-crypto/elliptic"
	xrand "github.com/insolar/x-crypto/rand"
	"github.com/stretchr/testify/require"
)

const HOST = "http://localhost:19101"
const TestUrl = HOST + "/api"

var (
	memRef  string
	memRefK string
	keys    *memberKeys
	keysK   *memberKeys
)

func TestCreateMemberP256K(t *testing.T) {
	t.Skip()
	privateKey, err := xecdsa.GenerateKey(xelliptic.P256K(), xrand.Reader)
	require.NoError(t, err)
	pk, err := exportPrivateKeyPEM(*privateKey)

	require.NoError(t, err)
	seed, err := requester.GetSeed(TestUrl)
	require.NoError(t, err)
	t.Log("seed:" + string(seed))

	datas := DataToSign{
		Reference: "1tJDL5m9pKyq2mbanYfgwQ5rSQdrpsXbzc1Dk7a53d.1tJDJLGWcX3TCXZMzZodTYWZyJGVdsajgGqyq8Vidw",
		Method:    "CreateMember",
		Seed:      seed,
		Params:    `{"name":"name"}`,
	}
	keysK = &memberKeys{Private: string(pk)}

	jws, jwk, err := createSignedData(keysK, &datas)
	require.NoError(t, err)
	params := requester.PostParams{
		"jws": jws,
		"jwk": jwk,
	}

	body, err := requester.GetResponseBody(TestUrl+"/call", params)
	require.NoError(t, err)

	t.Log(string(body))
	response, err := getResponse(body)
	require.NoError(t, err)
	require.NotNil(t, response)

	memRefK = response.Result.(string)
}

func TestGetBalanceP256K(t *testing.T) {
	t.Skip()
	seed, err := requester.GetSeed(TestUrl)
	require.NoError(t, err)
	t.Log("seed:" + string(seed))

	datas := DataToSign{
		Reference: memRefK,
		Method:    "GetBalance",
		Seed:      seed,
		Params:    `{"reference":"` + memRefK + `"}`,
	}

	jws, jwk, err := createSignedData(keysK, &datas)
	require.NoError(t, err)
	params := requester.PostParams{
		"jws": jws,
		"jwk": jwk,
	}

	body, err := requester.GetResponseBody(TestUrl+"/call", params)
	require.NoError(t, err)

	t.Log(string(body))
	response, err := getResponse(body)
	require.NoError(t, err)
	require.NotNil(t, response)
}

func TestCreateMemberP256(t *testing.T) {
	t.Skip()
	keyProcessor := platformpolicy.NewKeyProcessor()
	privateKey, err := keyProcessor.GeneratePrivateKey()
	require.NoError(t, err)
	pk, err := keyProcessor.ExportPrivateKeyPEM(privateKey)

	require.NoError(t, err)
	seed, err := requester.GetSeed(TestUrl)
	require.NoError(t, err)
	t.Log("seed:" + string(seed))

	datas := DataToSign{
		Reference: "1tJDL5m9pKyq2mbanYfgwQ5rSQdrpsXbzc1Dk7a53d.1tJDJLGWcX3TCXZMzZodTYWZyJGVdsajgGqyq8Vidw",
		Method:    "CreateMember",
		Seed:      seed,
		Params:    `{"name":"name"}`,
	}
	keys = &memberKeys{Private: string(pk)}

	jws, jwk, err := createSignedData(keys, &datas)
	require.NoError(t, err)
	params := requester.PostParams{
		"jws": jws,
		"jwk": jwk,
	}

	body, err := requester.GetResponseBody(TestUrl+"/call", params)
	require.NoError(t, err)

	t.Log(string(body))
	response, err := getResponse(body)
	require.NoError(t, err)
	require.NotNil(t, response)

	memRef = response.Result.(string)
}

func TestGetBalanceP256(t *testing.T) {
	t.Skip()
	seed, err := requester.GetSeed(TestUrl)
	require.NoError(t, err)
	t.Log("seed:" + string(seed))

	datas := DataToSign{
		Reference: memRef,
		Method:    "GetBalance",
		Seed:      seed,
		Params:    `{"reference":"` + memRef + `"}`,
	}

	jws, jwk, err := createSignedData(keys, &datas)
	require.NoError(t, err)
	params := requester.PostParams{
		"jws": jws,
		"jwk": jwk,
	}

	body, err := requester.GetResponseBody(TestUrl+"/call", params)
	require.NoError(t, err)

	t.Log(string(body))
	response, err := getResponse(body)
	require.NoError(t, err)
	require.NotNil(t, response)
}

func exportPrivateKeyPEM(privateKey xecdsa.PrivateKey) ([]byte, error) {
	x509Encoded, err := x509.MarshalECPrivateKey(&privateKey)
	if err != nil {
		return nil, errors.Wrap(err, "[ ExportPrivateKey ]")
	}
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})
	return pemEncoded, nil
}