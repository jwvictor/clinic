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

	"x-cli": `---
name: x-twitter
description: >
  Use when the user needs to interact with X (Twitter) — post tweets, search,
  read timeline, manage account. The x CLI is installed and authenticated.
allowed-tools: Bash(x:*)
---

You have the ` + "`x`" + ` CLI (v{{.Version}}) installed{{if .AuthUser}} and authenticated{{end}}.

## Key commands
- ` + "`x post \"your tweet text\"`" + ` — post a tweet
- ` + "`x search \"query\"`" + ` — search tweets
- ` + "`x timeline`" + ` — view your home timeline
- ` + "`x replies`" + ` — view replies to your tweets
- ` + "`x user <handle>`" + ` — view a user's profile
- ` + "`x delete <tweet-id>`" + ` — delete a tweet

## Conventions
- Keep tweets under 280 characters
- Use ` + "`--json`" + ` for structured output when parsing results
- NEVER post without explicit user confirmation of the tweet content
- NEVER delete tweets without user confirmation
{{- if .NeedsAuth}}

## Auth
Managed by Clinic. API key injected via ` + "`X_API_KEY`" + ` env var.
If auth fails, run ` + "`clinic auth x-cli`" + ` or ` + "`clinic doctor`" + `.
{{- end}}
`,

	// "postiz" uses vendor skills from gitroomhq/postiz-agent

	"discordo": `---
name: discord
description: >
  Use when the user needs to interact with Discord from the terminal — browse
  servers, read channels, send messages. Discordo is a lightweight Discord TUI.
allowed-tools: Bash(discordo:*)
---

You have ` + "`discordo`" + ` (v{{.Version}}) installed{{if .AuthUser}} and authenticated{{end}}.

## Key commands
- ` + "`discordo`" + ` — launch the interactive TUI
- Navigate with vim-style keybindings (j/k to move, Enter to select)
- Ctrl+N to switch servers, Ctrl+L to switch channels
- Type and press Enter to send messages

## Conventions
- Discordo is primarily an interactive TUI — it's best launched for the user to interact with directly
- NEVER send messages to Discord channels without explicit user confirmation
- Be aware that Discord tokens should never be logged or displayed
{{- if .NeedsAuth}}

## Auth
Managed by Clinic. Token injected via ` + "`DISCORD_TOKEN`" + ` env var.
If auth fails, run ` + "`clinic auth discordo`" + ` or ` + "`clinic doctor`" + `.
{{- end}}
`,

	// "notion" uses vendor skills from 4ier/notion-cli

	"slack": `---
name: slack
description: >
  Use when the user needs to interact with Slack — create apps, manage
  workflows, deploy functions. The official Slack CLI is installed.
allowed-tools: Bash(slack:*)
---

You have the ` + "`slack`" + ` CLI (v{{.Version}}) installed{{if .AuthUser}} and authenticated{{end}}.

## Key commands
- ` + "`slack create <app-name>`" + ` — create a new Slack app
- ` + "`slack deploy`" + ` — deploy app to Slack
- ` + "`slack run`" + ` — run app locally in development mode
- ` + "`slack trigger create`" + ` — create a workflow trigger
- ` + "`slack function list`" + ` — list app functions
- ` + "`slack auth info`" + ` — show current auth status
- ` + "`slack feedback`" + ` — send feedback to Slack

## Conventions
- Use ` + "`slack run`" + ` for local development, ` + "`slack deploy`" + ` for production
- Slack apps use the Deno runtime for functions
- ` + "`manifest.json`" + ` or ` + "`manifest.ts`" + ` in the project root defines app configuration
- Use ` + "`slack trigger`" + ` to create entry points for workflows
{{- if .NeedsAuth}}

## Auth
Managed by Clinic. Token injected via ` + "`SLACK_TOKEN`" + ` env var.
If auth fails, run ` + "`clinic auth slack`" + ` or ` + "`clinic doctor`" + `.
{{- end}}
`,

	"yt-dlp": `---
name: yt-dlp
description: >
  Use when the user needs to download video or audio from YouTube or other
  sites, extract metadata, get subtitles, or convert media formats. yt-dlp
  is installed.
allowed-tools: Bash(yt-dlp:*)
---

You have ` + "`yt-dlp`" + ` (v{{.Version}}) installed.

## Key commands
- ` + "`yt-dlp <url>`" + ` — download best quality video
- ` + "`yt-dlp -x --audio-format mp3 <url>`" + ` — extract audio as MP3
- ` + "`yt-dlp -f \"bestvideo+bestaudio\" <url>`" + ` — download best video + audio separately and merge
- ` + "`yt-dlp --list-formats <url>`" + ` — list all available formats
- ` + "`yt-dlp --write-subs --sub-langs en <url>`" + ` — download with subtitles
- ` + "`yt-dlp --write-info-json --skip-download <url>`" + ` — get metadata only
- ` + "`yt-dlp -o \"%(title)s.%(ext)s\" <url>`" + ` — custom output filename
- ` + "`yt-dlp --flat-playlist <playlist-url>`" + ` — list playlist contents without downloading

## Conventions
- Use ` + "`-f`" + ` to select specific quality/format (e.g., ` + "`-f 720`" + ` for 720p)
- Use ` + "`-o`" + ` to control output filename template
- Use ` + "`--restrict-filenames`" + ` for safe filenames (no spaces/special chars)
- Use ` + "`--download-archive done.txt`" + ` to avoid re-downloading
- Supports 1000+ sites beyond YouTube — just pass any supported URL
- Use ` + "`--cookies-from-browser chrome`" + ` if a video requires authentication
`,

	"circumflex": `---
name: hackernews
description: >
  Use when the user wants to browse Hacker News — read top stories, view
  comments, or find tech news. Circumflex (clx) is installed.
allowed-tools: Bash(clx:*)
---

You have ` + "`clx`" + ` (circumflex, v{{.Version}}) installed.

## Key commands
- ` + "`clx`" + ` — launch the interactive Hacker News TUI
- Navigate with arrow keys or vim-style j/k
- Enter to open article in Reader Mode
- ` + "`c`" + ` to view comments
- ` + "`o`" + ` to open in browser
- ` + "`f`" + ` to toggle favorites

## Conventions
- Circumflex is primarily an interactive TUI — launch it for the user to browse
- Reader Mode renders articles in the terminal (no browser needed)
- Comments are rendered with syntax highlighting and threading
- Use it when the user asks about tech news, trending topics, or HN discussions
`,

	// "linear" uses vendor skills from schpet/linear-cli

	"rclone": `---
name: rclone
description: >
  Use when the user needs to sync, copy, or manage files across cloud storage
  providers — S3, Google Cloud Storage, Dropbox, OneDrive, SFTP, and 70+ others.
  rclone is installed.
allowed-tools: Bash(rclone:*)
---

You have ` + "`rclone`" + ` (v{{.Version}}) installed.

## Key commands
- ` + "`rclone listremotes`" + ` — show configured remotes
- ` + "`rclone ls remote:path`" + ` — list files
- ` + "`rclone copy src:path dst:path`" + ` — copy files (non-destructive)
- ` + "`rclone sync src:path dst:path`" + ` — sync (makes dst match src, deletes extras)
- ` + "`rclone move src:path dst:path`" + ` — move files
- ` + "`rclone mkdir remote:path`" + ` — create directory
- ` + "`rclone check src:path dst:path`" + ` — check if files match
- ` + "`rclone mount remote:path /local/mount`" + ` — mount remote as local filesystem
- ` + "`rclone config`" + ` — interactive remote configuration

## Conventions
- Use ` + "`--dry-run`" + ` before any ` + "`sync`" + ` or destructive operation
- NEVER run ` + "`rclone sync`" + ` without user confirmation — it deletes files at the destination
- Use ` + "`rclone copy`" + ` instead of ` + "`sync`" + ` when you just want to add files
- Use ` + "`--progress`" + ` or ` + "`-P`" + ` to show transfer progress
- Use ` + "`--filter`" + ` or ` + "`--include/--exclude`" + ` to limit what gets transferred
- Remote paths use ` + "`remote:path`" + ` format (e.g., ` + "`s3:mybucket/folder`" + `)
{{- if .NeedsAuth}}

## Auth
Each remote is configured separately via ` + "`rclone config`" + `.
Run ` + "`clinic auth rclone`" + ` or ` + "`rclone config`" + ` to add a new remote.
{{- end}}
`,

	"shopify": `---
name: shopify
description: >
  Use when the user needs to build Shopify apps, manage themes, or interact
  with Shopify stores. The Shopify CLI is installed.
allowed-tools: Bash(shopify:*)
---

You have the ` + "`shopify`" + ` CLI (v{{.Version}}) installed{{if .AuthUser}} and authenticated{{end}}.

## Key commands
- ` + "`shopify app dev`" + ` — start local app development server
- ` + "`shopify app deploy`" + ` — deploy app to Shopify
- ` + "`shopify app generate extension`" + ` — scaffold a new extension
- ` + "`shopify theme dev`" + ` — start local theme development with hot reload
- ` + "`shopify theme push`" + ` — push theme to store
- ` + "`shopify theme pull`" + ` — pull theme from store
- ` + "`shopify theme list`" + ` — list themes on a store
- ` + "`shopify hydrogen dev`" + ` — develop Hydrogen storefront locally
- ` + "`shopify auth login`" + ` — authenticate with Shopify

## Conventions
- Use ` + "`--store`" + ` flag or ` + "`SHOPIFY_FLAG_STORE`" + ` env var to target a specific store
- Use ` + "`shopify theme dev`" + ` for live preview during theme development
- For CI/CD, set ` + "`SHOPIFY_CLI_PARTNERS_TOKEN`" + ` (apps) or ` + "`SHOPIFY_CLI_THEME_TOKEN`" + ` (themes)
- ` + "`shopify.app.toml`" + ` defines app configuration
{{- if .NeedsAuth}}

## Auth
Managed by Clinic. Token injected via ` + "`SHOPIFY_CLI_PARTNERS_TOKEN`" + ` env var.
If auth fails, run ` + "`clinic auth shopify`" + ` or ` + "`clinic doctor`" + `.
{{- end}}
`,

	"datadog": `---
name: datadog
description: >
  Use when the user needs to interact with Datadog from CI/CD — upload test
  results, source maps, deploy markers, or run Synthetic tests. The datadog-ci
  CLI is installed.
allowed-tools: Bash(datadog-ci:*)
---

You have ` + "`datadog-ci`" + ` (v{{.Version}}) installed.

## Key commands
- ` + "`datadog-ci sourcemaps upload`" + ` — upload JS source maps for error tracking
- ` + "`datadog-ci junit upload`" + ` — upload JUnit test results
- ` + "`datadog-ci synthetics run-tests`" + ` — trigger Synthetic monitoring tests
- ` + "`datadog-ci tag`" + ` — add tags to CI pipeline traces
- ` + "`datadog-ci deployment mark`" + ` — mark a deployment in Datadog
- ` + "`datadog-ci dsyms upload`" + ` — upload iOS dSYM files for crash symbolication
- ` + "`datadog-ci git-metadata upload`" + ` — upload git metadata for linking commits

## Conventions
- Requires ` + "`DD_API_KEY`" + ` for all commands
- Some commands also need ` + "`DD_APP_KEY`" + `
- Set ` + "`DD_SITE`" + ` for non-US regions (e.g., ` + "`datadoghq.eu`" + `, ` + "`us5.datadoghq.com`" + `)
- Designed for CI pipelines — all config is via env vars, no interactive login
- NEVER log or display API keys
{{- if .NeedsAuth}}

## Auth
Purely env-var based. Set ` + "`DD_API_KEY`" + ` and optionally ` + "`DD_APP_KEY`" + `.
Get keys from Datadog → Organization Settings → API Keys.
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
