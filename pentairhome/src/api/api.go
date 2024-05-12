package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
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
	HttpClient   *http.Client
	Context      context.Context
	IDToken      *string
	AccessKeyId  *string
	AWSRegion    *string
	SecretKey    *string
	SessionToken *string
	CredsCache   *aws.CredentialsCache
}

func NewAPIClient(ctx context.Context, idToken, accessKey, secretKey, sessionToken string) *APIClient {
	config := config.FetchConfiguration()

	credsCache := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(
		accessKey,
		secretKey,
		sessionToken,
	))

	return &APIClient{
		HttpClient:   new(http.Client),
		Context:      ctx,
		IDToken:      &idToken,
		AccessKeyId:  &accessKey,
		AWSRegion:    &config.AWSRegion,
		SecretKey:    &secretKey,
		SessionToken: &sessionToken,
		CredsCache:   credsCache,
	}
}

type RequestOptions struct {
	MaxRetries int
	RetryCount int
}

func (client APIClient) MakeRequest(endpoint, method string, body io.Reader, options ...RequestOptions) []byte {
	url := fmt.Sprintf("%s%s", urlBase, endpoint)
	req, err := http.NewRequest(method, url, body)

	if err != nil {
		log.Fatalf("failed to create request: %s", err)
	}

	awscred, err := client.CredsCache.Retrieve(client.Context)

	if err != nil {
		log.Fatalf("failed to retrieve credentials: %s", err)
	}

	req.Header.Set("x-amz-id-token", *client.IDToken)
	req.Header.Set("user-agent", "aws-amplify/4.3.10 react-native")
	req.Header.Set("content-type", "application/json; charset=UTF-8")

	signer := v4.NewSigner()
	contentHash := getPayloadHash(req)

	if err := signer.SignHTTP(client.Context, awscred, req, contentHash, "execute-api", *client.AWSRegion, time.Now()); err != nil {
		log.Fatalf("failed to sign request: %s", err)
	}

	httpResp, httpErr := client.HttpClient.Do(req)

	if httpErr != nil {
		log.Fatalf("failed to make request: %s", httpErr)
	}

	bodyBytes, readErr := io.ReadAll(httpResp.Body)

	if readErr != nil {
		log.Fatalf("failed to read response body: %s", readErr)
	}

	httpResp.Body.Close()

	return bodyBytes
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
