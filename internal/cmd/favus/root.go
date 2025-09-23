package favus

import (
	"fmt"
	"os"

	"github.com/GoCOMA/Favus/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfgPath      string
	debug        bool
	profile      string
	loadedConfig *config.Config

	// ldflags -X로 주입 가능
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func GetLoadedConfig() *config.Config { return loadedConfig }

type CommandType string

const (
	CmdUpload      CommandType = "upload"
	CmdResume      CommandType = "resume"
	CmdLsOrphans   CommandType = "ls-orphans"
	CmdDelete      CommandType = "delete"
	CmdListUploads CommandType = "list-uploads"
	CmdKillOrphans CommandType = "kill-orphans"
	CmdListBuckets CommandType = "list-buckets"
	CmdConfigure   CommandType = "configure"
	CmdUI          CommandType = "ui"
	CmdStopUI      CommandType = "stop-ui"
	CmdVersion     CommandType = "version"
	CmdHelp        CommandType = "help"
	CmdCompletion  CommandType = "completion"
)

func shouldSkipConfigLoading(cmdName string) bool {
	skipCommands := []string{string(CmdVersion), string(CmdHelp), string(CmdCompletion)}
	for _, skip := range skipCommands {
		if cmdName == skip {
			return true
		}
	}
	return false
}

func loadConfigFromFile(cfgPath string) (*config.Config, error) {
	if cfgPath == "" {
		return config.LoadConfig("")
	}
	return config.LoadConfig(cfgPath)
}

func applyCommandSpecificOverrides(cmdName string, cfg *config.Config) {
	switch CommandType(cmdName) {
	case CmdUpload:
		if bucket != "" {
			cfg.Bucket = bucket
		}
		if objectKey != "" {
			cfg.Key = objectKey
		}
	case CmdResume:
		if resumeBucket != "" {
			cfg.Bucket = resumeBucket
		}
		if resumeKey != "" {
			cfg.Key = resumeKey
		}
		if uploadID != "" {
			cfg.UploadID = uploadID
		}
	case CmdLsOrphans:
		if lsOrphansBucket != "" {
			cfg.Bucket = lsOrphansBucket
		}
		if lsOrphansRegion != "" {
			cfg.Region = lsOrphansRegion
		}
	}
}

func requiresInteractiveConfig(cmdName string, cfg *config.Config) bool {
	switch CommandType(cmdName) {
	case CmdUpload:
		return cfg.Bucket == "" || cfg.Key == ""
	case CmdLsOrphans:
		return cfg.Bucket == "" || cfg.Region == ""
	case CmdDelete:
		return false // delete handles prompts internally
	default:
		return false
	}
}

func promptForCommandConfig(cmdName string) *config.Config {
	switch CommandType(cmdName) {
	case CmdUpload:
		return config.PromptForUploadConfig(bucket, objectKey)
	case CmdResume:
		return config.PromptForResumeConfig()
	case CmdLsOrphans:
		return config.PromptForSimpleBucket(lsOrphansBucket, lsOrphansRegion)
	default:
		return nil
	}
}

// Root command
var rootCmd = &cobra.Command{
	Use:   "favus",
	Short: "Favus - Reliable multipart uploader for S3",
	Long: `

 #####     ###  ### ##  ### ##   #### 
  #  ##     ##   #  ##  ##  #   ##  # 
 ####      # #   # ##   ## ##   ####  
 ## #     ## #   ###   ##  #      ### 
##       ## ##   ##    ##  #   ##  #  
###     ###  ##  #      ###    ####   
                                      
                                      

Welcome to Favus – an S3 multipart upload automation tool.
Favus chunks large files, uploads them via S3 multipart,
safely resumes broken transfers by reconciling with S3 (ListParts),
and visualizes progress in the terminal (and optionally streams to a Web UI via 'favus ui').

• Config: YAML + ENV overrides (S3_BUCKET_NAME, AWS_REGION, CHUNK_SIZE)
• AWS auth: profiles/ENV; if missing in TTY, prompts for keys
• S3-compatible endpoints supported via AWS_ENDPOINT_URL (e.g., LocalStack/MinIO)

Use 'favus --help' to see available commands.
`,
	Example: `
  # Upload a 5GB file to S3 (multipart)
  favus upload --file ./video.mp4 --bucket my-bucket --key uploads/video.mp4

  # Resume an interrupted upload (status file from previous run)
  favus resume --file /tmp/video.mp4_abcd1234.upload_status --upload-id abcd1234

  # List ongoing multipart uploads in a bucket
  favus list-uploads --bucket my-bucket

  # Scan bucket for incomplete (orphan) multipart uploads
  favus ls-orphans --bucket my-bucket

  # Delete an object
  favus delete --bucket my-bucket --key uploads/video.mp4

  # Start local UI bridge to stream CLI events to a Web UI
  favus ui --endpoint ws://127.0.0.1:8765/ws --open

  # Stop the local UI bridge
  favus stop-ui
`,
	SilenceUsage:  true,
	SilenceErrors: true,

	PersistentPreRunE: setupConfigForCommand,
}

func setupConfigForCommand(cmd *cobra.Command, _ []string) error {
	if debug {
		fmt.Println("[Favus] Debug mode enabled")
	}

	// Skip config loading for informational commands
	if shouldSkipConfigLoading(cmd.Name()) {
		return nil
	}

	// Load config from file if specified, otherwise from ENV
	cfg, err := loadConfigFromFile(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Apply command-specific flag overrides
	applyCommandSpecificOverrides(cmd.Name(), cfg)

	// Check if interactive config is needed
	if requiresInteractiveConfig(cmd.Name(), cfg) {
		fmt.Println("⚠️  No config file and insufficient environment variables. Switching to interactive mode.")
		interactiveConfig := promptForCommandConfig(cmd.Name())
		if interactiveConfig == nil {
			return fmt.Errorf("unknown command for interactive config: %s", cmd.Name())
		}
		loadedConfig = interactiveConfig
	} else {
		loadedConfig = cfg
	}

	return nil
}

// Execute runs the root command and handles any errors
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// global flags
	rootCmd.PersistentFlags().StringVarP(&cfgPath, "config", "c", "", "Path to config file (YAML). If omitted, ENV is used and may fall back to prompts.")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging")
	rootCmd.PersistentFlags().StringVar(&profile, "profile", "", "AWS named profile to use")

	// version
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show Favus version info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Favus %s (commit: %s, date: %s)\n", version, commit, date)
		},
	})

	// subcommands are added in their own files' init()
}
