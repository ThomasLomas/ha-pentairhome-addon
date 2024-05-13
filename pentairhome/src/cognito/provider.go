package cognito

import (
	"context"
	"fmt"
	"pentairhome/config"
	"time"

	cognitosrp "github.com/alexrudd/cognito-srp/v4"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	cip "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

func AuthenticateWithUsernameAndPassword(ctx context.Context, username, password string) (*types.AuthenticationResultType, error) {
	configuration := config.FetchConfiguration()

	csrp, err := cognitosrp.NewCognitoSRP(
		username,
		password,
		configuration.AWSUserPoolID,
		configuration.AWSClientID,
		nil,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create cognito srp: %s", err)
	}

	cfg, err := awsConfig.LoadDefaultConfig(
		ctx,
		awsConfig.WithRegion(config.FetchConfiguration().AWSRegion),
		awsConfig.WithCredentialsProvider(aws.AnonymousCredentials{}),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %s", err)
	}

	cipClient := cip.NewFromConfig(cfg)

	authResp, err := cipClient.InitiateAuth(ctx, &cip.InitiateAuthInput{
		AuthFlow:       types.AuthFlowTypeUserSrpAuth,
		ClientId:       aws.String(csrp.GetClientId()),
		AuthParameters: csrp.GetAuthParams(),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to initiate auth: %s", err)
	}

	if authResp.ChallengeName != types.ChallengeNameTypePasswordVerifier {
		return nil, fmt.Errorf("unexpected challenge name: %s", authResp.ChallengeName)
	}

	challengeResponses, err := csrp.PasswordVerifierChallenge(authResp.ChallengeParameters, time.Now())

	if err != nil {
		return nil, fmt.Errorf("failed to respond to password verifier challenge: %s", err)
	}

	resp, err := cipClient.RespondToAuthChallenge(ctx, &cip.RespondToAuthChallengeInput{
		ChallengeName:      types.ChallengeNameTypePasswordVerifier,
		ChallengeResponses: challengeResponses,
		ClientId:           aws.String(csrp.GetClientId()),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to respond to auth challenge: %s", err)
	}

	return resp.AuthenticationResult, nil
}
