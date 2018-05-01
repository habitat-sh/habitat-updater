package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	habclient "github.com/habitat-sh/habitat-operator/pkg/client/clientset/versioned/typed/habitat/v1beta1"
	"github.com/jasonlvhit/gocron"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type habPkg struct {
	Origin     string
	Name       string
	Version    string
	Release    string
	Deployment string
	Channel    string
}

type Service struct {
	Channel string `json:"channel"`
	Pkg     Pkg    `json:"pkg"`
}
type Pkg struct {
	Ident   string `json:"ident"`
	Origin  string `json:"origin"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Release string `json:"release"`
}

type BldrChannel struct {
	Ident Ident
}

type Ident struct {
	Origin  string `json:"origin"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Release string `json:"release"`
}

func main() {
	gocron.Every(1).Minute().Do(checker)
	<-gocron.Start()
}

func checker() {
	fmt.Println("Polling for updates...\n")
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	habclient := habclient.NewForConfigOrDie(config)

	pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{
		LabelSelector: "habitat=true",
	})
	if err != nil {
		panic(err.Error())
	}

	services := make(map[string]habPkg)

	for _, pod := range pods.Items {
		for k, v := range fetchSupInfo(pod.Status.PodIP, pod.GetLabels()["habitat-name"], services) {
			services[k] = v
		}
	}

	fmt.Printf("%d services loaded\n", len(services))

	for _, v := range services {
		client := &http.Client{}
		var bldrResp = BldrChannel{}
		req, err := http.NewRequest("GET", fmt.Sprintf("https://bldr.habitat.sh/v1/depot/channels/%s/%s/pkgs/%s/latest", v.Origin, v.Channel, v.Name), nil)
		if err != nil {
			panic(err.Error())
		}
		req.Header.Set("User-Agent", "Kubernetes-Updater-9000")
		resp, err := client.Do(req)
		if err != nil {
			panic(err.Error())
		}
		responseData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		json.Unmarshal(responseData, &bldrResp)
		if bldrResp.Ident.Release != "" && v.Release != "" {
			bldrRelease, err := strconv.Atoi(bldrResp.Ident.Release)
			if err != nil {
				panic(err.Error())
			}
			svcRelease, err := strconv.Atoi(v.Release)
			if err != nil {
				panic(err.Error())
			}
			if bldrRelease > svcRelease {
				fmt.Printf("Newer version of %s available\n", v.Name)
				updateDeploymentImage(habclient, v.Deployment, bldrResp.Ident)
			} else {
				fmt.Printf("Latest version of %s installed\n", v.Name)
			}
		} else {
			fmt.Printf("Release info empty for %s, skipping\n", v.Name)
		}
	}
}

func fetchSupInfo(ip string, deployment string, services map[string]habPkg) map[string]habPkg {
	var supResp []Service
	resp, err := http.Get(fmt.Sprintf("http://%s:9631/services", ip))
	if err != nil {
		panic(err.Error())
	}
	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}
	json.Unmarshal(responseData, &supResp)
	for _, svc := range supResp {
		services[svc.Pkg.Ident] = habPkg{
			Origin:     svc.Pkg.Origin,
			Name:       svc.Pkg.Name,
			Version:    svc.Pkg.Version,
			Release:    svc.Pkg.Release,
			Deployment: deployment,
			Channel:    svc.Channel,
		}
	}
	return services
}

func updateDeploymentImage(client *habclient.HabitatV1beta1Client, deployment string, newMetadata Ident) {
	service, err := client.Habitats(v1.NamespaceDefault).Get(deployment, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}
	dockerIdent := strings.Split(service.Spec.Image, ":")
	dockerIdentBeta2 := strings.Split(service.Spec.V1beta2.Image, ":")
	service.Spec.Image = fmt.Sprintf("%s:%s-%s", dockerIdent[0], newMetadata.Version, newMetadata.Release)
	service.Spec.V1beta2.Image = fmt.Sprintf("%s:%s-%s", dockerIdentBeta2[0], newMetadata.Version, newMetadata.Release)

	_, err = client.Habitats(v1.NamespaceDefault).Update(service)
	if err != nil {
		panic(err.Error())
	}
}
