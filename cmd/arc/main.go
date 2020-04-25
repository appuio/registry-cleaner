package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/appuio/arc"
)

const (
	root = "/docker/registry/v2"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	osc := getOpenShiftClient()
	s3c := getS3Client()
	reg := getRegistryClient()

	osImages := must(osc.ImageMap()).(map[string]string)
	log.Printf("Found %d Images in OpenShift", len(osImages))

	repos, err := s3c.ReadRepos()
	if err != nil {
		return err
	}
	log.Printf("Got %d repos from S3", len(repos))

	for name, r := range repos {
		if !r.HasLayers() && !r.HasManifests() {
			log.Println(name, "has neither Layers nor Manifests, deleting")
			if err := reg.DeleteRepo(name); err != nil {
				log.Fatalln("Error deleting Repo:", err)
			}
			continue
		}

		hasMatch := false
		for _, digest := range r.ManifestRevisions {
			if osImages[digest] != "" {
				log.Printf("Found at least one manifest revision for '%s' in OpenShift, skipping", name)
				hasMatch = true
				break
			}
		}

		if !hasMatch {
			log.Printf("Found no manifest revision for '%s' in OpenShift, deleting!", name)
			if err := reg.DeleteRepo(name); err != nil {
				log.Fatalln("Error deleting Repo:", err)
			}
		}
	}

	return nil
}

func must(v interface{}, err error) interface{} {
	if err != nil {
		log.Fatalln(err)
	}
	return v
}

func getOpenShiftClient() *arc.OpenShiftClient {
	home := must(os.UserHomeDir()).(string)
	kubeconfig := filepath.Join(home, ".kube", "config")

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
