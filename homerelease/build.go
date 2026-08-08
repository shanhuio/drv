package homerelease

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"time"

	"shanhu.io/drv/drvapi"
	"shanhu.io/drv/semver"
	"shanhu.io/g/jsonutil"
	"shanhu.io/std/docker"
	"shanhu.io/std/errcode"
)

type builder struct {
	out string
}

// imageSum from lets docker image build/pull output.
type imageSum struct {
	Origin string `json:",omitempty"`
}

// readImageStep reads the lets image sum for an image and builds its
// StepVersion, deriving the version from the origin image tag. sumFile is the
// path to the image's <name>.imgsum file; img is the checksum of its tarball.
func readImageStep(
	sumFile string, img *dockerImage,
) (*drvapi.StepVersion, error) {
	sum := new(imageSum)
	if err := jsonutil.ReadFile(sumFile, sum); err != nil {
		return nil, errcode.Annotatef(err, "read sum file %q", sumFile)
	}
	if sum.Origin == "" {
		return nil, errcode.InvalidArgf("origin missing in %q", sumFile)
	}

	_, tag := docker.ParseImageTag(sum.Origin)
	major, err := semver.Major(tag)
	if err != nil {
		return nil, errcode.Annotatef(err, "parse tag %q", sum.Origin)
	}

	return &drvapi.StepVersion{
		Major:    major,
		Version:  tag,
		Source:   sum.Origin,
		Image:    img.id,
		ImageSum: img.sum,
	}, nil
}

const (
	nextcloudMinVersion = 20
	nextcloudMaxVersion = 34 // NEXTCLOUD_UPGRADE
)

func (b *builder) buildRelease(name string) error {
	arts := new(drvapi.Artifacts)
	const repo = "shanhu.io/drv/dockers"

	images := make(map[string]*dockerImage)
	imageObjs := make(map[string]string)

	var imageNames []string
	var nextcloudImageNames []string

	for i := nextcloudMinVersion; i <= nextcloudMaxVersion; i++ {
		nextcloudImageNames = append(
			nextcloudImageNames, fmt.Sprintf("nextcloud%d", i),
		)
	}

	imageNames = append(imageNames, nextcloudImageNames...)
	imageNames = append(
		imageNames,

		"postgres12",
		"redis",

		"core",
		"doorway",
		"ncfront",
		"homeboot",
		"toolbox",
	)

	for _, d := range imageNames {
		log.Printf("checksuming %s", d)
		rel := filePath("", repo, d+".tar.gz")
		tgz := filepath.Join(b.out, rel)
		img, err := sumDockerTgz(tgz)
		if err != nil {
			return errcode.Annotatef(err, "checksum for %q", d)
		}
		images[d] = img
		imageObjs[img.sum] = rel
	}

	log.Println("building artifacts jsonx")
	imageSums := make(map[string]string)
	for _, entry := range []struct {
		name   string
		images []string
		steps  *[]*drvapi.StepVersion
		final  *string
	}{{
		name:   "nextcloud",
		images: nextcloudImageNames,
		steps:  &arts.Nextclouds,
		final:  &arts.Nextcloud,
	}, {
		name:   "postgres",
		images: []string{"postgres12"},
		steps:  &arts.Postgreses,
		final:  &arts.Postgres,
	}, {
		name:   "redis",
		images: []string{"redis"},
		final:  &arts.Redis,
	}} {
		final := ""
		for _, img := range entry.images {
			id := images[img]
			sumFile := filePath(b.out, repo, img+".imgsum")
			step, err := readImageStep(sumFile, id)
			if err != nil {
				return errcode.Annotatef(err, "image %q", img)
			}

			if entry.steps != nil {
				*entry.steps = append(*entry.steps, step)
			}
			final = id.id

			imageSums[id.id] = id.sum
		}
		if entry.final != nil {
			*entry.final = final
		}
	}

	for _, entry := range []struct {
		name string
		id   *string
	}{
		{name: "core", id: &arts.Jarvis},
		{name: "doorway", id: &arts.Doorway},
		{name: "ncfront", id: &arts.NCFront},
		{name: "homeboot", id: &arts.HomeBoot},
		{name: "toolbox", id: &arts.Toolbox},
	} {
		id := images[entry.name]
		*entry.id = id.id
		imageSums[id.id] = id.sum
	}

	arts.ImageSums = imageSums

	log.Printf("writing out artifacts.json")
	artsOut := filePath(b.out, repo, "artifacts.json")
	if err := jsonutil.WriteFileReadable(artsOut, arts); err != nil {
		return errcode.Annotate(err, "write out artifacts")
	}

	log.Printf("writing out release.json")
	rel := &drvapi.Release{
		Name:      name,
		Time:      time.Now(),
		Arch:      runtime.GOARCH,
		Artifacts: arts,
	}
	relOut := filePath(b.out, repo, "release.json")
	if err := jsonutil.WriteFileReadable(relOut, rel); err != nil {
		return errcode.Annotate(err, "write out release")
	}

	log.Printf("writing out objects")
	objOut := filePath(b.out, repo, "objs.tar")
	if err := writeObjects(objOut, imageObjs, b.out); err != nil {
		return errcode.Annotate(err, "writing out object archive")
	}

	return nil
}

func cmdBuild(args []string) error {
	flags := cmdFlags.New()
	out := flags.String("out", "_/out", "output directory")
	name := flags.String("name", "", "release name")
	_ = flags.ParseArgs(args)

	b := &builder{
		out: *out,
	}

	return b.buildRelease(*name)
}
