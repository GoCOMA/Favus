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
	lsObjectsBucket         string
	lsObjectsPrefix         string
	lsObjectsMax            int32
	lsObjectsWithIncomplete bool
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

		// 1) Load effective config prepared by PersistentPreRunE
		conf := GetLoadedConfig()
		if conf == nil {
			return fmt.Errorf("config not loaded")
		}

		// 2) Flag overrides
		effBucket := strings.TrimSpace(lsObjectsBucket)
		if effBucket == "" {
			effBucket = strings.TrimSpace(conf.Bucket)
		}
		if effBucket == "" {
			return fmt.Errorf("S3 bucket name is required (use --bucket or config/ENV)")
		}
		conf.Bucket = effBucket

		// 3) AWS config (LocalStack or real AWS)
		awsCfg, err := awsutils.LoadAWSConfig(profile)
		if err != nil {
			return fmt.Errorf("load aws config: %w", err)
		}

		// 3) Create uploader and list
		up, err := uploader.NewUploaderWithAWSConfig(conf, awsCfg)
		if err != nil {
			return fmt.Errorf("init uploader: %w", err)
		}

		// 4) Load object list from S3
		objects, err := up.ListObjects(strings.TrimSpace(lsObjectsPrefix), lsObjectsMax)
		if err != nil {
			return fmt.Errorf("fail to load objects: %w", err)
		}

		printedAnything := false

		if len(objects) > 0 {
			fmt.Println("Objects:")
			fmt.Fprintf(os.Stdout, "%5s  %-12s  %-12s  %-20s  %s\n", "#", "Size", "Storage", "Modified", "Key")
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
				fmt.Fprintf(os.Stdout, "%5d  %-12d  %-12s  %-20s  %s\n", i+1, size, sc, mod, aws.ToString(o.Key))
			}
			printedAnything = true
		}

		// 5) Optionally list incomplete multipart uploads
		if lsObjectsWithIncomplete {
			uploads, err := up.ListMultipartUploads()
			if err != nil {
				return fmt.Errorf("fail to load incomplete uploads: %w", err)
			}
			if len(uploads) > 0 {
				if printedAnything {
					fmt.Println()
				}
				fmt.Println("Incomplete multipart uploads:")
				fmt.Fprintf(os.Stdout, "%-36s  %-20s  %s\n", "UploadID", "Initiated(UTC)", "Key")
				for _, u := range uploads {
					uid := "-"
					key := "-"
					initiated := "-"
					if u.UploadId != nil {
						uid = *u.UploadId
					}
					if u.Key != nil {
						key = *u.Key
					}
					if u.Initiated != nil {
						initiated = u.Initiated.UTC().Format(time.RFC3339)
					}
					fmt.Fprintf(os.Stdout, "%-36s  %-20s  %s\n", uid, initiated, key)
				}
				printedAnything = true
			}
		}

		if !printedAnything {
			fmt.Println("(no results)")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listObjectsCmd)
	listObjectsCmd.Flags().StringVarP(&lsObjectsBucket, "bucket", "b", "", "Target S3 bucket name (overrides config/ENV)")
	listObjectsCmd.Flags().StringVarP(&lsObjectsPrefix, "prefix", "p", "", "Optional key prefix to filter objects")
	listObjectsCmd.Flags().Int32VarP(&lsObjectsMax, "max", "m", 0, "Max number of results to return (0 = unlimited)")
	listObjectsCmd.Flags().BoolVarP(&lsObjectsWithIncomplete, "with-incomplete", "i", false, "Also show incomplete multipart uploads")
}
