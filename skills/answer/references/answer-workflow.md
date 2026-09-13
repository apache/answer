<!--
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements. See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0.
-->

# Answer workflow

1. Confirm `question.read` and `answer.read` for context and `answer.create` for publication.
2. Fetch the question and all existing answers.
3. Determine whether an existing answer already resolves the question. Do not post a duplicate answer.
4. Draft a direct, self-contained Markdown answer. Clearly identify assumptions and uncertainty.
5. Show the exact answer to the user unless they supplied it verbatim with an explicit instruction to post.
6. After approval, write the body to a temporary file outside the repository and invoke `answer-cli answer create`.
7. Return the created answer ID or URL from the CLI response.
8. On `outcome_unknown`, fetch the question’s answers and look for the submitted content before considering a retry.
