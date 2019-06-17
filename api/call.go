//
// Copyright 2019 Insolar Technologies GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package api

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/insolar/insolar/metrics"

	"github.com/insolar/insolar/api/requester"

	"github.com/pkg/errors"

	"github.com/insolar/insolar/api/seedmanager"
	"github.com/insolar/insolar/application/extractor"
	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/reply"
	"github.com/insolar/insolar/insolar/utils"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/instrumentation/instracer"
)

const (
	TimeoutError = -215
	ResultError  = -217
)

// UnmarshalRequest unmarshals request to api
func UnmarshalRequest(req *http.Request, params interface{}) ([]byte, error) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, errors.Wrap(err, "[ UnmarshalRequest ] Can't read body. So strange")
	}
	if len(body) == 0 {
		return nil, errors.New("[ UnmarshalRequest ] Empty body")
	}

	err = json.Unmarshal(body, &params)
	if err != nil {
		return body, errors.Wrap(err, "[ UnmarshalRequest ] Can't unmarshal input params")
	}
	return body, nil
}

func (ar *Runner) checkSeed(paramsSeed string) error {
	decoded, err := base64.StdEncoding.DecodeString(paramsSeed)
	if err != nil {
		return errors.New("[ checkSeed ] Decode error")
	}
	seed := seedmanager.SeedFromBytes(decoded)
	if seed == nil {
		return errors.New("[ checkSeed ] Bad seed param")
	}

	if !ar.SeedManager.Exists(*seed) {
		return errors.New("[ checkSeed ] Incorrect seed")
	}

	return nil
}

func (ar *Runner) makeCall(ctx context.Context, request requester.Request, rawBody []byte, signature string, pulseTimeStamp int64) (interface{}, error) {
	ctx, span := instracer.StartSpan(ctx, "SendRequest "+request.Method)
	defer span.End()

	reference, err := insolar.NewReferenceFromBase58(request.Params.Reference)
	if err != nil {
		return nil, errors.Wrap(err, "[ makeCall ] failed to parse params.Reference")
	}

	requestArgs, err := insolar.MarshalArgs(rawBody, signature, pulseTimeStamp)
	if err != nil {
		return nil, errors.Wrap(err, "[ makeCall ] failed to marshal arguments")
	}

	res, err := ar.ContractRequester.SendRequest(
		ctx,
		reference,
		"Call",
		[]interface{}{*ar.CertificateManager.GetCertificate().GetRootDomainReference(), requestArgs},
	)

	if err != nil {
		return nil, errors.Wrap(err, "[ makeCall ] Can't send request")
	}

	result, contractErr, err := extractor.CallResponse(res.(*reply.CallMethod).Result)

	if err != nil {
		return nil, errors.Wrap(err, "[ makeCall ] Can't extract response")
	}

	if contractErr != nil {
		return nil, errors.Wrap(errors.New(contractErr.S), "[ makeCall ] Error in called method")
	}

	return result, nil
}

func processError(err error, extraMsg string, resp *requester.ContractAnswer, insLog insolar.Logger, traceID string) {
	errResponse := &requester.Error{Message: "Messagebus timeout exceeded", Code: ResultError, TraceID: traceID}
	resp.Error = errResponse
	insLog.Error(errors.Wrapf(err, "[ CallHandler ] %s", extraMsg))
}

