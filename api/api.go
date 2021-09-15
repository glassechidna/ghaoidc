package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"os"
	"regexp"
	"time"
)

func main() {
	lambda.Start(handle)
}

func handle(ctx context.Context, input *events.APIGatewayV2HTTPRequest) (interface{}, error) {
	sess := session.Must(session.NewSession())
	api := sts.New(sess)

	roleArn := input.Headers["ghaoidc-role-arn"]
	sessionName := fmt.Sprintf("%d", time.Now().Unix())
	keyPrefix := os.Getenv("TAG_KEY_PREFIX")

	claims := input.RequestContext.Authorizer.JWT.Claims
	if claims["repository_owner"] != os.Getenv("PERMITTED_GITHUB_OWNER") {
		return nil, errors.New("disallowed")
	}

	assume, err := api.AssumeRole(&sts.AssumeRoleInput{
		RoleArn:           aws.String(roleArn),
		RoleSessionName:   aws.String(sessionName),
		Tags:              getTags(claims, keyPrefix),
		TransitiveTagKeys: transitiveTags(input.Headers["ghaoidc-transitive-tags"], keyPrefix),
	})
	if err != nil {
		return nil, err
	}

	return assume, nil
}

func getTags(claims map[string]string, prefix string) []*sts.Tag {
	allowListSlice := regexp.MustCompile(`\s+`).Split(os.Getenv("CLAIMS_ALLOW_LIST"), -1)
	allowMap := map[string]struct{}{}
	for _, claim := range allowListSlice {
		allowMap[claim] = struct{}{}
	}

	tags := []*sts.Tag{}

	for key, val := range claims {
		if _, allowed := allowMap[key]; allowed {
			tags = append(tags, &sts.Tag{
				Key:   aws.String(prefix + key),
				Value: aws.String(val),
			})
		}
	}

	return tags
}

func transitiveTags(header, prefix string) []*string {
	re := regexp.MustCompile(`\s+`)
	keys := re.Split(header, -1)

	if len(keys) == 1 && keys[0] == "" {
		return nil
	}

	ptrs := []*string{}
	for _, key := range keys {
		ptrs = append(ptrs, aws.String(prefix+key))
	}

	return ptrs
}
