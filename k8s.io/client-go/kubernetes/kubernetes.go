package main

import (
	"context"
	"fmt"
	"net/http"

	"flag"

	"github.com/whatap/go-api/instrumentation/github.com/gin-gonic/gin/whatapgin"
	"github.com/whatap/go-api/instrumentation/k8s.io/client-go/kubernetes/whatapkubernetes"
	"github.com/whatap/go-api/trace"

	// "fmt"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/gin-gonic/gin"
)

var kubeconfig *string

func InitKubeConfig() {
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "~/.kube", "absolute path to the kubeconfig file")
	}
	flag.Parse()
}

func int32Ptr(i int32) *int32 { return &i }

func GetKubeClientSet(ctx context.Context) (*kubernetes.Clientset, error) {
	if ctx == nil {
		ctx = context.TODO()
	}
	cfg, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}
	// set whatap roundTripper
	cfg.WrapTransport = whatapkubernetes.WrapRoundTripper()

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func GetConfigMaps(ctx context.Context, namesplace string, name string) (*apiv1.ConfigMap, error) {
	if ctx == nil {
		ctx = context.TODO()
	}

	clientset, err := GetKubeClientSet(ctx)
	if err != nil {
		return nil, err
	}

	cm, err := clientset.CoreV1().ConfigMaps(namesplace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func GetPodsList(ctx context.Context, namespace string) (*apiv1.PodList, error) {
	if ctx == nil {
		ctx = context.TODO()
	}

	clientset, err := GetKubeClientSet(ctx)
	if err != nil {
		return nil, err
	}
	// 파드를 나열하기 위해 API에 접근한다
	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return pods, nil
}

func GetDeploymentsList(ctx context.Context, namespace string) (*appsv1.DeploymentList, error) {
	if ctx == nil {
		ctx = context.TODO()
	}
	clientset, err := GetKubeClientSet(ctx)
	if err != nil {
		return nil, err
	}

	list, err := clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list, err
}

func main() {

	portPtr := flag.Int("p", 8080, "web port. default 8080")
	udpPortPtr := flag.Int("up", 6600, "agent port(udp). defalt 6600")
	dataSourcePtr := flag.String("ds", "whatap:whatap1234!@tcp(localhost:3306)/whatap_demo", "dataSourceName")
	flag.Parse()

	InitKubeConfig()

	port, udpPort, _ := *portPtr, *udpPortPtr, *dataSourcePtr
	// port, udpPort, dataSource := *portPtr, *udpPortPtr, *dataSourcePtr

	// Whatap go
	config := make(map[string]string)
	config["net_udp_port"] = "127.0.0.1"
	config["net_udp_port"] = fmt.Sprintf("%d", udpPort)
	trace.Init(config)
	defer trace.Shutdown()

	r := gin.Default()
	r.Use(whatapgin.Middleware())

	r.LoadHTMLGlob("templates/k8s.io/client-go/kubernetes/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Title":   "k8s.io/client-go/kubernetes",
			"Content": c.Request.RequestURI,
		},
		)
	})

	r.GET("/configmap/:namespace/:name", func(c *gin.Context) {
		ctx := c.Request.Context()
		namespace := c.Param("namespace")
		name := c.Param("name")

		if cm, err := GetConfigMaps(ctx, namespace, name); err == nil {
			c.JSON(200, cm)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"msg":  "HTTP 500 O.K",
				"code": 500,
				"data": err.Error(),
			})
		}
	})

	r.GET("/pods/:namespace", func(c *gin.Context) {
		ctx := c.Request.Context()
		namespace := c.Param("namespace")
		if pods, err := GetPodsList(ctx, namespace); err == nil {
			c.JSON(200, pods)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"msg":  "HTTP 500 O.K",
				"code": 500,
				"data": err.Error(),
			})
		}
	})

	r.GET("/deployments/:namespace", func(c *gin.Context) {
		ctx := c.Request.Context()
		namespace := c.Param("namespace")
		if list, err := GetDeploymentsList(ctx, namespace); err == nil {
			c.JSON(200, list)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"msg":  "HTTP 500 O.K",
				"code": 500,
				"data": err.Error(),
			})
		}
	})
	fmt.Println("Start kubernetes :", port, ", Agent Udp Port:", udpPort)
	r.Run(fmt.Sprintf(":%d", port))
}
