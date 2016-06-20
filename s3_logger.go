package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/s3"
	"github.com/twitchscience/aws_utils/uploader"
	"github.com/twitchscience/gologging/gologging"
	keygen "github.com/twitchscience/gologging/key_name_generator"
)

const serviceName = "analyticsd"

func (app *AppContext) setupS3Logger() (err error) {
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

	rotateCoordinator := gologging.NewRotateCoordinator(
		app.config.max_log_lines,
		app.config.max_log_age,
	)

	metricsLogger := MetricsLogger{app.metrics}

	app.s3log, err = gologging.StartS3Logger(
		rotateCoordinator,
		instanceInfo,
		&metricsLogger,
		&uploader.S3UploaderBuilder{
			Bucket: bucket,
			KeyNameGenerator: &KeyNameGenerator{
				Info:   instanceInfo,
				Prefix: app.config.key_prefix,
			},
		},
		&metricsLogger,
		app.config.num_workers,
	)
	if err != nil {
		return
	}

	// Make sure logger is flushed when shutdown signal is received
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		log.Println("interrupted, closing logger...")
		app.s3log.Close()
		os.Exit(0)
	}()

	return nil
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
type MetricsLogger struct {
	Metrics
}

// Called when a log file is successfully uploaded to S3
func (s *MetricsLogger) SendMessage(r *uploader.UploadReceipt) error {
	log.Println("Uploaded to S3 key", r.KeyName)
	s.Uploads.Add(1)
	return nil
}

// Called when an error occurs uploading log files
func (s *MetricsLogger) SendError(e error) {
	log.Printf("error uploading logs to S3: %T %s", e, e)
	s.UploadErrors.Add(1)
}
