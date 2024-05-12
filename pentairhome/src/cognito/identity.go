package cognito

import (
	"context"
	"pentairhome/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	ci "github.com/aws/aws-sdk-go-v2/service/cognitoidentity"
	cit "github.com/aws/aws-sdk-go-v2/service/cognitoidentity/types"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

func GetCredentialsFromAuthentication(ctx context.Context, authenticationResult *types.AuthenticationResultType) *cit.Credentials {
	appConfiguration := config.FetchConfiguration()

	cfg, _ := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(appConfiguration.AWSRegion),
	)
	cognitoIdentityService := ci.NewFromConfig(cfg)

	logins := map[string]string{
		appConfiguration.GetLoginKey(): *authenticationResult.IdToken,
	}

	idRes, err := cognitoIdentityService.GetId(ctx, &ci.GetIdInput{
		IdentityPoolId: aws.String(appConfiguration.AWSIdentityPoolId),
		Logins:         logins,
	})

	if err != nil {
		panic(err)
	}

	credsRes, err := cognitoIdentityService.GetCredentialsForIdentity(ctx, &ci.GetCredentialsForIdentityInput{
		IdentityId: idRes.IdentityId,
		Logins:     logins,
	})

	if err != nil {
		panic(err)
	}

	return credsRes.Credentials
}
