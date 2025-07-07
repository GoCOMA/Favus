package favus

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgPath string
	debug   bool

	// 빌드시 ldflags로 주입 가능
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// 루트 명령 정의
var rootCmd = &cobra.Command{
	Use:   "favus",
	Short: "Favus - Reliable multipart uploader for S3",
	Long: `Favus is a command-line utility for automated multipart uploads to S3.
It chunks large files, uploads them concurrently, resumes broken transfers,
and visualizes progress. Minimal config. Maximum reliability.`,
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
			// TODO: load configuration
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		printBanner()
		_ = cmd.Help()
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

// ASCII Art Banner
func printBanner() {
	fmt.Println(`

 #####     ###  ### ##  ### ##   #### 
  #  ##     ##   #  ##  ##  #   ##  # 
 ####      # #   # ##   ## ##   ####  
 ## #     ## #   ###   ##  #      ### 
##       ## ##   ##    ##  #   ##  #  
###     ###  ##  #      ###    ####   
                                      
                                      

Welcome to Favus – S3 multipart upload automation tool.
Use 'favus --help' to see available commands.
`)
}
