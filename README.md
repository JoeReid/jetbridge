# JetBridge - The missing bridge between NATS and AWS Serverless

Both NATS/Jetstream and AWS Lambda are awesome tools, but they don't play well together.
JetBridge provides developers with a simple way to trigger Lambda executions to handle
Jetstream messages.

To keep deployment and management as simple as possible, JetBridge:

* Is a single statically-linked go binary which runs as a stateless service .
* Requires only a single DynamoDB table for state management and peer-discovery.
* Can be Auto-Scaled horizontally using only CPU and memory utilisation metrics.
* Can be managed via:
    * REST API (In development)
    * CLI tool (In development)
    * Terraform Provider (Planned for the future)
    * GitHub Action (Planned for the future)