func (ar *Runner) callHandler() func(http.ResponseWriter, *http.Request) {
	return func(response http.ResponseWriter, req *http.Request) {
		traceID := utils.RandTraceID()
		ctx, insLog := inslogger.WithTraceField(context.Background(), traceID)

		ctx, span := instracer.StartSpan(ctx, "callHandler")
		defer span.End()

		contractRequest := requester.Request{}
		contractAnswer := requester.ContractAnswer{}

		startTime := time.Now()
		defer func() {
			success := "success"
			if contractAnswer.Error != nil {
				success = "fail"
			}
			metrics.APIContractExecutionTime.WithLabelValues(contractRequest.Method, success).Observe(time.Since(startTime).Seconds())
		}()

		insLog.Infof("[ callHandler ] Incoming contractRequest: %s", req.RequestURI)

		defer func() {
			res, err := json.MarshalIndent(contractAnswer, "", "    ")
			if err != nil {
				res = []byte(`{"error": "can't marshal ContractAnswer to json'"}`)
			}
			response.Header().Add("Content-Type", "application/json")
			_, err = response.Write(res)
			if err != nil {
				insLog.Errorf("Can't write response\n")
			}
		}()

		rawBody, err := UnmarshalRequest(req, &contractRequest)
		if err != nil {
			processError(err, "Can't unmarshal contractRequest", &contractAnswer, insLog, traceID)
			return
		}

		contractAnswer.JSONRPC = contractRequest.JSONRPC
		contractAnswer.ID = contractRequest.ID

		digest := req.Header.Get(requester.Digest)
		richSignature := req.Header.Get(requester.Signature)

		signature, err := validateRequestHeaders(digest, richSignature, rawBody)
		if err != nil {
			processError(err, "Can't validate contractRequest", &contractAnswer, insLog, traceID)
			return
		}

		if len(contractRequest.LogLevel) > 0 {
			logLevelNumber, err := insolar.ParseLevel(contractRequest.LogLevel)
			if err != nil {
				processError(err, "Can't parse logLevel", &contractAnswer, insLog, traceID)
				return
			}
			ctx = inslogger.WithLoggerLevel(ctx, logLevelNumber)
		}

		err = ar.checkSeed(contractRequest.Params.Seed)
		if err != nil {
			processError(err, "Can't checkSeed", &contractAnswer, insLog, traceID)
			return
		}

		// pulse, err := ar.PulseAccessor.Latest(ctx)
		// if err != nil {
		// 	processError(err, "Can't get last pulse", &contractAnswer, insLog, traceID)
		// 	return
		// }

		var result interface{}
		ch := make(chan interface{}, 1)
		go func() {
			// result, err = ar.makeCall(ctx, contractRequest, rawBody, signature, pulse.PulseTimestamp)
			result, err = ar.makeCall(ctx, contractRequest, rawBody, signature, 0)
			ch <- nil
		}()
		select {

		case <-ch:
			if err != nil {
				processError(err, "Can't makeCall", &contractAnswer, insLog, traceID)
				return
			}
			contractResult := &requester.Result{ContractResult: result, TraceID: traceID}
			contractAnswer.Result = contractResult
			return

		case <-time.After(time.Duration(ar.cfg.Timeout) * time.Second):
			errResponse := &requester.Error{Message: "Messagebus timeout exceeded", Code: TimeoutError, TraceID: traceID}
			contractAnswer.Error = errResponse
			return
		}
	}
}

func validateRequestHeaders(digest string, richSignature string, body []byte) (string, error) {
	// Digest = "SHA-256=<hashString>"
	// Signature = "keyId="member-pub-key", algorithm="ecdsa", headers="digest", signature=<signatureString>"
	if len(digest) == 0 || len(richSignature) == 0 || len(body) == 0 {
		return "", errors.Errorf("Invalid input data length digest: %d, signature: %d, body: %d", len(digest), len(richSignature), len(body))
	}
	h := sha256.New()
	_, err := h.Write(body)
	if err != nil {
		return "", errors.Wrap(err, "Cant get hash")
	}
	sha := base64.URLEncoding.EncodeToString(h.Sum(nil))
	if sha == digest[strings.IndexByte(digest, '=')+1:] {
		sig := richSignature[strings.Index(richSignature, "signature=")+10:]
		if len(sig) == 0 {
			return "", errors.New("Empty signature")
		}
		return sig, nil

	}
	return "", errors.New("Cant get signature from header")
}
