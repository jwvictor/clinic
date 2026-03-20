package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/togglemedia/clinic/internal/config"
	"github.com/togglemedia/clinic/internal/registry"
)

var listAll bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed tools in your workspace",
	RunE: func(cmd *cobra.Command, args []string) error {
		if listAll {
			return listAvailable()
		}
		return listInstalled()
	},
}

func init() {
	listCmd.Flags().BoolVarP(&listAll, "all", "a", false, "List all available tools (not just installed)")
}

func listInstalled() error {
	lf, err := config.LoadLockfile()
	if err != nil {
		return err
	}

	if len(lf.Tools) == 0 {
		fmt.Println("No tools in workspace. Run 'clinic init --stack <name>' to get started.")
		return nil
	}

	if lf.Project.Stack != "" {
		fmt.Printf("Stack: %s\n\n", lf.Project.Stack)
	}

	// Sort tool names
	names := make([]string, 0, len(lf.Tools))
	for name := range lf.Tools {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Printf("%-16s %-12s %s\n", "Tool", "Version", "Installed Via")
	fmt.Printf("%-16s %-12s %s\n", "────", "───────", "─────────────")

	for _, name := range names {
		t := lf.Tools[name]
		fmt.Printf("%-16s %-12s %s\n", name, t.Version, t.InstalledVia)
	}

	fmt.Printf("\n%d tools\n", len(lf.Tools))
	return nil
}

func listAvailable() error {
	reg := registry.Load()

	// Group by category
	categories := map[string][]registry.ToolDef{}
	for _, t := range reg.Tools {
		categories[t.Category] = append(categories[t.Category], t)
	}

	// Build sorted category list — known categories first in a logical order,
	// then any new categories alphabetically
	knownOrder := []string{"cloud", "deploy", "iac", "k8s", "payments", "ecommerce", "observability", "backend", "utility", "secrets", "social", "productivity", "email", "media", "finance", "news"}
	seen := map[string]bool{}
	var categoryOrder []string
	for _, cat := range knownOrder {
		if _, ok := categories[cat]; ok {
			categoryOrder = append(categoryOrder, cat)
			seen[cat] = true
		}
	}
	// Append any categories not in the known list
	var extra []string
	for cat := range categories {
		if !seen[cat] {
			extra = append(extra, cat)
		}
	}
	sort.Strings(extra)
	categoryOrder = append(categoryOrder, extra...)

	for _, cat := range categoryOrder {
		tools := categories[cat]
		sort.Slice(tools, func(i, j int) bool { return tools[i].Name < tools[j].Name })

		fmt.Printf("%s\n", categoryLabel(cat))
		for _, t := range tools {
			fmt.Printf("  %-16s %s\n", t.Name, t.Description)
		}
		fmt.Println()
	}

	fmt.Printf("%d tools available\n", len(reg.Tools))
	return nil
}

func categoryLabel(cat string) string {
	labels := map[string]string{
		"cloud":         "Cloud Providers",
		"deploy":        "Deployment Platforms",
		"iac":           "Infrastructure as Code",
		"k8s":           "Kubernetes",
		"payments":      "Payments & Commerce",
		"ecommerce":     "E-Commerce",
		"observability": "Observability",
		"backend":       "Backend & Databases",
		"utility":       "Utilities",
		"secrets":       "Secrets Management",
		"social":        "Social Media",
		"productivity":  "Productivity",
		"email":         "Email",
		"media":         "Media",
		"finance":       "Finance",
		"news":          "News",
	}
	if l, ok := labels[cat]; ok {
		return l
	}
	return cat
}
