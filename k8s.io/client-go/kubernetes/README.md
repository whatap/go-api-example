# k8s.io/client-go/kubernetes

It configures the WrapTransport function of restclient.Config through whatapkubernetes.WrapRoundTripper().

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
