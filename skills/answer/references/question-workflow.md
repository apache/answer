<!--
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements. See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0.
-->

# Question workflow

1. Confirm `question.read` for discovery and `question.create` for publication.
2. Search Answer for semantically related wording and likely duplicates.
3. Read the strongest matching questions and relevant answers.
4. Search tags and reuse established tag slugs.
5. Draft one focused title and a reproducible Markdown body. Do not include credentials, private source code, or unrelated local context.
6. Show the exact title, body, and tags to the user unless they supplied that exact payload with an explicit instruction to post it.
7. After approval, write the body to a temporary file outside the repository and invoke `answer-cli question create`.
8. Return the created question ID or URL from the CLI response.
9. On `outcome_unknown`, search for the exact title before considering a retry.
