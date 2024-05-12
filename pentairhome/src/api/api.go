package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"pentairhome/config"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	credentials "github.com/aws/aws-sdk-go-v2/credentials"
)

const urlBase = "https://api.pentair.cloud/"
const defaultEmptyPayloadHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

type APIClient struct {
	Context      context.Context
	IDToken      *string
	AccessKeyId  *string
	SecretKey    *string
	SessionToken *string
}

func (client APIClient) MakeRequest(endpoint string, method string, body io.Reader) []byte {
	config := config.FetchConfiguration()
	httpClient := new(http.Client)

	url := fmt.Sprintf("%s%s", urlBase, endpoint)
	req, _ := http.NewRequest(method, url, body)

	creds := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(
		*client.AccessKeyId,
		*client.SecretKey,
		*client.SessionToken,
	))

	awscred, err := creds.Retrieve(client.Context)

	if err != nil {
		panic(err)
	}

	req.Header.Set("x-amz-id-token", *client.IDToken)
	req.Header.Set("user-agent", "aws-amplify/4.3.10 react-native")
	req.Header.Set("content-type", "application/json; charset=UTF-8")

	signer := v4.NewSigner()
	contentHash := getPayloadHash(req)
	if err := signer.SignHTTP(client.Context, awscred, req, contentHash, "execute-api", config.AWSRegion, time.Now()); err != nil {
		panic(fmt.Sprintf("aws signer: failed to sign request: %s", err))
	}
	resp, _ := httpClient.Do(req)
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	return b
}

func getPayloadHash(request *http.Request) string {
	if request.Body == nil {
		return defaultEmptyPayloadHash
	}

	bodyBytes, _ := io.ReadAll(request.Body)
	request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	sha256Hash := sha256.Sum256(bodyBytes)
	return hex.EncodeToString(sha256Hash[:])
}
