package controller

import (
	"reflect"
	"time"

	"github.com/golang/glog"
	tapi "github.com/k8sdb/elasticsearch/api"
	tcs "github.com/k8sdb/elasticsearch/client/clientset"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/cache"
	rest "k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/wait"
	"k8s.io/kubernetes/pkg/watch"
)

type Controller struct {
	Client tcs.ExtensionInterface
	// sync time to sync the list.
	SyncPeriod time.Duration
}

func New(c *rest.Config) *Controller {
	return &Controller{
		Client:     tcs.NewExtensionsForConfigOrDie(c),
		SyncPeriod: time.Minute * 2,
	}
}

// Blocks caller. Intended to be called as a Go routine.
func (w *Controller) RunAndHold() {
	lw := &cache.ListWatch{
		ListFunc: func(opts api.ListOptions) (runtime.Object, error) {
			return w.Client.Certificate(api.NamespaceAll).List(api.ListOptions{})
		},
		WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
			return w.Client.Certificate(api.NamespaceAll).Watch(api.ListOptions{})
		},
	}
	_, controller := cache.NewInformer(lw,
		&tapi.Certificate{},
		w.SyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				glog.Infoln("Got one added tpr", obj.(*tapi.Certificate))
				w.doStuff(obj.(*tapi.Certificate))
			},
			DeleteFunc: func(obj interface{}) {
				glog.Infoln("Got one deleted tpr", obj.(*tapi.Certificate))
				w.doStuff(obj.(*tapi.Certificate))
			},
			UpdateFunc: func(old, new interface{}) {
				oldObj, ok := old.(*tapi.Certificate)
				if !ok {
					return
				}
				newObj, ok := new.(*tapi.Certificate)
				if !ok {
					return
				}
				if !reflect.DeepEqual(oldObj.Spec, newObj.Spec) {
					glog.Infoln("Got one updated tpr", newObj)
					w.doStuff(newObj)
				}
			},
		},
	)
	controller.Run(wait.NeverStop)
}

func (pl *Controller) doStuff(release *tapi.Certificate) {

}
