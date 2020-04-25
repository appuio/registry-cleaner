package arc

import (
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	s3root     = "docker/registry/v2"
	s3RepoRoot = s3root + "/repositories"
)

// S3Client interacts with the registry S3 storage
type S3Client struct {
	MaxKeys *int64
	Bucket  *string
	s3      *s3.S3
}

// NewS3Client configures & returns a new S3Client
func NewS3Client(accessKey, secretKey, endpoint, bucket string) *S3Client {
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:         aws.String(endpoint),
		Region:           aws.String("us-east-1"),
		S3ForcePathStyle: aws.Bool(true),
	}
	s3Session := session.New(s3Config)

	return &S3Client{
		MaxKeys: aws.Int64(1000),
		Bucket:  aws.String(bucket),
		s3:      s3.New(s3Session),
	}
}

// ReadRepos lists all objects under s3RepoRoot and parses their meaning
func (c S3Client) ReadRepos() (map[string]*Repository, error) {
	repos := make(map[string]*Repository)
	ls := c.newListObjectsInput(s3RepoRoot)

	numObjects := 0
	for {
		res, err := c.s3.ListObjects(ls)
		if err != nil {
			return nil, err
		}

		numObjects += len(res.Contents)
		log.Printf("Got %d objects from S3", numObjects)

		for _, obj := range res.Contents {
			k := strings.TrimPrefix(*obj.Key, s3RepoRoot+"/")
			fragments := strings.Split(k, "/")
			repoName := fragments[0] + "/" + fragments[1]

			if repos[repoName] == nil {
				repos[repoName] = &Repository{Name: repoName}
			}
			r := repos[repoName]

			switch fragments[2] {
			case "_layers":
				if fragments[3] != "sha256" {
					log.Fatalln("Unknown algorithm:", fragments[3])
				}
				r.Layers = append(r.Layers, fragments[4])

			case "_manifests":
				switch fragments[3] {
				case "revisions":
					if fragments[4] != "sha256" {
						log.Fatalln("Unknown algorithm:", fragments[4])
					}
					r.ManifestRevisions = append(r.ManifestRevisions, fragments[4]+":"+fragments[5])

				default:
					log.Fatalln("Unknown manifest type:", fragments[3])
				}

			case "_uploads":
				r.Uploads = append(r.Uploads, strings.Join(fragments[3:], "/"))

			default:
				log.Fatalln("Unknown component:", fragments[2])
			}
		}

		if res.NextMarker == nil {
			break
		}

		ls.Marker = res.NextMarker
	}

	return repos, nil
}

func (c S3Client) newListObjectsInput(prefix string) *s3.ListObjectsInput {
	return &s3.ListObjectsInput{
		Bucket:  c.Bucket,
		MaxKeys: c.MaxKeys,
		Prefix:  aws.String(prefix),
	}
}
