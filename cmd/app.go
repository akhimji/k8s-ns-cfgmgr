package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ghodss/yaml"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"

	"gopkg.in/src-d/go-git.v4"
	. "gopkg.in/src-d/go-git.v4/_examples"
)

func gitClone(url string, directory string) {
	// Clone the given repository to the given directory
	Info("git pull %s %s --recursive", url, directory)

	_, err := git.PlainClone(directory, false, &git.CloneOptions{
		URL:               url,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})

	CheckIfError(err)

	// ... retrieving the branch being pointed by HEAD
	//ref, err := r.Head()
	//CheckIfError(err)
	// ... retrieving the commit object
	//commit, err := r.CommitObject(ref.Hash())
	//CheckIfError(err)

	//fmt.Println(commit)
}

func gitPull(path string) {
	// We instance\iate a new repository targeting the given path (the .git folder)
	r, err := git.PlainOpen(path)
	CheckIfError(err)

	// Get the working directory for the repository
	w, err := r.Worktree()
	CheckIfError(err)

	// Pull the latest changes from the origin remote and merge into the current branch
	Info("git pull origin")
	_ = w.Pull(&git.PullOptions{RemoteName: "origin"})
	// CheckIfError(err)

	// Print the latest commit that was just pulled
	//ref, err := r.Head()
	//CheckIfError(err)
	//commit, err := r.CommitObject(ref.Hash())
	//CheckIfError(err)

	//fmt.Println(commit)
}

func deleteDeployment(clientset *kubernetes.Clientset, deployment string, ns string) {
	// build client set
	deploymentsClient := clientset.AppsV1().Deployments(ns)
	// build delete policy
	deletePolicy := metav1.DeletePropagationForeground
	// From Docs "PropagationPolicy    *DeletionPropagation"  in json format and *DeletionPropagation is pointer to metav1.DeletePropagationForeground
	//(&deletePolicy is pointer to deletePolicy)
	deleteOptions := metav1.DeleteOptions{PropagationPolicy: &deletePolicy}
	err := deploymentsClient.Delete(deployment, &deleteOptions)
	if err != nil {
		fmt.Println("Error Deleting Deployment:")
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Deleted deployment.")

}

func parseK8sYaml(fileR []byte) ([]string, []string) {
	var yamlDeployments []string
	var svcDeployments []string
	acceptedK8sTypes := regexp.MustCompile(`(Role|ClusterRole|RoleBinding|ClusterRoleBinding|ServiceAccount|Deployment|Service)`)
	fileAsString := string(fileR[:])
	sepYamlfiles := strings.Split(fileAsString, "---")
	//retVal := make([]runtime.Object, 0, len(sepYamlfiles))
	for _, f := range sepYamlfiles {
		if f == "\n" || f == "" {
			// ignore empty cases
			continue
		}

		decode := scheme.Codecs.UniversalDeserializer().Decode
		obj, groupVersionKind, err := decode([]byte(f), nil, nil)
		if err != nil {
			//log.Println(fmt.Sprintf("Error while decoding YAML object. Err was: %s", err))
			continue
		}

		if !acceptedK8sTypes.MatchString(groupVersionKind.Kind) {
			log.Printf("The custom-roles configMap contained K8s object types which are not supported! Skipping object with type: %s", groupVersionKind.Kind)
		} else {
			if groupVersionKind.Kind == "Deployment" {
				//fmt.Println("This is type Deployment")
				//b := []byte(f)
				//createDeploymentFromYaml(clientset, b, "hipster-cli")
				//fmt.Println(f)
				deployment := obj.(*appsv1.Deployment)
				//fmt.Println("Name:", deployment.GetName())
				yamlDeployments = append(yamlDeployments, deployment.GetName())

			}
			if groupVersionKind.Kind == "Service" {
				services := obj.(*v1.Service)
				//fmt.Println("Name:", deployment.GetName())
				svcDeployments = append(svcDeployments, services.GetName())
			}
		}

	}
	return yamlDeployments, svcDeployments
}

func repairdeployment(fileR []byte, repairdep string, namespace string, clientset *kubernetes.Clientset) {
	fileAsString := string(fileR[:])
	sepYamlfiles := strings.Split(fileAsString, "---")
	//retVal := make([]runtime.Object, 0, len(sepYamlfiles))
	for _, f := range sepYamlfiles {
		if f == "\n" || f == "" {
			// ignore empty cases
			continue
		}

		decode := scheme.Codecs.UniversalDeserializer().Decode
		obj, groupVersionKind, err := decode([]byte(f), nil, nil)
		if err != nil {
			//log.Println(fmt.Sprintf("Error while decoding YAML object. Err was: %s", err))
			continue
		}

		if groupVersionKind.Kind == "Deployment" {
			deployment := obj.(*appsv1.Deployment)
			if deployment.GetName() == repairdep {
				log.Println("Repairing Missing Deployment:", deployment.GetName())
				b := []byte(f)
				createDeploymentFromYaml(clientset, b, namespace)
			}

		}

	}
}

func repairservice(fileR []byte, repairsv string, namespace string, clientset *kubernetes.Clientset) {
	fileAsString := string(fileR[:])
	sepYamlfiles := strings.Split(fileAsString, "---")
	//retVal := make([]runtime.Object, 0, len(sepYamlfiles))
	for _, f := range sepYamlfiles {
		if f == "\n" || f == "" {
			// ignore empty cases
			continue
		}

		decode := scheme.Codecs.UniversalDeserializer().Decode
		obj, groupVersionKind, err := decode([]byte(f), nil, nil)
		if err != nil {
			//log.Println(fmt.Sprintf("Error while decoding YAML object. Err was: %s", err))
			continue
		}

		if groupVersionKind.Kind == "Service" {
			service := obj.(*v1.Service)
			if service.GetName() == repairsv {
				log.Println("Repairing Missing Service:", service.GetName())
				b := []byte(f)
				createServiceFromYaml(clientset, b, namespace)
			}
		}

	}
}

func buildClient() (*kubernetes.Clientset, error) {
	var kubeconfig string
	var cfgFile string
	if cfgFile != "" {
		kubeconfig = cfgFile
		log.Println(" ✓ Using kubeconfig file via flag: ", kubeconfig)
	} else {
		kubeconfig = os.Getenv("kubeconfig")
		if kubeconfig != "" {
			log.Println(" ✓ Using kubeconfig via OS ENV")
		} else {
			kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
			if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
				log.Println(" X kubeconfig Not Found, use --kubeconfig")
				os.Exit(1)
			} else {
				log.Println(" ✓ Using kubeconfig file via homedir: ", kubeconfig)
			}

		}

	}

	// Bootstrap k8s configuration
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	return clientset, err
}

