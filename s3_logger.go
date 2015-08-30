package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/s3"
	"github.com/twitchscience/aws_utils/uploader"
	"github.com/twitchscience/gologging/gologging"
	keygen "github.com/twitchscience/gologging/key_name_generator"
)

const maxLogLines = 1000
const maxLogAge = time.Minute
const serviceName = "analyticsd"

func (app *AppContext) setupS3Logger() (*gologging.UploadLogger, error) {
	auth, err := aws.EnvAuth()

	if err != nil {
		log.Fatalln("Failed to find AWS credentials in env")
	}

	awsConnection := s3.New(
		auth,
		getAWSRegion(app.config.aws_region),
	)
	bucket := awsConnection.Bucket(app.config.bucket)
	instanceInfo := keygen.BuildInstanceInfo(
		&keygen.EnvInstanceFetcher{},
		serviceName,
		app.config.logging_dir,
	)

	rotateCoordinator := gologging.NewRotateCoordinator(maxLogLines, maxLogAge)

	return gologging.StartS3Logger(
		rotateCoordinator,
		instanceInfo,
		&stderrNotifier{},
		&uploader.S3UploaderBuilder{
			Bucket: bucket,
			KeyNameGenerator: &KeyNameGenerator{
				Info:   instanceInfo,
				Prefix: app.config.key_prefix,
			},
		},
		&stderrNotifier{},
		app.config.num_workers,
	)
}

func (app *AppContext) Logf(fmt string, args ...interface{}) {
	// Allow passing a single string instead of format + args
	if len(args) == 0 {
		args = append(args, fmt)
		fmt = "%s"
	}
	// Intercept call to UploadLogger in debug mode
	if app.config.debug {
		app.logger.Printf("logging in debug mode: "+fmt+"\n", args...)
	} else {
		app.s3log.Log(fmt, args...)
	}
}

// Translate a region name ("us-west-1") into a aws.Region
func getAWSRegion(regionName string) aws.Region {
	region, found := aws.Regions[regionName]
	if !found {
		panic("Unknown AWS region: " + regionName)
	}
	return region
}

// Generates S3 key names for the uploaded logs
type KeyNameGenerator struct {
	Info   *keygen.InstanceInfo
	Prefix string
}

func (gen *KeyNameGenerator) GetKeyName(filename string) string {
	now := time.Now()
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf(gen.Prefix+"%s/%d.%s.%08x.log.gz",
		now.Format("2006-01-02"), // current date
		now.Unix(),               // UNIX timestamp
		gen.Info.Node,            // hostname from environment variable
		b,                        // 8 random bytes
	)
}

// Logging of uploads and errors
type stderrNotifier struct{}

// Called when a log file is successfully uploaded to S3
func (s *stderrNotifier) SendMessage(r *uploader.UploadReceipt) error {
	// TODO: report statsd metric
	log.Println("uploaded to S3 key", r.KeyName)
	return nil
}

// Called when an error occurs uploading log files
func (s *stderrNotifier) SendError(e error) {
	// TODO: use statsd or SQS here
	log.Printf("error uploading logs to S3: %T %s", e, e)
}
