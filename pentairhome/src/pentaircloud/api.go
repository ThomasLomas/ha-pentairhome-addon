package pentaircloud

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
func (client APIClient) MakeRequest(endpoint, method string, body io.Reader) ([]byte, error) {
	url := fmt.Sprintf("%s%s", "https://api.pentair.cloud/", endpoint)
	req, err := http.NewRequest(method, url, body)

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %s", err)
	}

	awscred, err := client.CredsCache.Retrieve(client.Context)

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credentials: %s", err)
	}

	req.Header.Set("x-amz-id-token", *client.IDToken)
	req.Header.Set("user-agent", "aws-amplify/4.3.10 react-native")
	req.Header.Set("content-type", "application/json; charset=UTF-8")

	signer := v4.NewSigner()
	contentHash, contentHashErr := getPayloadHash(req)

	if contentHashErr != nil {
		return nil, fmt.Errorf("failed to get payload hash: %s", contentHashErr)
	}

	if err := signer.SignHTTP(client.Context, awscred, req, contentHash, "execute-api", *client.AWSRegion, time.Now()); err != nil {
		return nil, fmt.Errorf("failed to sign request: %s", err)
	}

	httpResp, httpErr := client.HttpClient.Do(req)

	if httpErr != nil {
		return nil, fmt.Errorf("failed to make request: %s", httpErr)
	}

	bodyBytes, readErr := io.ReadAll(httpResp.Body)

	if readErr != nil {
		return nil, fmt.Errorf("failed to read response body: %s", readErr)
	}

	httpResp.Body.Close()

	return bodyBytes, nil
}

func getPayloadHash(request *http.Request) (string, error) {
	if request.Body == nil {
		return "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", nil
	}

	bodyBytes, readErr := io.ReadAll(request.Body)

	if readErr != nil {
		return "", fmt.Errorf("failed to read request body: %s", readErr)
	}

	request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	sha256Hash := sha256.Sum256(bodyBytes)
	return hex.EncodeToString(sha256Hash[:]), nil
}
