package cognito

import (
	"context"
	"pentairhome/config"
	"time"

	cognitosrp "github.com/alexrudd/cognito-srp/v4"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	cip "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

func AuthenticateWithUsernameAndPassword(ctx context.Context, username, password string) *types.AuthenticationResultType {
	configuration := config.FetchConfiguration()

	csrp, _ := cognitosrp.NewCognitoSRP(
		username,
		password,
		configuration.AWSUserPoolID,
		configuration.AWSClientID,
		nil,
	)

	cfg, _ := awsConfig.LoadDefaultConfig(
		ctx,
		awsConfig.WithRegion(config.FetchConfiguration().AWSRegion),
		awsConfig.WithCredentialsProvider(aws.AnonymousCredentials{}),
	)

	cipClient := cip.NewFromConfig(cfg)

	authResp, err := cipClient.InitiateAuth(ctx, &cip.InitiateAuthInput{
		AuthFlow:       types.AuthFlowTypeUserSrpAuth,
		ClientId:       aws.String(csrp.GetClientId()),
		AuthParameters: csrp.GetAuthParams(),
	})

	if err != nil {
		panic(err)
	}

	if authResp.ChallengeName != types.ChallengeNameTypePasswordVerifier {
		panic("unexpected challenge")
	}

	challengeResponses, _ := csrp.PasswordVerifierChallenge(authResp.ChallengeParameters, time.Now())

	resp, err := cipClient.RespondToAuthChallenge(ctx, &cip.RespondToAuthChallengeInput{
		ChallengeName:      types.ChallengeNameTypePasswordVerifier,
		ChallengeResponses: challengeResponses,
		ClientId:           aws.String(csrp.GetClientId()),
	})

	if err != nil {
		panic(err)
	}

	return resp.AuthenticationResult
}