func createServiceFromYaml(clientset *kubernetes.Clientset, podAsYaml []byte, ns string) {
	log.Println("Attempting Service Deployment..")
	var p v1.Service
	err := yaml.Unmarshal(podAsYaml, &p)
	if err != nil {
		log.Println("Error Unmarshalling: ", err)
	}
	service, err := clientset.CoreV1().Services(ns).Create(&p)
	if err != nil {
		log.Println("Error creating service: ", err)
	}
	fmt.Printf("Created Service %q.\n", service.GetObjectMeta().GetName())
}

func createDeploymentFromYaml(clientset *kubernetes.Clientset, podAsYaml []byte, ns string) error {
	log.Println("Attempting Deployment..")
	var deployment appsv1.Deployment
	err := yaml.Unmarshal(podAsYaml, &deployment)
	if err != nil {
		log.Println("Error Unmarshaling:", err)
	}

	deploymentsClient := clientset.AppsV1().Deployments(ns)
	result, err := deploymentsClient.Create(&deployment)
	//pod, poderr := clientset.CoreV1().Pods(ns).Create(&deployment)
	if err != nil {
		log.Println("Error Creating Deployment:")
		log.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
	return nil
}

func getDeployment(namespace string, name string, client *kubernetes.Clientset) *appsv1.Deployment {
	d, err := client.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil
	}

	return d
}

func sliceDiff(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [inputfile]\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

func getNamespaces(clientset *kubernetes.Clientset, ns string) {
	var ok string
	namespace, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		log.Fatalln("Failed to get namespace:", err)
		os.Exit(1)
	} else {
		for _, name := range namespace.Items {
			if name.GetName() == ns {
				ok = "found"
				break
			}
			ok = "not found"
		}
	}
	if ok == "not found" {
		log.Fatalln("Namespace Not Found..")
		os.Exit(1)
	}

}
func walkMatch(root, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func findInSlice(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func deleteSvc(clientset *kubernetes.Clientset, svcname string, ns string) {
	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{PropagationPolicy: &deletePolicy}
	if err := clientset.CoreV1().Services(ns).Delete(svcname, &deleteOptions); err != nil {
		fmt.Println("Error Failed to Delete Pod:")
		fmt.Println(err)
		os.Exit(1)
	} else {
		fmt.Println("Deleting Svc:", svcname)
	}
}
