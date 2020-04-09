package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"time"

	. "gopkg.in/src-d/go-git.v4/_examples"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {

	var namespace string
	namespace = "hipster-cli"
	var url string
	var directory string
	master := "/tmp/master.yaml"
	_, err := os.Stat(master)
	if os.IsNotExist(err) {
		os.Remove(master)
	}

	url = "https://github.com/alyarctiq/go-git"
	//url = "https://github.com/GoogleCloudPlatform/microservices-demo.git"
	directory = "/tmp/repo"

	log.Println("Cloning Git Repo")
	gitClone(url, directory)

	clientset, _ := buildClient()

	getNamespaces(clientset, namespace)

	log.Println("Starting Watch Loop...")
	for {
		log.Println("Repeat Loop...")
		var yamlDeployments []string
		var yamlServices []string
		var currentDeployments []string
		var currentServices []string
		var data []byte
		var err error
		//var alldata []byte

		files, _ := walkMatch("/tmp/repo/", "*.yaml")

		for _, name := range files {
			log.Println("Loading Yaml Files:", name)
			data, err = ioutil.ReadFile(name)
			if err != nil {
				log.Println("File reading error", err)
				os.Exit(1)
			}
			// If the file doesn't exist, create it, or append to the file
			f, err := os.OpenFile("/tmp/master.yaml", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatal(err)
			}
			if _, err := f.Write(data); err != nil {
				log.Fatal(err)
			}
			if err := f.Close(); err != nil {
				log.Fatal(err)
			}

		}

		log.Println("Loading Master Files: ", master)
		data, err = ioutil.ReadFile(master)
		log.Println("Strip Comments...")

		// strip out comments from yaml file //
		re := regexp.MustCompile("(?m)[\r\n]+^.*#.*$")
		res := re.ReplaceAllString(string(data), "")
		data = []byte(res)
		// --- //

		yamlDeployments, yamlServices = parseK8sYaml(data)

		deploymentsClient := clientset.AppsV1().Deployments(namespace)
		deployments, _ := deploymentsClient.List(metav1.ListOptions{})
		services, _ := clientset.CoreV1().Services(namespace).List(metav1.ListOptions{})

		for _, d := range deployments.Items {
			a := getDeployment(namespace, d.Name, clientset)
			currentDeployments = append(currentDeployments, a.GetName())
			//fmt.Println("Current Deployment: ", d.GetName())
			//fmt.Printf(" * %s (%d replicas)\n", d.Name, *d.Spec.Replicas)
		}

		for _, services := range services.Items {
			//fmt.Println("Current Service: ", services.GetName())
			currentServices = append(currentServices, services.GetName())
		}

		//debugging//
		// fmt.Println("yaml dep parsed:", yamlDeployments)
		// fmt.Println("current dep:", currentDeployments)
		// fmt.Println("yaml svc parsed:", yamlServices)
		// fmt.Println("current svc:", currentServices)

		repairDep := sliceDiff(yamlDeployments, currentDeployments)
		repairSvc := sliceDiff(yamlServices, currentServices)

		// alpha := []string{"apples", "oranges", "pears", "plumbs"}
		// //beta := []string{"apples", "oranges", "pears"}
		// fmt.Println("here....", sliceDiff(alpha, yamlDeployments))
		// time.Sleep(10 * time.Second)

		//fmt.Println("Dep Delta:", repairDep)
		//fmt.Println("Svc Delta:", repairSvc)

		if len(repairDep) == 0 {
			repairDep := sliceDiff(currentDeployments, yamlDeployments)
			for _, name := range repairDep {
				_, found := findInSlice(yamlDeployments, name)
				if !found {
					fmt.Println("Delete Dep From Cluster:", name)
					deleteDeployment(clientset, name, namespace)
					time.Sleep(2 * time.Second)
				} else {
					fmt.Println("Maintain Dep From Cluster:", name)
				}
			}
		}
		if len(repairSvc) == 0 {
			repairSvc := sliceDiff(currentServices, yamlServices)
			for _, name := range repairSvc {
				_, found := findInSlice(yamlServices, name)
				if !found {
					deleteSvc(clientset, name, namespace)
					fmt.Println("Delete Svc From Cluster:", name)
				} else {
					fmt.Println("Maintain Svc From Cluster:", name)
				}
			}
		}
		//fmt.Println("Dep Delta:", repairDep)
		//fmt.Println("Svc Delta:", repairSvc)

		for _, name := range repairDep {
			repairdeployment(data, name, namespace, clientset)
			//fmt.Println(name)
		}
		for _, name := range repairSvc {
			repairservice(data, name, namespace, clientset)
			//fmt.Println(name)
		}
		time.Sleep(10 * time.Second)
		log.Println("Checking for Repo Update")
		gitPull(directory)
		err = os.Remove(master)
		check(err)

	}
}
