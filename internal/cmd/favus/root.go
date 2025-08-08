package favus

import (
	"fmt"
	"github.com/GoCOMA/Favus/internal/config"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgPath      string
	debug        bool
	profile      string
	loadedConfig *config.Config // 설정 파일 내용

	// 빌드시 ldflags로 주입 가능
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func GetLoadedConfig() *config.Config {
	return loadedConfig
}

// 루트 명령 정의
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
                                      
                                      

Welcome to Favus – S3 multipart upload automation tool!!
Favus is a command-line utility for automated multipart uploads to S3.
It chunks large files, uploads them concurrently, resumes broken transfers,
and visualizes progress. Minimal config. Maximum reliability.
Use 'favus --help' to see available commands.
`,
	Example: `
  # Upload a 5GB file to S3
  favus upload --file video.mp4 --bucket my-bucket --key uploads/video.mp4

  # Resume an interrupted upload
  favus resume --file video.mp4 --upload-id xyz123`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if debug {
			fmt.Println("[Favus] Debug mode enabled")
		}
		if cfgPath != "" {
			fmt.Printf("[Favus] Loading config from %s\n", cfgPath)

			var err error
			loadedConfig, err = config.LoadConfig(cfgPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			fmt.Printf("[DEBUG] Config loaded → bucket: %s, key: %s\n", loadedConfig.Bucket, loadedConfig.Key)
		} else {
			if cfgPath != "" {
				fmt.Printf("[Favus] Loading config from %s\n", cfgPath)
				var err error
				loadedConfig, err = config.LoadConfig(cfgPath)
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}
			} else {
				fmt.Println("⚠️  config.yaml 파일이 제공되지 않았습니다.")
				fmt.Println("💬 필요한 값을 직접 입력하여 계속 진행합니다.")
				switch cmd.Name() {
				case "upload":
					loadedConfig = config.PromptForUploadConfig(bucket, objectKey)
				case "resume":
					loadedConfig = config.PromptForResumeConfig()
				case "ls-orphans":
					loadedConfig = config.PromptForSimpleBucket(lsOrphansBucket, lsOrphansRegion)
				default:
					fmt.Println("Unknown command for interactive config")
					os.Exit(1)
				}
			}
		}
		return nil
	},
}

// Execute는 main.go에서 호출됩니다.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// 전역 플래그 등록
	rootCmd.PersistentFlags().StringVarP(&cfgPath, "config", "c", "", "Path to config file")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging")
	rootCmd.PersistentFlags().StringVar(&profile, "profile", "", "AWS named profile to use")

	// 버전 명령 등록
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show Favus version info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Favus %s (commit: %s, date: %s)\n", version, commit, date)
		},
	})

	// upload, resume 명령은 각 파일의 init()에서 rootCmd.AddCommand()로 등록
}
