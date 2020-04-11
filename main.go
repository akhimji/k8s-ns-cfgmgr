package main

import (
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/alyarctiq/k8s-cfgmgr/cmd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {

	// Clean Up
	master := "/tmp/master.yaml"
	_, err := os.Stat(master)
	if os.IsNotExist(err) {
		os.Remove(master)
	}

	var namespace string
	namespace = os.Getenv("NAMESPACE")
	if namespace == "" {
		log.Fatalln("NAMESPACE ENV Var Not Set")
		os.Exit(1)
	}
	log.Println("OS ENV NS: ", namespace)
	var url string
	var directory string
	var basefolder string
	var folder string
	var path string

	basefolder = "/tmp/repo"
	url = os.Getenv("URL")
	if url == "" {
		log.Fatalln("URL ENV Var Not Set")
		os.Exit(1)
	}
	directory = "/tmp/repo"
	log.Println("OS ENV URL: ", url)
	log.Println("Cloning Git Repo")
	cmd.GitClone(url, directory)

	folder = os.Getenv("FOLDER")
	if folder == "." {
		path = basefolder
	} else {
		path = basefolder + folder
	}
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		log.Println("Path Not Found:", path)
		os.Exit(1)
	} else {
		log.Println("Search Path: Found", path)
	}

	clientset, _ := cmd.BuildClient()

	cmd.GetNamespaces(clientset, namespace)

	log.Println("Starting Watch Loop...")
	for {
		//log.Println("Repeat Loop...")
		var yamlDeployments []string
		var yamlServices []string
		var currentDeployments []string
		var currentServices []string
		var data []byte
		var err error
		//var alldata []byte

		files, _ := cmd.WalkMatch(path, "*.yaml")

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
		//log.Println("Strip Comments...")

		// strip out comments from yaml file //
		re := regexp.MustCompile("(?m)[\r\n]+^.*#.*$")
		res := re.ReplaceAllString(string(data), "")
		data = []byte(res)
		// --- //

		yamlDeployments, yamlServices = cmd.ParseK8sYaml(data)

		deploymentsClient := clientset.AppsV1().Deployments(namespace)
		deployments, _ := deploymentsClient.List(metav1.ListOptions{})
		services, _ := clientset.CoreV1().Services(namespace).List(metav1.ListOptions{})

		for _, d := range deployments.Items {
			a := cmd.GetDeployment(namespace, d.Name, clientset)
			currentDeployments = append(currentDeployments, a.GetName())
			//log.Println("Current Deployment: ", d.GetName())
			//fmt.Printf(" * %s (%d replicas)\n", d.Name, *d.Spec.Replicas)
		}

		for _, services := range services.Items {
			//log.Println("Current Service: ", services.GetName())
			currentServices = append(currentServices, services.GetName())
		}

		//debugging//
		// log.Println("yaml dep parsed:", yamlDeployments)
		// log.Println("current dep:", currentDeployments)
		// log.Println("yaml svc parsed:", yamlServices)
		// log.Println("current svc:", currentServices)

		repairDep := cmd.SliceDiff(yamlDeployments, currentDeployments)
		repairSvc := cmd.SliceDiff(yamlServices, currentServices)

		// alpha := []string{"apples", "oranges", "pears", "plumbs"}
		// //beta := []string{"apples", "oranges", "pears"}
		// log.Println("here....", sliceDiff(alpha, yamlDeployments))
		// time.Sleep(10 * time.Second)

		//log.Println("Dep Delta:", repairDep)
		//log.Println("Svc Delta:", repairSvc)

		if len(repairDep) == 0 {
			repairDep := cmd.SliceDiff(currentDeployments, yamlDeployments)
			for _, name := range repairDep {
				_, found := cmd.FindInSlice(yamlDeployments, name)
				if !found {
					log.Println("Delete Deployment From Cluster:", name)
					cmd.DeleteDeployment(clientset, name, namespace)
					time.Sleep(2 * time.Second)
				} else {
					log.Println("Maintain Deployment From Cluster:", name)
				}
			}
		}
		if len(repairSvc) == 0 {
			repairSvc := cmd.SliceDiff(currentServices, yamlServices)
			for _, name := range repairSvc {
				_, found := cmd.FindInSlice(yamlServices, name)
				if !found {
					cmd.DeleteSvc(clientset, name, namespace)
					log.Println("Delete Svc From Cluster:", name)
				} else {
					log.Println("Maintain Svc From Cluster:", name)
				}
			}
		}
		//log.Println("Dep Delta:", repairDep)
		//log.Println("Svc Delta:", repairSvc)

		for _, name := range repairDep {
			cmd.Repairdeployment(data, name, namespace, clientset)
			//log.Println(name)
		}
		for _, name := range repairSvc {
			cmd.Repairservice(data, name, namespace, clientset)
			//log.Println(name)
		}
		time.Sleep(10 * time.Second)
		//log.Println("Checking for Repo Update")
		cmd.GitPull(directory)
		err = os.Remove(master)
		cmd.Check(err)

	}
}
