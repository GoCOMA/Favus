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
		utils.Fatal("Failed to load configuration: %v", err)
	}

	// Initialize the S3 uploader.
	s3Uploader, err := uploader.NewUploader(cfg) // Changed to NewUploader for consistency
	if err != nil {
		utils.Fatal("Failed to initialize S3 uploader: %v", err)
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
			utils.Fatal("Usage: favus upload <local_file_path> <s3_key>")
		}
		localFilePath := os.Args[2]
		s3Key := os.Args[3]
		if err := s3Uploader.UploadFile(localFilePath, s3Key); err != nil {
			utils.Fatal("Upload failed: %v", err)
		}
		utils.Info("File uploaded successfully.")

	case "resume":
		if len(os.Args) != 3 {
			utils.Fatal("Usage: favus resume <status_file_path>")
		}
		statusFilePath := os.Args[2]
		if err := s3Uploader.ResumeUpload(statusFilePath); err != nil {
			utils.Fatal("Resume failed: %v", err)
		}
		utils.Info("File upload resumed and completed successfully.")

	case "delete":
		if len(os.Args) != 3 {
			utils.Fatal("Usage: favus delete <s3_key>")
		}
		s3Key := os.Args[2]
		if err := s3Uploader.DeleteFile(s3Key); err != nil {
			utils.Fatal("Deletion failed: %m", err) // Fixed typo: %v -> %m
		}
		utils.Info("File deleted successfully.")

	case "list-uploads":
		if len(os.Args) != 2 {
			utils.Fatal("Usage: favus list-uploads")
		}
		uploads, err := s3Uploader.ListMultipartUploads()
		if err != nil {
			utils.Fatal("Failed to list multipart uploads: %v", err)
		}
		if len(uploads) == 0 {
			utils.Info("No ongoing multipart uploads found.")
			return
		}
		utils.Info("Ongoing multipart uploads:")
		for _, upload := range uploads {
			utils.Info("  UploadID: %s, Key: %s, Initiated: %s", *upload.UploadId, *upload.Key, upload.Initiated.Format(time.RFC3339))
		}

	default:
		utils.Fatal("Unknown command: %s", command)
	}
}