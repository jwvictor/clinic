package registry

func builtinStacks() []StackDef {
	return []StackDef{
		{
			Name:        "saas-founder",
			Description: "Full-stack SaaS founder toolkit",
			Tools:       []string{"gh", "vercel", "stripe", "supabase", "firebase", "sentry-cli", "gws", "jq", "ngrok"},
		},
		{
			Name:        "devops",
			Description: "DevOps/platform engineer toolkit",
			Tools:       []string{"aws", "gcloud", "az", "terraform", "kubectl", "helm", "docker", "vault", "gh", "sentry-cli", "jq"},
		},
		{
			Name:        "indie-hacker",
			Description: "Ship fast on modern platforms",
			Tools:       []string{"gh", "fly", "railway", "stripe", "supabase", "wrangler", "sentry-cli", "jq", "ngrok"},
		},
		{
			Name:        "frontend",
			Description: "Frontend developer deploying to edge platforms",
			Tools:       []string{"gh", "vercel", "netlify", "wrangler", "firebase", "sentry-cli", "jq"},
		},
		{
			Name:        "gcp-stack",
			Description: "Google Cloud focused stack",
			Tools:       []string{"gcloud", "gws", "firebase", "kubectl", "helm", "docker", "gh", "terraform", "jq"},
		},
		{
			Name:        "creator",
			Description: "Content creator and social media toolkit",
			Tools:       []string{"x-cli", "late", "yt-dlp", "ticker", "circumflex", "notion", "slack", "discordo"},
		},
	}
}
