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
                                      
                                      

Welcome to Favus – S3 multipart upload automation tool!
Favus chunks large files, uploads them concurrently, resumes broken transfers,
and visualizes progress. Minimal config. Maximum reliability.
Use 'favus --help' to see available commands.
`,
	Example: `
  # Upload a 5GB file to S3
  favus upload --file video.mp4 --bucket my-bucket --key uploads/video.mp4

  # Resume an interrupted upload
  favus resume --file upload.status --upload-id xyz123
`,
	SilenceUsage:  true,
	SilenceErrors: true,

	// Config loading order:
	// 1) --config provided → load file (ENV overlay is handled inside LoadConfig)
	// 2) else LoadConfig("") to get defaults + ENV overlay
	// 3) apply CLI flags on top to check completeness
	// 4) if still missing required fields → prompt
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if debug {
			fmt.Println("[Favus] Debug mode enabled")
		}

		// skip for informational commands
		switch cmd.Name() {
		case "version", "help", "completion":
			return nil
		}

		// 1) config file path
		if cfgPath != "" {
			var err error
			loadedConfig, err = config.LoadConfig(cfgPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			return nil
		}

		// 2) ENV-only baseline
		envCfg, _ := config.LoadConfig("")
		needPrompt := false

		// 3) consider flags per command (flags are package vars from each command file)
		switch cmd.Name() {
		case "upload":
			// required: Bucket, Key (file path is validated in upload.go)
			effBucket, effKey := envCfg.Bucket, envCfg.Key
			if bucket != "" {
				effBucket = bucket
			}
			if objectKey != "" {
				effKey = objectKey
			}
			if effBucket == "" || effKey == "" {
				needPrompt = true
			} else {
				envCfg.Bucket, envCfg.Key = effBucket, effKey
			}

		case "resume":
			// required: Bucket, Key, UploadID (status file path is validated in resume.go)
			effBucket, effKey, effUploadID := envCfg.Bucket, envCfg.Key, envCfg.UploadID
			if resumeBucket != "" {
				effBucket = resumeBucket
			}
			if resumeKey != "" {
				effKey = resumeKey
			}
			if uploadID != "" {
				effUploadID = uploadID
			}
			if effBucket == "" || effKey == "" || effUploadID == "" {
				needPrompt = true
			} else {
				envCfg.Bucket, envCfg.Key, envCfg.UploadID = effBucket, effKey, effUploadID
			}

		case "ls-orphans":
			// required: Bucket, Region
			effBucket, effRegion := envCfg.Bucket, envCfg.Region
			if lsOrphansBucket != "" {
				effBucket = lsOrphansBucket
			}
			if lsOrphansRegion != "" {
				effRegion = lsOrphansRegion
			}
			if effBucket == "" || effRegion == "" {
				needPrompt = true
			} else {
				envCfg.Bucket, envCfg.Region = effBucket, effRegion
			}

		case "delete":
			// delete는 커맨드 내부에서 프롬프트 처리 가능하므로 여기선 강제하지 않음
			needPrompt = false

		default:
			needPrompt = false
		}

		// 4) decide
		if !needPrompt {
			loadedConfig = envCfg
			return nil
		}

		// interactive fallback
		fmt.Println("⚠️  No config file and insufficient environment variables. Switching to interactive mode.")
		switch cmd.Name() {
		case "upload":
			loadedConfig = config.PromptForUploadConfig(bucket, objectKey)
		case "resume":
			loadedConfig = config.PromptForResumeConfig()
		case "ls-orphans":
			loadedConfig = config.PromptForSimpleBucket(lsOrphansBucket, lsOrphansRegion)
		default:
			return fmt.Errorf("unknown command for interactive config")
		}
		return nil
	},
}

// Called from cmd/main.go
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
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
