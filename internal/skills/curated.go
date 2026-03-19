package skills

// curatedSkills contains hand-written skill templates per tool.
// These are used when the tool doesn't ship its own skills.
// The key is the tool name; the value is the SKILL.md content template.
var curatedSkills = map[string]string{

	"gh": `---
name: github
description: >
  Use when the user needs to interact with GitHub — repos, issues, pull
  requests, Actions, releases, or the GitHub API. The gh CLI is installed
  and authenticated.
allowed-tools: Bash(gh:*)
---

You have the ` + "`gh`" + ` CLI (v{{.Version}}) installed{{if .AuthUser}} and authenticated as **{{.AuthUser}}**{{end}}.

## Key commands
- ` + "`gh pr create/list/view/merge`" + ` — pull requests
- ` + "`gh issue create/list/view/close`" + ` — issues
- ` + "`gh run list/view/watch`" + ` — GitHub Actions runs
- ` + "`gh release create/list/view`" + ` — releases
- ` + "`gh repo create/clone/fork/view`" + ` — repositories
- ` + "`gh api <endpoint>`" + ` — arbitrary GitHub API calls (REST or GraphQL)

## Conventions
- Use ` + "`--json <fields> -q <jq-expr>`" + ` for structured output when parsing results
- Prefer gh subcommands over raw ` + "`gh api`" + ` calls when a subcommand exists
- For complex queries, use ` + "`gh api graphql -f query='...'`" + `
- Check ` + "`gh pr checks`" + ` before merging
- Use ` + "`gh run watch`" + ` to monitor CI in real-time

## Dynamic context
- Current repo: ` + "!`gh repo view --json nameWithOwner -q .nameWithOwner 2>/dev/null || echo 'not in a GitHub repo'`" + `
- Open PRs: ` + "!`gh pr list --limit 5 --json number,title 2>/dev/null || echo 'none'`" + `
{{- if .NeedsAuth}}

## Auth
Managed by Clinic. Token injected via ` + "`GH_TOKEN`" + ` env var.
If auth fails, run ` + "`clinic auth gh`" + ` or ` + "`clinic doctor`" + `.
{{- end}}
`,

	"aws": `---
name: aws
description: >
  Use when the user needs to interact with AWS services — EC2, S3, Lambda,
  IAM, CloudFormation, ECS, RDS, and all other AWS services. The aws CLI
  is installed and authenticated.
allowed-tools: Bash(aws:*)
---

You have the ` + "`aws`" + ` CLI (v{{.Version}}) installed{{if .AuthUser}} and authenticated{{end}}.

## Key commands
- ` + "`aws s3 ls/cp/sync`" + ` — S3 bucket operations
- ` + "`aws ec2 describe-instances`" + ` — EC2 management
- ` + "`aws lambda invoke/update-function-code`" + ` — Lambda functions
- ` + "`aws ecs list-services/update-service`" + ` — ECS containers
- ` + "`aws cloudformation deploy`" + ` — infrastructure deployment
- ` + "`aws sts get-caller-identity`" + ` — verify current auth

## Conventions
- Always use ` + "`--output json`" + ` for structured output, pipe through ` + "`jq`" + ` for filtering
- Use ` + "`--query`" + ` (JMESPath) for server-side filtering when possible
- Use ` + "`--region`" + ` flag or ` + "`AWS_DEFAULT_REGION`" + ` env var
- Prefer ` + "`aws cloudformation deploy`" + ` over ` + "`create-stack`" + ` for idempotency
- Check ` + "`aws sts get-caller-identity`" + ` if auth seems broken
{{- if .NeedsAuth}}

## Auth
Managed by Clinic. Uses AWS credential chain (env vars, config files, instance metadata).
If auth fails, run ` + "`clinic auth aws`" + ` or ` + "`clinic doctor`" + `.
{{- end}}
`,

	"gcloud": `---
name: gcloud
description: >
  Use when the user needs to interact with Google Cloud Platform — Compute Engine,
  Cloud Run, GKE, BigQuery, Cloud Storage, IAM, and all other GCP services.
  The gcloud CLI is installed and authenticated.
allowed-tools: Bash(gcloud:*)
---

You have the ` + "`gcloud`" + ` CLI (v{{.Version}}) installed{{if .AuthUser}} and authenticated{{end}}.

## Key commands
- ` + "`gcloud run deploy/services list`" + ` — Cloud Run
- ` + "`gcloud compute instances list/create/delete`" + ` — Compute Engine
- ` + "`gcloud container clusters list`" + ` — GKE
- ` + "`bq query/ls/mk`" + ` — BigQuery (separate binary, installed with gcloud SDK)
- ` + "`gsutil ls/cp/rsync`" + ` — Cloud Storage (separate binary, installed with gcloud SDK)
- ` + "`gcloud projects list`" + ` — project management
- ` + "`gcloud auth list`" + ` — verify current auth

## Conventions
- Use ` + "`--format=json`" + ` for structured output, pipe through ` + "`jq`" + `
- Set project with ` + "`--project`" + ` flag or ` + "`gcloud config set project`" + `
- Set region with ` + "`--region`" + ` flag or ` + "`gcloud config set compute/region`" + `
- Use ` + "`gcloud config configurations`" + ` to manage multiple projects/accounts
{{- if .NeedsAuth}}

## Auth
Managed by Clinic. Service account via ` + "`GOOGLE_APPLICATION_CREDENTIALS`" + ` or user auth.
If auth fails, run ` + "`clinic auth gcloud`" + ` or ` + "`clinic doctor`" + `.
{{- end}}
`,

	"stripe": `---
name: stripe
description: >
  Use when the user needs to interact with Stripe — payments, subscriptions,
  customers, invoices, webhook testing, and the Stripe API. The stripe CLI
  is installed and authenticated.
allowed-tools: Bash(stripe:*)
---

You have the ` + "`stripe`" + ` CLI (v{{.Version}}) installed{{if .AuthUser}} and authenticated{{end}}.

## Key commands
- ` + "`stripe listen --forward-to localhost:3000/webhook`" + ` — forward webhooks to local server
- ` + "`stripe trigger payment_intent.succeeded`" + ` — trigger test events
- ` + "`stripe customers list`" + ` — list resources (works for any resource type)
- ` + "`stripe logs tail`" + ` — real-time API request logs
- ` + "`stripe resources`" + ` — list all available resource types

## Conventions
- Use ` + "`--data`" + ` or ` + "`-d`" + ` for creating/updating resources
- Default output is JSON — pipe through ` + "`jq`" + ` for filtering
- Use ` + "`stripe listen`" + ` for local webhook development
- ` + "`stripe trigger`" + ` sends test events — safe in test mode
- Verify you're in test mode (keys start with ` + "`sk_test_`" + `)
{{- if .NeedsAuth}}

## Auth
Managed by Clinic. Token injected via ` + "`STRIPE_API_KEY`" + ` env var.
If auth fails, run ` + "`clinic auth stripe`" + ` or ` + "`clinic doctor`" + `.
{{- end}}
`,

	"terraform": `---
name: terraform
description: >
  Use when the user needs to manage infrastructure as code — plan, apply,
  and destroy cloud resources declaratively. Terraform is installed.
allowed-tools: Bash(terraform:*)
---

You have ` + "`terraform`" + ` (v{{.Version}}) installed.

## Key commands
- ` + "`terraform init`" + ` — initialize working directory, download providers
- ` + "`terraform plan`" + ` — preview changes without applying
- ` + "`terraform apply`" + ` — apply changes (always plan first!)
- ` + "`terraform destroy`" + ` — tear down infrastructure
- ` + "`terraform state list/show`" + ` — inspect current state
- ` + "`terraform fmt`" + ` — format HCL files
- ` + "`terraform validate`" + ` — check config syntax

## Conventions
- ALWAYS run ` + "`terraform plan`" + ` before ` + "`apply`" + ` and show the user the plan
- NEVER run ` + "`terraform destroy`" + ` without explicit user confirmation
- Use ` + "`-auto-approve`" + ` only when the user has reviewed the plan
- Use ` + "`terraform fmt -recursive`" + ` to format all files
- Use ` + "`-target`" + ` flag sparingly — prefer full plans
`,

	"kubectl": `---
name: kubectl
description: >
  Use when the user needs to manage Kubernetes clusters — pods, deployments,
  services, config maps, secrets, and other cluster resources. kubectl is
  installed.
allowed-tools: Bash(kubectl:*)
---

You have ` + "`kubectl`" + ` (v{{.Version}}) installed{{if .AuthUser}} and connected to a cluster{{end}}.

## Key commands
- ` + "`kubectl get pods/deployments/services`" + ` — list resources
- ` + "`kubectl describe <resource> <name>`" + ` — detailed info
- ` + "`kubectl logs <pod> [-f]`" + ` — view/stream logs
- ` + "`kubectl apply -f <file>`" + ` — apply manifests
- ` + "`kubectl exec -it <pod> -- <cmd>`" + ` — run commands in a pod
- ` + "`kubectl port-forward <pod> <local>:<remote>`" + ` — tunnel to a pod

## Conventions
- Use ` + "`-o json`" + ` or ` + "`-o yaml`" + ` for structured output
- Use ` + "`-n <namespace>`" + ` to target specific namespaces
- Use ` + "`kubectl get all`" + ` for a quick overview
- Prefer ` + "`kubectl apply`" + ` over ` + "`kubectl create`" + ` for idempotency
- Use ` + "`kubectl diff -f <file>`" + ` to preview changes before applying
{{- if .NeedsAuth}}

## Auth
Uses kubeconfig at ` + "`~/.kube/config`" + `. Run ` + "`clinic auth kubectl`" + ` if cluster connection fails.
{{- end}}
`,

	"firebase": `---
name: firebase
description: >
  Use when the user needs to manage Firebase services — Authentication,
  Firestore, Hosting, Cloud Functions, Extensions, and Emulators. The
  firebase CLI is installed.
allowed-tools: Bash(firebase:*)
---

You have the ` + "`firebase`" + ` CLI (v{{.Version}}) installed{{if .AuthUser}} and authenticated{{end}}.

## Key commands
- ` + "`firebase deploy [--only hosting|functions|firestore]`" + ` — deploy services
- ` + "`firebase emulators:start`" + ` — run local emulators
- ` + "`firebase projects:list`" + ` — list projects
- ` + "`firebase use <project-id>`" + ` — switch active project
- ` + "`firebase functions:log`" + ` — view Cloud Functions logs
- ` + "`firebase hosting:channel:deploy <channel>`" + ` — preview deployments

## Conventions
- Use ` + "`--only`" + ` flag to deploy specific services (faster, safer)
- Use ` + "`firebase emulators:start`" + ` for local development
- Use ` + "`--project`" + ` flag to target specific projects
- ` + "`firebase.json`" + ` in the project root defines deployment config
{{- if .NeedsAuth}}

## Auth
Managed by Clinic. Service account via ` + "`GOOGLE_APPLICATION_CREDENTIALS`" + ` or user auth.
If auth fails, run ` + "`clinic auth firebase`" + ` or ` + "`clinic doctor`" + `.
{{- end}}
`,

	"supabase": `---
name: supabase
description: >
  Use when the user needs to manage Supabase — local development, database
  migrations, edge functions, and project management. The supabase CLI
  is installed.
allowed-tools: Bash(supabase:*)
---

You have the ` + "`supabase`" + ` CLI (v{{.Version}}) installed{{if .AuthUser}} and authenticated{{end}}.

## Key commands
- ` + "`supabase start`" + ` — start local development stack (Postgres, Auth, Storage, etc.)
- ` + "`supabase stop`" + ` — stop local stack
- ` + "`supabase db diff`" + ` — generate migration from local changes
- ` + "`supabase db push`" + ` — push migrations to remote
- ` + "`supabase migration new <name>`" + ` — create a new migration
- ` + "`supabase functions serve/deploy`" + ` — edge functions
- ` + "`supabase gen types typescript`" + ` — generate TypeScript types from schema

## Conventions
- Use ` + "`supabase db diff`" + ` to capture schema changes as migrations
- Use ` + "`supabase gen types`" + ` after schema changes for type safety
- Local dev runs on ` + "`localhost:54321`" + ` (API) and ` + "`localhost:54323`" + ` (Studio)
{{- if .NeedsAuth}}

## Auth
Managed by Clinic. Token injected via ` + "`SUPABASE_ACCESS_TOKEN`" + ` env var.
If auth fails, run ` + "`clinic auth supabase`" + ` or ` + "`clinic doctor`" + `.
{{- end}}
`,

	"fly": `---
name: fly
description: >
  Use when the user needs to deploy and manage apps on Fly.io — deployments,
  scaling, volumes, secrets, and machine management. The fly CLI is installed.
allowed-tools: Bash(fly:*)
---

You have the ` + "`fly`" + ` CLI (v{{.Version}}) installed{{if .AuthUser}} and authenticated as **{{.AuthUser}}**{{end}}.

## Key commands
- ` + "`fly deploy`" + ` — deploy from Dockerfile or fly.toml
- ` + "`fly status`" + ` — app status and running machines
- ` + "`fly logs`" + ` — stream app logs
- ` + "`fly secrets set KEY=value`" + ` — manage secrets
- ` + "`fly scale count/memory/vm`" + ` — scaling
- ` + "`fly volumes list/create`" + ` — persistent storage
- ` + "`fly ssh console`" + ` — SSH into a running machine
- ` + "`fly apps list`" + ` — list all apps

## Conventions
- ` + "`fly.toml`" + ` in the project root defines app config
- Use ` + "`fly deploy --strategy rolling`" + ` for zero-downtime deploys
- Use ` + "`fly secrets`" + ` (not env vars in fly.toml) for sensitive values
{{- if .NeedsAuth}}

## Auth
Managed by Clinic. Token injected via ` + "`FLY_ACCESS_TOKEN`" + ` env var.
If auth fails, run ` + "`clinic auth fly`" + ` or ` + "`clinic doctor`" + `.
{{- end}}
`,
}
