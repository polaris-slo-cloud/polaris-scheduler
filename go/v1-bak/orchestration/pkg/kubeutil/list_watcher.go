package kubeutil

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	_listWatcherImpl *listWatcherImpl

	_ ListWatcher = _listWatcherImpl
)

// ListWatcher allows watching a list of Kubernetes objects.
// It emits the entire list upon initialization and whenever an element is added, changed, or deleted.
type ListWatcher interface {

	// Gets the channel that emits the list of objects after init and on every change.
	WatchChan() chan client.ObjectList

	// Stops the watch.
	Stop()
}

type listWatcherImpl struct {
	client     client.WithWatch
	listType   client.ObjectList
	watch      watch.Interface
	outputChan chan client.ObjectList
}

// Creates a new ListWatcher for the specified listType and starts the watch.
func StartListWatcher(listType client.ObjectList, cfg *rest.Config, scheme *runtime.Scheme) (ListWatcher, error) {
	cl, err := client.NewWithWatch(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}

	watcher := &listWatcherImpl{
		client:     cl,
		listType:   listType.DeepCopyObject().(client.ObjectList),
		outputChan: make(chan client.ObjectList, 1),
	}
	if err := watcher.initWatch(); err != nil {
		return nil, err
	}
	return watcher, nil
}

func (me *listWatcherImpl) WatchChan() chan client.ObjectList {
	return me.outputChan
}

func (me *listWatcherImpl) Stop() {
	me.watch.Stop()
	close(me.outputChan)
}

func (me *listWatcherImpl) initWatch() error {
	// Fetch the initial list.
	initialList, err := me.getCurrentList()
	if err != nil {
		return err
	}

	// Initializes the watch object, starting from the current resource version.
	listType := me.listType.DeepCopyObject().(client.ObjectList)
	watch, err := me.client.Watch(context.TODO(), listType, &client.ListOptions{
		Raw: &metav1.ListOptions{ResourceVersion: initialList.GetResourceVersion()},
	})
	if err != nil {
		return err
	}
	me.watch = watch

	// Write the initial list to the output channel and run the watch loop.
	me.publishList(initialList)
	go me.runWatch()
	return nil
}

func (me *listWatcherImpl) runWatch() {
	resultsChan := me.watch.ResultChan()
	for event := range resultsChan {
		// ToDo: Implement some form of caching to avoid refetching the entire list every time.
		var _ = event
		list, err := me.getCurrentList()
		if err == nil {
			me.publishList(list)
		}
	}
}

func (me *listWatcherImpl) getCurrentList() (client.ObjectList, error) {
	list := me.listType.DeepCopyObject().(client.ObjectList)
	if err := me.client.List(context.TODO(), list); err != nil {
		return nil, err
	}
	return list, nil
}

func (me *listWatcherImpl) publishList(list client.ObjectList) {
	me.outputChan <- list
}
