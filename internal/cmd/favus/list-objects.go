package favus

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/GoCOMA/Favus/internal/awsutils"
	"github.com/GoCOMA/Favus/internal/uploader"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/spf13/cobra"
)

var (
	lsObjectsBucket string
	lsObjectsPrefix string
	lsObjectsMax    int32
)

var listObjectsCmd = &cobra.Command{
	Use:   "ls-objects",
	Short: "List completed objects in an S3 bucket",
	Long:  `Lists S3 objects (not multipart sessions) in the specified bucket. Supports optional prefix filtering and max results.`,
	Example: `
  favus ls-objects --bucket my-bucket
  favus ls-objects --bucket my-bucket --prefix uploads/
  favus ls-objects --bucket my-bucket --prefix logs/ --max 50`,
	RunE: func(cmd *cobra.Command, args []string) error {
		conf := GetLoadedConfig()
		if conf == nil {
			return fmt.Errorf("config not loaded (PersistentPreRunE should have populated it)")
		}

		// Flag overrides
		effBucket := strings.TrimSpace(lsObjectsBucket)
		if effBucket == "" {
			effBucket = strings.TrimSpace(conf.Bucket)
		}
		if effBucket == "" {
			return fmt.Errorf("S3 bucket name is required (use --bucket or config/ENV)")
		}
		conf.Bucket = effBucket

		// AWS cfg
		awsCfg, err := awsutils.LoadAWSConfig(profile)
		if err != nil {
			return fmt.Errorf("load aws config: %w", err)
		}

		up, err := uploader.NewUploaderWithAWSConfig(conf, awsCfg)
		if err != nil {
			return fmt.Errorf("init uploader: %w", err)
		}

		objects, err := up.ListObjects(strings.TrimSpace(lsObjectsPrefix), lsObjectsMax)
		if err != nil {
			return err
		}

		if len(objects) == 0 {
			fmt.Println("(no objects)")
			return nil
		}

		// Print header
		fmt.Fprintf(os.Stdout, "%5s  %-12s  %-12s  %-10s  %s\n", "#", "Size", "Storage", "Modified", "Key")
		for i, o := range objects {
			size := aws.ToInt64(o.Size)
			sc := string(o.StorageClass)
			if sc == "" {
				sc = "-"
			}
			mod := "-"
			if o.LastModified != nil {
				mod = o.LastModified.UTC().Format(time.RFC3339)
			}
			fmt.Fprintf(os.Stdout, "%5d  %-12d  %-12s  %-10s  %s\n", i+1, size, sc, mod, aws.ToString(o.Key))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listObjectsCmd)
	listObjectsCmd.Flags().StringVarP(&lsObjectsBucket, "bucket", "b", "", "Target S3 bucket name (overrides config/ENV)")
	listObjectsCmd.Flags().StringVarP(&lsObjectsPrefix, "prefix", "p", "", "Optional key prefix to filter objects")
	listObjectsCmd.Flags().Int32VarP(&lsObjectsMax, "max", "m", 0, "Max number of results to return (0 = unlimited)")
}
