<!--
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership. The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License. You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied. See the License for the
specific language governing permissions and limitations
under the License.
-->

# Personal access tokens

A personal access token (PAT) lets an HTTP client act as an existing Answer
user. The token's scopes limit the operations it can request. The user's
current permissions, reputation, account status, visibility rules, rate
limits, CAPTCHA requirements, and moderation policies still apply.

Effective authority is the intersection of token scopes, current account
permissions, and instance policy. Administrator-managed instance API keys
remain a separate credential type.

## Enable and create

1. As an administrator, open **Admin → Security** and enable **Personal access
   tokens**. The feature is disabled by default.
2. Set the reauthentication window if needed: the default is 60 minutes and
   the supported range is 5–120 minutes.
3. As the token owner, open **Settings → Personal access tokens**.
4. Choose **Create token**, enter a name, explicitly select scopes, and choose
   an expiration. Scopes are not preselected and do not imply one another.
5. Confirm the current password if recent authentication is required. Users
   authenticating through an external identity provider can sign out and sign
   in again through that provider.
6. Copy the secret from its one-time display. The complete secret cannot be
   retrieved later; the server stores its SHA-256 digest and a display suffix.

Expiration is mandatory and must be no more than 365 days in the future.

| Scope | Permits |
| --- | --- |
| `question.read` | Question reads and related tag discovery |
| `question.create` | Question creation and supporting tag discovery |
| `answer.read` | Answer reads |
| `answer.create` | Answer creation |
| `vote.write` | Upvotes, downvotes, and vote retraction |

Search requires both `question.read` and `answer.read`. A create scope does
not include the corresponding read scope. Routes outside the PAT allowlist
are denied even when the owner is an administrator.

## Use an existing REST endpoint

The supported transport is `Authorization: Bearer answer_pat_…`. Use your
instance's HTTPS URL and read the secret into a variable:

```bash
export ANSWER_URL=https://answer.example.com
read -rsp 'Personal access token: ' ANSWER_PAT
echo

curl --fail-with-body --silent --show-error \
  -H "Authorization: Bearer $ANSWER_PAT" \
  "$ANSWER_URL/answer/api/v1/personal-access-tokens/current"
```

Self-inspection returns the owner and token metadata, including scopes and
expiry. It does not return the complete secret. A bare token header or a
token supplied through the `Authorization` query parameter is rejected.

For the following examples, select the matching create/read/vote scopes.
Choose an existing tag on the instance in place of `support`.

```bash
curl --fail-with-body --silent --show-error \
  -H "Authorization: Bearer $ANSWER_PAT" \
  -H 'Content-Type: application/json' \
  -d '{"title":"How should we organize shared project knowledge?","content":"Which practices work well for maintaining useful questions and answers?","tags":[{"slug_name":"support","display_name":"support"}]}' \
  "$ANSWER_URL/answer/api/v1/question"
```

Use the returned question ID to post an answer:

```bash
export QUESTION_ID='<question-id>'
curl --fail-with-body --silent --show-error \
  -H "Authorization: Bearer $ANSWER_PAT" \
  -H 'Content-Type: application/json' \
  -d "{\"question_id\":\"$QUESTION_ID\",\"content\":\"Search before posting, write focused questions, and keep accepted answers current.\"}" \
  "$ANSWER_URL/answer/api/v1/answer"
```

For voting, choose an object owned by a different user. The owner must have
the required account permission and reputation:

```bash
export OBJECT_ID='<another-user-question-or-answer-id>'
curl --fail-with-body --silent --show-error \
  -H "Authorization: Bearer $ANSWER_PAT" \
  -H 'Content-Type: application/json' \
  -d "{\"object_id\":\"$OBJECT_ID\",\"is_cancel\":false}" \
  "$ANSWER_URL/answer/api/v1/vote/up"
```

Use `is_cancel: true` to retract that vote. Use `/vote/down` for a downvote.
Content can enter the normal review flow, and requests can require CAPTCHA.

## Revoke or disable

Owners can list and revoke their own tokens in account settings. Revocation
is immediate and permanent. Expiry is also terminal. Disabling PAT support
blocks existing PAT authentication and new token creation; owners can still
revoke tokens. Re-enabling support resumes access only for tokens that remain
unexpired and unrevoked.

Deleting an account or recovering its password revokes its PATs. The normal
password-change flow also offers a choice to revoke PATs. Current suspension,
verification, role, and external-user status continue to constrain access.

Token management requires a browser login session; a PAT cannot create, list,
or revoke tokens. The management endpoints are:

| Method | Path | Credential |
| --- | --- | --- |
| POST | `/answer/api/v1/user/reauthenticate` | Browser session |
| GET / POST / DELETE | `/answer/api/v1/personal-access-tokens` | Browser session |
| GET | `/answer/api/v1/personal-access-tokens/current` | PAT |

OpenAPI describes request and response fields. Missing scope returns HTTP
403 with reason `error.personal_access_token.insufficient_scope`; disabled
access returns HTTP 403 with reason `error.feature.disabled`. Invalid,
expired, or revoked credentials return HTTP 401. Account and content policy
errors retain their existing server reasons and status codes.
