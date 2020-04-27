package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/appuio/arc"
)

const usage = `usage: arc [command]

commands:
  uploads: Delete repos that have uploads but no manifests or layers
  repos:   Delete repositories that are not known to OpenShift
  blobs:   Delete orphaned blobs
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatalln(err)
	}
}

func run(args []string) error {
	if len(args) != 1 {
		fmt.Print(usage)
		os.Exit(1)
	}

	osc := getOpenShiftClient()
	s3c := getS3Client()
	reg := getRegistryClient()

	switch args[0] {
	case "uploads":
		return cleanupInvalidUploads(s3c, reg)

	case "repos":
		return cleanupOrphanedRepos(s3c, reg, osc)

	case "blobs":
		return cleanupOrphanedBlobs(s3c)

	default:
		log.Println("Invalid command", args[0])
		fmt.Print(usage)
		os.Exit(1)
	}

	return nil
}

func cleanupInvalidUploads(s3c *arc.S3Client, reg *arc.RegistryClient) error {
	repos := must(s3c.ReadRepos()).(map[string]*arc.Repository)
	log.Printf("Got %d repos from S3", len(repos))

	for name, r := range repos {
		if !r.HasLayers() && !r.HasManifests() {
			log.Println(name, "has neither Layers nor Manifests, deleting")
			if err := reg.DeleteRepo(name); err != nil {
				return fmt.Errorf("Error deleting repo: %w", err)
			}
			continue
		}
	}

	return nil
}

func cleanupOrphanedRepos(s3c *arc.S3Client, reg *arc.RegistryClient, osc *arc.OpenShiftClient) error {
	osImages := must(osc.ImageMap()).(map[string]string)
	log.Printf("Found %d Images in OpenShift", len(osImages))

	repos := must(s3c.ReadRepos()).(map[string]*arc.Repository)
	log.Printf("Got %d repos from S3", len(repos))

	for name, r := range repos {
		hasMatch := false
		for _, digest := range r.ManifestRevisions {
			if osImages[digest] != "" {
				log.Printf("Found at least one manifest revision for '%s' in OpenShift, skipping", name)
				hasMatch = true
				break
			}
		}

		if hasMatch {
			continue
		}

		log.Printf("Found no manifest revision for '%s' in OpenShift, deleting!", name)
		if err := reg.DeleteRepo(name); err != nil {
			return fmt.Errorf("Error deleting Repo: %w", err)
		}
	}

	return nil
}

func cleanupOrphanedBlobs(s3c *arc.S3Client) error {
	return errors.New("Not yet implemented")
}

func must(v interface{}, err error) interface{} {
	if err != nil {
		log.Fatalln(err)
	}
	return v
}

func getOpenShiftClient() *arc.OpenShiftClient {
	kubeconfig := os.Getenv("ARC_KUBECONFIG_PATH")
	if kubeconfig == "" {
		home := must(os.UserHomeDir()).(string)
		kubeconfig = filepath.Join(home, ".kube", "config")
	}
	log.Println("Using kubeconfig at", kubeconfig)

	osc := must(arc.NewOpenShiftClient(kubeconfig)).(*arc.OpenShiftClient)

	log.Println("Using OpenShift at", osc.Host)
	return osc
}

func getS3Client() *arc.S3Client {
	s3c := arc.NewS3Client(
		os.Getenv("ARC_S3_ACCESS_KEY"),
		os.Getenv("ARC_S3_SECRET_KEY"),
		os.Getenv("ARC_S3_ENDPOINT"),
		os.Getenv("ARC_S3_BUCKET"),
	)

	log.Println("Using S3 at", os.Getenv("ARC_S3_ENDPOINT"), os.Getenv("ARC_S3_BUCKET"))
	return s3c
}

func getRegistryClient() *arc.RegistryClient {
	reg := must(arc.NewRegistryClientS3(
		os.Getenv("ARC_S3_ACCESS_KEY"),
		os.Getenv("ARC_S3_SECRET_KEY"),
		os.Getenv("ARC_S3_ENDPOINT"),
		os.Getenv("ARC_S3_BUCKET"),
	)).(*arc.RegistryClient)

	log.Println("Using Registry at", os.Getenv("ARC_S3_ENDPOINT"), os.Getenv("ARC_S3_BUCKET"))
	return reg
}
