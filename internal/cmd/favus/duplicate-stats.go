package favus

import (
	"context"
	"fmt"

	"github.com/GoCOMA/Favus/internal/config"
	"github.com/GoCOMA/Favus/internal/duplicate"
	"github.com/spf13/cobra"
)

var duplicateStatsCmd = &cobra.Command{
	Use:   "duplicate-stats",
	Short: "Show duplicate checker statistics",
	Long: `Display statistics about the Cuckoo Filter and Count-Min Sketch
used for duplicate file detection.`,
	RunE: runDuplicateStats,
}

func runDuplicateStats(_ *cobra.Command, _ []string) error {
	// Use default config for duplicate checker
	conf := &config.Config{
		Region: "ap-northeast-2",
	}

	// Create duplicate checker
	dc, err := duplicate.NewDuplicateChecker(conf)
	if err != nil {
		return fmt.Errorf("failed to create duplicate checker: %w", err)
	}
	defer dc.Close()

	// Get statistics
	stats, err := dc.GetStats(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get statistics: %w", err)
	}

	// Display statistics
	fmt.Println("Duplicate Checker Statistics")
	fmt.Println("================================")

	if cfStats, ok := stats["cuckoo_filter"]; ok {
		fmt.Printf("🪣 Cuckoo Filter: %+v\n", cfStats)
	}

	if cmsStats, ok := stats["count_min_sketch"]; ok {
		fmt.Printf("📈 Count-Min Sketch: %+v\n", cmsStats)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(duplicateStatsCmd)
}
