[![Build Status](https://travis-ci.org/codejanovic/docker-planb.svg?branch=develop)](https://travis-ci.org/codejanovic/docker-planb)
[![License](https://img.shields.io/github/license/mashape/apistatus.svg?maxAge=2592000)]()

# docker-planb
Simple docker volume backup utility written in Go. 

This project does only filter the volumes to backup and delegates everything else to a container based on [loomchild/volume-backup](https://github.com/loomchild/volume-backup). 

Running **Planb** requires at least a destination directory, where the backed up volumes may be persisted as archived (compressed) files. On every run, **Planb** groups all archives into a new subdirectory that is named by the date of its execution. The name of the subdirectory may be configured through the environment variable `DESTINATION_FOLDER_FORMAT` as seen below.
By default every volume found on the docker daemon will be backed up - to backup only certain volumes a regex may be passed through the environment variable `VOLUME_FILTER`.

# Docker

## Environment Variables
The container may be configure by the following environment variables

| ENV        | Default           | Required | Description  |
| :------------- |:-------------| :-----|:-----|
| DESTINATION      | no default |   Yes | destinaton directory on the host, where the backups should be stored in|
| VOLUME_FILTER     | . | No | a regex to match valid volumes for backup. By default all volumes are considered valid|
| DESTINATION_FOLDER_FORMAT | 2006-01-02_15h-04m-05s| No | name of the folder grouping all backup files, based on the timestamp of the execution. See [Go date format](https://stackoverflow.com/questions/20234104/how-to-format-current-time-using-a-yyyymmddhhmmss-format) for formatting help |
| LOOMCHILD_IMAGE_VERSION | latest |  No | which version to use of [loomchild/volume-backup](https://github.com/loomchild/volume-backup)|


## Run with `docker run`
```bash
docker run --rm --name planb \
-e DESTINATION=/some/destination/folder/on/your/host \
-v /var/run/docker.sock:/var/run/docker.sock  \
codejanovic/planb:develop
```

## Run with `docker-compose`
```docker-compose
version: '3'
services:
  planb:
    container_name: planb
    image: codejanovic/planb:develop          
    volumes:                                    
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - DESTINATION=/some/destination/folder/on/your/host
      - VOLUME_FILTER=.    
```
