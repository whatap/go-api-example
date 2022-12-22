#  k8s.io/client-go/kubernetes

restclient.Config 의 WrapTransport 함수를 whatapkubernetes.WrapRoundTripper()를 통해서 설정합니다. 
```
import(
	"github.com/whatap/go-api/instrumentation/k8s.io/client-go/kubernetes/whatapkubernetes"
)

func main() {
	
	... 
	
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

	...	
}
```

