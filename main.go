package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/coreos/go-semver/semver"
	keptnv1alpha2 "github.com/keptn/lifecycle-toolkit/operator/apis/lifecycle/v1alpha2"
	urcli "github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
	"hash/fnv"
	"io"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/cli-runtime/pkg/printers"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var appList = make(map[string]keptnv1alpha2.KeptnApp)

type config struct {
	Scheme             *runtime.Scheme
	InputPath          string
	OutputPath         string
	VersionUpgradeMode string
}

var c config

const workloadAnnotation = "keptn.sh/workload"
const versionAnnotation = "keptn.sh/version"
const appAnnotation = "keptn.sh/app"
const k8sRecommendedWorkloadAnnotations = "app.kubernetes.io/name"
const k8sRecommendedVersionAnnotations = "app.kubernetes.io/version"
const k8sRecommendedAppAnnotations = "app.kubernetes.io/part-of"

func main() {
	app := &urcli.App{
		Name: "keptn-config-generator",
		Flags: []urcli.Flag{
			&urcli.StringFlag{
				Name:        "bump",
				Value:       "patch",
				Usage:       "bump major, minor or patch",
				Destination: &c.VersionUpgradeMode,
			},
			&urcli.StringFlag{
				Name:        "inputPath",
				Value:       "data",
				Usage:       "input path",
				Destination: &c.InputPath,
			},
			&urcli.StringFlag{
				Name:        "outputPath",
				Value:       "output",
				Usage:       "output path",
				Destination: &c.OutputPath,
			},
		},
		Usage: "fight the loneliness!",
		Action: func(*urcli.Context) error {
			execute()
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func execute() {
	c.Scheme = runtime.NewScheme()

	err := apps.AddToScheme(c.Scheme)
	if err != nil {
		fmt.Println("could not add apps to scheme: %w", err)
	}

	err = core.AddToScheme(c.Scheme)
	if err != nil {
		fmt.Println("could not add apps to scheme: %w", err)
	}

	err = keptnv1alpha2.AddToScheme(c.Scheme)
	if err != nil {
		fmt.Println("could not add keptn to scheme: %w", err)
	}

	filepath.Walk(c.InputPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf(err.Error())
		}
		if filepath.Ext(info.Name()) == ".yaml" || filepath.Ext(info.Name()) == ".yml" {
			err := processYaml(path)
			if err != nil {
				log.Fatalf(err.Error())
			}
		}
		return nil
	})

	for _, v := range appList {
		if _, err := os.Stat(c.OutputPath + "/app-" + v.Name + ".yaml"); err == nil {
			yamlFile, err := os.ReadFile(c.OutputPath + "/app-" + v.Name + ".yaml")
			if err != nil {
				panic("Unable to open file")
			}
			app := keptnv1alpha2.KeptnApp{}
			err = yaml.Unmarshal(yamlFile, &app)
			if err != nil {
				panic("Unable to unmarshal file")
			}

			version := semver.New(app.Spec.Version)
			if err != nil {
				panic("Unable to parse version")
			}
			switch c.VersionUpgradeMode {
			case "patch":
				version.BumpPatch()
			case "major":
				version.BumpMajor()
			case "minor":
				version.BumpMinor()
			}
			v.Spec.Version = version.String()
		} else if errors.Is(err, os.ErrNotExist) {
			version := semver.Version{
				Major:      0,
				Minor:      0,
				Patch:      1,
				PreRelease: "",
				Metadata:   "",
			}
			v.Spec.Version = version.String()
		}

		y := printers.YAMLPrinter{}
		newFile, _ := os.Create(c.OutputPath + "/app-" + v.Name + ".yaml")
		defer newFile.Close()
		err := y.PrintObj(&v, newFile)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func processYaml(file string) error {
	yamlFile, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	splitInput, err := splitYAML(yamlFile)
	if err != nil {
		return err
	}

	factory := serializer.NewCodecFactory(c.Scheme)
	decoder := factory.UniversalDeserializer()

	for _, input := range splitInput {
		obj, _, err := decoder.Decode([]byte(input), nil, nil)
		if err != nil {
			return err
		}

		var isWorkload = false
		var data keptnv1alpha2.KeptnWorkloadRef
		var app string

		switch obj.(type) {
		case *apps.Deployment, *apps.StatefulSet, *apps.DaemonSet:
			data, app, isWorkload = parseDeployment(obj)
		default:
			continue
		}

		if isWorkload {
			if application, ok := appList[app]; ok {
				application.Spec.Workloads = append(application.Spec.Workloads, data)
				appList[app] = application
			} else {
				appList[app] = keptnv1alpha2.KeptnApp{
					TypeMeta: metav1.TypeMeta{
						Kind:       "KeptnApp",
						APIVersion: "lifecycle.keptn.sh/v1alpha2",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: app,
					},
					Spec: keptnv1alpha2.KeptnAppSpec{
						Workloads: []keptnv1alpha2.KeptnWorkloadRef{
							data,
						},
					},
				}
			}
		}
	}
	return err
}

func splitYAML(resources []byte) ([][]byte, error) {

	dec := yaml.NewDecoder(bytes.NewReader(resources))

	var res [][]byte
	for {
		var value interface{}
		err := dec.Decode(&value)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		valueBytes, err := yaml.Marshal(value)
		if err != nil {
			return nil, err
		}
		res = append(res, valueBytes)
	}
	return res, nil
}

func parseDeployment(obj interface{}) (keptnv1alpha2.KeptnWorkloadRef, string, bool) {
	workload := ""
	gotWorkloadAnnotation := false
	version := ""
	application := ""
	gotVersionAnnotation := false
	gotAppAnnotation := false
	containerVersion := ""

	switch obj.(type) {
	case *apps.Deployment:
		deployment := obj.(*apps.Deployment)
		workload, gotWorkloadAnnotation = getLabelOrAnnotation(&deployment.Spec.Template.ObjectMeta, workloadAnnotation, k8sRecommendedWorkloadAnnotations)
		version, gotVersionAnnotation = getLabelOrAnnotation(&deployment.Spec.Template.ObjectMeta, versionAnnotation, k8sRecommendedVersionAnnotations)
		application, gotAppAnnotation = getLabelOrAnnotation(&deployment.Spec.Template.ObjectMeta, appAnnotation, k8sRecommendedAppAnnotations)
		containerVersion = calculateVersion(deployment.Spec.Template)

	case *apps.StatefulSet:
		statefulset := obj.(*apps.StatefulSet)
		workload, gotWorkloadAnnotation = getLabelOrAnnotation(&statefulset.Spec.Template.ObjectMeta, workloadAnnotation, k8sRecommendedWorkloadAnnotations)
		version, gotVersionAnnotation = getLabelOrAnnotation(&statefulset.Spec.Template.ObjectMeta, versionAnnotation, k8sRecommendedVersionAnnotations)
		application, gotAppAnnotation = getLabelOrAnnotation(&statefulset.Spec.Template.ObjectMeta, appAnnotation, k8sRecommendedAppAnnotations)
		containerVersion = calculateVersion(statefulset.Spec.Template)
	case *apps.DaemonSet:
		daemonset := obj.(*apps.DaemonSet)
		workload, gotWorkloadAnnotation = getLabelOrAnnotation(&daemonset.Spec.Template.ObjectMeta, workloadAnnotation, k8sRecommendedWorkloadAnnotations)
		version, gotVersionAnnotation = getLabelOrAnnotation(&daemonset.Spec.Template.ObjectMeta, versionAnnotation, k8sRecommendedVersionAnnotations)
		application, gotAppAnnotation = getLabelOrAnnotation(&daemonset.Spec.Template.ObjectMeta, appAnnotation, k8sRecommendedAppAnnotations)
		containerVersion = calculateVersion(daemonset.Spec.Template)
	}

	if !gotWorkloadAnnotation {
		fmt.Println("No workload annotation found")
		return keptnv1alpha2.KeptnWorkloadRef{}, "", false
	}

	if !gotVersionAnnotation {
		fmt.Println("No version annotation found, calculating version")
		version = containerVersion
	}

	if !gotAppAnnotation {
		fmt.Println("No app annotation found, assuming workload name as app name")
		application = workload
	}

	return keptnv1alpha2.KeptnWorkloadRef{
		Name:    workload,
		Version: version,
	}, application, true
}

func getLabelOrAnnotation(resource *metav1.ObjectMeta, primaryAnnotation string, secondaryAnnotation string) (string, bool) {
	if resource.Annotations[primaryAnnotation] != "" {
		return resource.Annotations[primaryAnnotation], true
	}

	if resource.Labels[primaryAnnotation] != "" {
		return resource.Labels[primaryAnnotation], true
	}

	if secondaryAnnotation == "" {
		return "", false
	}

	if resource.Annotations[secondaryAnnotation] != "" {
		return resource.Annotations[secondaryAnnotation], true
	}

	if resource.Labels[secondaryAnnotation] != "" {
		return resource.Labels[secondaryAnnotation], true
	}
	return "", false
}

func calculateVersion(pod core.PodTemplateSpec) string {
	name := ""

	if len(pod.Spec.Containers) == 1 {
		image := strings.Split(pod.Spec.Containers[0].Image, ":")
		if len(image) > 1 && image[1] != "" && image[1] != "latest" {
			return image[1]
		}
	}

	for _, item := range pod.Spec.Containers {
		name = name + item.Name + item.Image
		for _, e := range item.Env {
			name = name + e.Name + e.Value
		}
	}

	h := fnv.New32a()
	h.Write([]byte(name))
	return fmt.Sprint(h.Sum32())
}
