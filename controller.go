/*
Copyright 2024 Forty Two Apps.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	kubelitedbv1 "github.com/fortytwoapps/kubelitedb/pkg/apis/kubelitedb/v1"
	clientset "github.com/fortytwoapps/kubelitedb/pkg/generated/clientset/versioned"
	kubelitedbscheme "github.com/fortytwoapps/kubelitedb/pkg/generated/clientset/versioned/scheme"
	informers "github.com/fortytwoapps/kubelitedb/pkg/generated/informers/externalversions/kubelitedb/v1"
	listers "github.com/fortytwoapps/kubelitedb/pkg/generated/listers/kubelitedb/v1"
)

const controllerAgentName = "kubelitedb-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a SQLiteInstance is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a SQLiteInstance fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by SQLiteInstance"
	// MessageResourceSynced is the message used for an Event fired when a SQLiteInstance
	// is synced successfully
	MessageResourceSynced = "SQLiteInstance synced successfully"
)

// Controller is the controller implementation for SQLiteInstance resources
type Controller struct {
	kubeclientset       kubernetes.Interface
	kubelitedbclientset clientset.Interface

	sqliteInstancesLister listers.SQLiteInstanceLister
	sqliteInstancesSynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface
	recorder  record.EventRecorder
}

// NewController returns a new KubeLiteDB controller
func NewController(
	ctx context.Context,
	kubeclientset kubernetes.Interface,
	kubelitedbclientset clientset.Interface,
	sqliteInstanceInformer informers.SQLiteInstanceInformer) *Controller {

	logger := klog.FromContext(ctx)

	// Create event broadcaster
	// Add kubelitedb types to the default Kubernetes Scheme so Events can be
	// logged for kubelitedb types.
	utilruntime.Must(kubelitedbscheme.AddToScheme(scheme.Scheme))
	logger.V(4).Info("Creating event broadcaster")

	eventBroadcaster := record.NewBroadcaster(record.WithContext(ctx))
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:       kubeclientset,
		kubelitedbclientset: kubelitedbclientset,

		sqliteInstancesLister: sqliteInstanceInformer.Lister(),
		sqliteInstancesSynced: sqliteInstanceInformer.Informer().HasSynced,
		workqueue:             workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "SQLiteInstances"),
		recorder:              recorder,
	}

	logger.Info("Setting up event handlers")
	// Set up an event handler for when SQLiteInstance resources change
	sqliteInstanceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueSQLiteInstance,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueSQLiteInstance(new)
		},
		DeleteFunc: controller.enqueueSQLiteInstance,
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(ctx context.Context, workers int) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()
	logger := klog.FromContext(ctx)

	// Start the informer factories to begin populating the informer caches
	logger.Info("Starting KubeLiteDB controller")

	// Wait for the caches to be synced before starting workers
	logger.Info("Waiting for informer caches to sync")

	if ok := cache.WaitForCacheSync(ctx.Done(), c.sqliteInstancesSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	logger.Info("Starting workers", "count", workers)
	// Launch workers to process SQLiteInstance resources
	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, c.runWorker, time.Second)
	}

	logger.Info("Started workers")
	<-ctx.Done()
	logger.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem(ctx context.Context) bool {
	obj, shutdown := c.workqueue.Get()
	logger := klog.FromContext(ctx)

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// SQLiteInstance resource to be synced.
		if err := c.syncHandler(ctx, key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		logger.Info("Successfully synced", "resourceName", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the SQLiteInstance resource
// with the current status of the resource.
func (c *Controller) syncHandler(ctx context.Context, key string) error {
	// logger := klog.LoggerWithValues(klog.FromContext(ctx), "resourceName", key)

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the SQLiteInstance resource with this namespace/name
	sqliteInstance, err := c.sqliteInstancesLister.SQLiteInstances(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("sqliteinstance '%s' in work queue no longer exists", key))
			return nil
		}
		return err
	}

	// TODO: Add sync logic here, e.g., create/update/delete related resources

	// Update the status block of the SQLiteInstance resource to reflect the
	// current state of the world
	err = c.updateSQLiteInstanceStatus(sqliteInstance)
	if err != nil {
		return err
	}

	c.recorder.Event(sqliteInstance, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) updateSQLiteInstanceStatus(sqliteInstance *kubelitedbv1.SQLiteInstance) error {
	sqliteInstanceCopy := sqliteInstance.DeepCopy()
	sqliteInstanceCopy.Status.Phase = "Running"
	// Update status fields here, e.g., sqliteInstanceCopy.Status.Phase = "Running"

	_, err := c.kubelitedbclientset.KubelitedbV1().SQLiteInstances(sqliteInstance.Namespace).UpdateStatus(context.TODO(), sqliteInstanceCopy, v1.UpdateOptions{})
	return err
}

// enqueueSQLiteInstance takes a SQLiteInstance resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than SQLiteInstance.
func (c *Controller) enqueueSQLiteInstance(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}
