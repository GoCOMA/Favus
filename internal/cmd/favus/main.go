package main

import (
	"fmt"
	"os"
	"time"

	"favus/internal/config"
	"favus/internal/uploader"
	"favus/pkg/utils"
)

func main() {
	fmt.Println("Favus - S3 Multipart Upload Automation Tool")

	// Load application configuration.
	cfg, err := config.LoadConfig()
	if err != nil {
		utils.Error(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	// Initialize the S3 uploader.
	s3Uploader, err := uploader.NewUploader(cfg) // Changed to NewUploader for consistency
	if err != nil {
		utils.Error(fmt.Sprintf("Failed to initialize S3 uploader: %v", err))
		os.Exit(1)
	}

	// Display usage if no command is provided.
	if len(os.Args) < 2 {
		fmt.Println("Usage: favus <command> [args...]")
		fmt.Println("Commands:")
		fmt.Println("  upload <local_file_path> <s3_key>")
		fmt.Println("  resume <status_file_path>")
		fmt.Println("  delete <s3_key>")
		fmt.Println("  list-uploads")
		os.Exit(1)
	}

	// Parse and execute the command.
	command := os.Args[1]

	switch command {
	case "upload":
		if len(os.Args) != 4 {
			utils.Error(fmt.Sprintf("Usage: favus upload <local_file_path> <s3_key>"))
			os.Exit(1)
		}
		localFilePath := os.Args[2]
		s3Key := os.Args[3]
		if err := s3Uploader.UploadFile(localFilePath, s3Key); err != nil {
			utils.Error(fmt.Sprintf("Upload failed: %v", err))
			os.Exit(1)
		}
		utils.Info("File uploaded successfully.")

	case "resume":
		if len(os.Args) != 3 {
			utils.Error(fmt.Sprintf("Usage: favus resume <status_file_path>"))
			os.Exit(1)
		}
		statusFilePath := os.Args[2]
		if err := s3Uploader.ResumeUpload(statusFilePath); err != nil {
			utils.Error(fmt.Sprintf("Resume failed: %v", err))
			os.Exit(1)
		}
		utils.Info("File upload resumed and completed successfully.")

	case "delete":
		if len(os.Args) != 3 {
			utils.Error(fmt.Sprintf("Usage: favus delete <s3_key>"))
			os.Exit(1)
		}
		s3Key := os.Args[2]
		if err := s3Uploader.DeleteFile(s3Key); err != nil {
			utils.Error(fmt.Sprintf("Deletion failed: %v", err))
			os.Exit(1)
		}
		utils.Info("File deleted successfully.")

	case "list-uploads":
		if len(os.Args) != 2 {
			utils.Error(fmt.Sprintf("Usage: favus list-uploads"))
			os.Exit(1)
		}
		uploads, err := s3Uploader.ListMultipartUploads()
		if err != nil {
			utils.Error(fmt.Sprintf("Failed to list multipart uploads: %v", err))
			os.Exit(1)
		}
		if len(uploads) == 0 {
			utils.Info("No ongoing multipart uploads found.")
			return
		}
		utils.Info("Ongoing multipart uploads:")
		for _, upload := range uploads {
			utils.Info(fmt.Sprintf("UploadID: %s, Key: %s, Initiated: %s", *upload.UploadId, *upload.Key, upload.Initiated.Format(time.RFC3339)))
		}

	default:
		utils.Error(fmt.Sprintf("Unknown command: %s", command))
		os.Exit(1)
	}
}
