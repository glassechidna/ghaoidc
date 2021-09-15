# ghaoidc
## Assumes roles in AWS that have useful role session tags

GitHub Actions has [(almost) launched OpenID Connect][launch] federation. This
means you can assume a role in AWS without needing to store long-lived credentials
as secrets in your GitHub repository. This is really great, but it could be 
**even better**. The JWT issued by GHA contains lots of useful claims, but STS
`AssumeRoleWithWebIdentity` can't use most of them. Hence this project.

This project is two parts: First, an API Gateway with JWT auth (with GHA as the issuer)
and a Lambda function behind it that assumes roles **using those claims as role
[session tags][session-tags]**. Second, a GHA "action" that acts as the client
of that API and requests credentials to be used in a workflow.

## Usage

Deploy the API Gateway and Lambda using the [`api.yml`](/api.yml) CloudFormation
template. I recommend creating a brand new AWS account solely for this purpose. 
The template has some inline documentation. **TODO: Include build instructions 
and an AWS SAR application once I'm happy I don't want to make massive changes**.

Next, create roles that your GHA workflows will be assuming. Look at 
[`example.yml`](/example.yml) for guidance on a trust policy.

Finally, include the GHA action in your workflows. A "hello world" example
looks like: 

```yaml
name: Example
on:
  push:

jobs:
  build:
    runs-on: ubuntu-latest
    permissions:
      id-token: write
      contents: read
    steps:
      - uses: actions/checkout@v2
      
        # this actions sets AWS_* environment variables for later steps
      - uses: glassechidna/ghaoidc@main
        with:
          apiUrl: ${{ secrets.CREDENTIALS_URL }}
          roleArn: arn:aws:iam::0123456789012:role/DeploymentRole
        
      - run: aws sts get-caller-identity --region us-east-1
```

There's also an optional `transitiveTags` input parameter. It's worth noting 
that you can choose to make the either `apiUrl`, `roleArn`, both or neither
values secret. It's just a tradeoff on usability. 

## How?

**TODO: flesh out explanation.**

![architecture diagram](/docs/diagram.png)

## Why?

**TODO: flesh out spiel.** Two main reasons: you get highly enriched entries
in CloudTrail and can trace actions back to specific GHA jobs with ease. Also,
you can use the session tags as variables in your IAM policies. This allows you
to parameterise your deployment IAM roles, giving you highly granular and isolated
permissions without needing a 1:1 mapping of repositories to roles.

## Caveats

While building this I learned that AWS STS enforces a limit of approximately
500 bytes on the combined total of role session tag keys and values. Which the
complete GHA OIDC JWT exceeds. You can choose which values to pass through as
tags in the CFN template - I've selected what I think are useful defaults.

[launch]: https://awsteele.com/blog/2021/09/15/aws-federation-comes-to-github-actions.html
[session-tags]: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_session-tags.html
