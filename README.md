# keptn-config-generator

Generates a KeptnApp Custom Resource for a given kubernetes manifest

## Pre-requisites
* Tested with go 1.19

## Usage
* clone this repository
  > git clone https://github.com/thschue/keptn-config-generator.git
* copy your manifests to the 'manifests' folder
* run `go run main.go`

## Parameters
* `--inputPath` - path to the folder containing the manifests
* `--outputPath` - path to the folder where the KeptnApp CR will be generated
* `--bump` - defines how the appVersion should be bumped if an App Manifest exists (default: patch, options: major, minor, patch