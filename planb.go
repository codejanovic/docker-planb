package main

import (
	"os"
	"regexp"
	"time"

	"github.com/fsouza/go-dockerclient"
	log "github.com/sirupsen/logrus"
)

var _destinationFolder = os.Getenv("DESTINATION")
var _volumeFilter = os.Getenv("VOLUME_FILTER")
var _destinationFolderFormat = os.Getenv("DESTINATION_FOLDER_FORMAT")
var _loomchildImageVersion = os.Getenv("LOOMCHILD_IMAGE_VERSION")
var _loomchildImage = "loomchild/volume-backup"
var _dockerEndpoint = "unix:///var/run/docker.sock"

func pullImage(client *docker.Client, image string, tag string) {
	err := client.PullImage(docker.PullImageOptions{
		Repository:   image,
		Tag:          tag,
		OutputStream: os.Stdout},
		docker.AuthConfiguration{})

	if err != nil {
		log.WithFields(log.Fields{
			"image": _loomchildImage,
			"error": err,
		}).Error("Oops, unable to pull image :-(")
		panic(err)
	}
}

func connectToDockerDaemon(address string) *docker.Client {
	client, err := docker.NewClient(address)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Oops, unable to connect to docker daemon :-(")
		panic(err)
	}
	return client
}

func listAllVolumes(client *docker.Client) []docker.Volume {
	volumes, err := client.ListVolumes(docker.ListVolumesOptions{})

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Oops, listing all docker volumes failed :-(")
		panic(err)
	}

	return volumes
}

func loomchildImageWithVersion() string {
	return _loomchildImage + ":" + _loomchildImageVersion
}

func createContainer(client *docker.Client, image string, commands []string, volumes []string, autoremove bool) *docker.Container {
	container, err := client.CreateContainer(docker.CreateContainerOptions{Config: &docker.Config{Image: image, Cmd: commands}, HostConfig: &docker.HostConfig{Binds: volumes, AutoRemove: autoremove}})
	if err != nil {
		log.WithFields(log.Fields{
			"image": image,
			"error": err,
		}).Error("Oops, creating container failed :-(")
		panic(err)
	}
	return container
}

func commandsForBackupStrategy(volume docker.Volume) []string {
	backupName := volume.Name
	return []string{
		"backup",
		backupName,
	}
}

func bindVolumesFor(volume docker.Volume, folder string) []string {
	return []string{
		volume.Name + ":/volume",
		_destinationFolder + "/" + folder + ":/backup",
	}
}

func startContainer(client *docker.Client, container *docker.Container) {
	err := client.StartContainer(container.ID, &docker.HostConfig{})
	if err != nil {
		log.WithFields(log.Fields{
			"container": container.Name,
			"image":     container.Image,
			"error":     err,
		}).Error("Oops, starting container failed :-(")

		panic(err)
	}
}

func backupVolume(client *docker.Client, volume docker.Volume, folder string) {
	log.WithFields(log.Fields{
		"volume": volume.Name,
	}).Info("initiating volume backup...")

	log.WithFields(log.Fields{
		"volume":      volume.Name,
		"image":       _loomchildImage,
		"destination": _destinationFolder + "/" + folder,
	}).Info("starting volume-backup container...")

	container := createContainer(client, loomchildImageWithVersion(), commandsForBackupStrategy(volume), bindVolumesFor(volume, folder), true)
	startContainer(client, container)

	log.WithFields(log.Fields{
		"volume": volume.Name,
	}).Info("finished volume backup!")
}

func main() {
	log.SetFormatter(&log.TextFormatter{ForceColors: true})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	regex, _ := regexp.Compile(_volumeFilter)

	client := connectToDockerDaemon(_dockerEndpoint)
	pullImage(client, _loomchildImage, "latest")
	volumes := listAllVolumes(client)

	log.Infof("found a total of %d volumes to inspect...\n", len(volumes))

	folder := time.Now().Format(_destinationFolderFormat)
	for _, volume := range volumes {
		if regex.MatchString(volume.Name) {
			log.WithFields(log.Fields{
				"volume": volume.Name,
				"filter": _volumeFilter,
			}).Info("volume matched volume-filter")
			backupVolume(client, volume, folder)
		} else {
			log.WithFields(log.Fields{
				"volume": volume.Name,
				"filter": _volumeFilter,
			}).Info("volume did not match volume-filter, skipping...")
		}

	}
}
