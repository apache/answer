<!--
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements. See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0.
-->

# answer-cli commands

All commands write JSON to stdout. Parse `ok`; do not infer success from prose.

## Authentication

```bash
answer-cli auth login --server https://answer.example.com --with-token < token.txt
answer-cli auth status
answer-cli auth logout
```

`auth logout` removes the local credential but does not revoke the server PAT.

## Read

```bash
answer-cli question search --query "keywords"
answer-cli question get QUESTION_ID
answer-cli answer list --question QUESTION_ID
answer-cli answer get ANSWER_ID
answer-cli tag search --query "tag"
```

## Write

```bash
answer-cli question create \
  --title "Question title" \
  --tag tag-slug \
  --body-file question.md

answer-cli answer create \
  --question QUESTION_ID \
  --body-file answer.md

answer-cli vote up OBJECT_ID
answer-cli vote down OBJECT_ID
answer-cli vote retract OBJECT_ID --direction up
```

Use `--body-file -` to read Markdown from stdin. Use `--input-json -` when a complete structured request is easier.

## Errors

- `invalid_token`: ask the user to configure a current PAT.
- `agent_access_disabled`: the instance administrator has disabled PAT use.
- `insufficient_token_scope`: report the required scope; do not seek another credential automatically.
- `permission_denied`: the owner lacks the required Answer permission or reputation.
- `captcha_required`: stop; ask the user to wait or complete the action in the web UI.
- `outcome_unknown`: inspect Answer before considering another write.
