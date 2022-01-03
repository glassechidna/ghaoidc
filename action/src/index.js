import core from "@actions/core";
import fetch from "node-fetch";

async function getJwt() {
    const { ACTIONS_ID_TOKEN_REQUEST_TOKEN, ACTIONS_ID_TOKEN_REQUEST_URL } = process.env;
    const resp = await fetch(`${ACTIONS_ID_TOKEN_REQUEST_URL}`, {
        headers: {Authorization: `bearer ${ACTIONS_ID_TOKEN_REQUEST_TOKEN}`}
    });

    const { value } = await resp.json();
    return value;
}

(async() => {
    const apiUrl = core.getInput("apiUrl");
    const roleArn = core.getInput("roleArn", {required: true});
    const transitiveTags = core.getInput("transitiveTags");
    const jwt = await getJwt();

    const resp = await fetch(apiUrl, {
        headers: {
            "ghaoidc-role-arn": roleArn,
            "ghaoidc-transitive-tags": transitiveTags,
            "authorization": jwt,
        },
    });

    const body = await resp.json();
    const { AccessKeyId, SecretAccessKey, SessionToken } = body.Credentials;

    core.setSecret(SecretAccessKey);
    core.setSecret(SessionToken);

    core.exportVariable("AWS_ACCESS_KEY_ID", AccessKeyId);
    core.exportVariable("AWS_SECRET_ACCESS_KEY", SecretAccessKey);
    core.exportVariable("AWS_SESSION_TOKEN", SessionToken);
})();
