package favus

import (
    "context"
	"fmt"
    "log"
    "os"

    "github.com/GoCOMA/Favus/internal/awsutils"
    "github.com/GoCOMA/Favus/internal/config"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

type CommandContext struct {
    Ctx     context.Context
    Config  *config.Config
    AwsCfg  aws.Config
    S3      *s3.Client
    Profile string
    Debug   bool
    Logger  *log.Logger
}

type ctxKey struct{}
var commandContextKey = ctxKey{}

// BuildCommandContext loads app config, AWS config and creates an S3 client.
// - profile: AWS profile (can be empty)
// - configPath: path to YAML config file (empty = use DefaultConfigPath / env)
// - debug: debug flag
func BuildCommandContext(parent context.Context, profile, configPath string, debug bool) (*CommandContext, context.Context, error) {
    // 1) Load application config (uses DefaultConfigPath() and env overrides if path=="")
    conf, err := config.LoadConfig(configPath)
    if err != nil {
        return nil, parent, fmt.Errorf("load app config: %w", err)
    }
    if conf == nil {
        conf = &config.Config{}
    }

    // 2) Load AWS SDK config (awsutils helper used elsewhere in project)
    awsCfg, err := awsutils.LoadAWSConfig(profile)
    if err != nil {
        return nil, parent, fmt.Errorf("load aws config: %w", err)
    }

    // 3) Create S3 client via centralized factory (AWS_ENDPOINT_URL optional)
    endpoint := os.Getenv("AWS_ENDPOINT_URL")
    usePathStyle := endpoint != ""
    s3Client := awsutils.NewS3Client(awsCfg, endpoint, usePathStyle)

    cc := &CommandContext{
        Ctx:     parent,
        Config:  conf,
        AwsCfg:  awsCfg,
        S3:      s3Client,
        Profile: profile,
        Debug:   debug,
        Logger:  log.Default(),
    }
    ctx := context.WithValue(parent, commandContextKey, cc)
    return cc, ctx, nil
}

// GetCommandContext retrieves the CommandContext from ctx (may return nil).
func GetCommandContext(ctx context.Context) *CommandContext {
    if v := ctx.Value(commandContextKey); v != nil {
        if cc, ok := v.(*CommandContext); ok {
            return cc
        }
    }
    return nil
}