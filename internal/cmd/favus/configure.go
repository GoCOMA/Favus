package favus

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/GoCOMA/Favus/internal/config"
	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Interactively set Bucket/Key/Region and persist them for all commands",
	Long: `Ask for S3 Bucket, default Key, and Region once and save them to a persistent config file.
After this, other commands will skip interactive prompts for these fields, unless you override with flags or ENV.`,
	RunE: runConfigure,
}

func runConfigure(_ *cobra.Command, _ []string) error {
	// Prompt for required configuration values
	bucket := PromptRequired("🔧 Enter S3 bucket name")
	key := PromptRequired("📝 Enter default S3 object key")
	region := PromptWithDefault("🌐 Enter AWS Region", DefaultRegion)

	// Get config file path and create directory if needed
	path := config.DefaultConfigPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	// Create YAML content
	content := []byte(fmt.Sprintf(
		"bucket: %s\nkey: %s\nregion: %s\n",
		bucket, key, region,
	))

	// Atomic save to prevent corruption
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, content, 0o644); err != nil {
		return fmt.Errorf("write temp config: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("install config: %w", err)
	}

	fmt.Printf("✅ Saved config to %s\n", path)
	fmt.Println("이제부터 Bucket/Key/Region은 이 파일에서 자동으로 불러와서, 각 명령에서 따로 묻지 않습니다.")
	fmt.Println("필요 시 플래그(--bucket/--key/--region) 또는 ENV로 언제든지 덮어쓸 수 있어요.")
	return nil
}

func init() {
	rootCmd.AddCommand(configureCmd)
}
