package cache

import (
	"AstroKube/pkg/client/clientset/versioned"
	astroInformer "AstroKube/pkg/client/informers/externalversions"
	astrov1 "AstroKube/pkg/client/listers/core/v1"
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appsv1 "k8s.io/client-go/listers/apps/v1"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"time"
)

var (
	CoreClientCache                *clientCache
	SubClientCache                 *clientCache
	CorePodInformerFactory         informers.SharedInformerFactory
	CoreSharedInformerFactory      informers.SharedInformerFactory
	CoreAstroSharedInformerFactory astroInformer.SharedInformerFactory
	SubSharedInformerFactory       informers.SharedInformerFactory
	SubAstroSharedInformerFactory  astroInformer.SharedInformerFactory
)

type clientCache struct {
	Config               *rest.Config
	Client               kubernetes.Interface
	AstroClient          versioned.Interface
	NodeLister           v1.NodeLister
	ClusterLister        astrov1.ClusterLister
	NamespaceLister      v1.NamespaceLister
	PodLister            v1.PodLister
	CmLister             v1.ConfigMapLister
	SecretLister         v1.SecretLister
	ServiceLister        v1.ServiceLister
	StatefulSetLister    appsv1.StatefulSetLister
	ReplicaSetLister     appsv1.ReplicaSetLister
	DeploymentLister     appsv1.DeploymentLister
	DaemonSetLister      appsv1.DaemonSetLister
	PvcLister            v1.PersistentVolumeClaimLister
	PvLister             v1.PersistentVolumeLister
	ServiceAccountLister v1.ServiceAccountLister
	ManagePodLister      v1.PodLister
	EventLister          v1.EventLister
	EventRecorder        record.EventRecorder
	ResyncPeriod         time.Duration
}

func InitClientCache(ctx context.Context, subClient, coreClient kubernetes.Interface, subAsClient, coreAsClient versioned.Interface, informerResyncPeriod time.Duration,
	clusterName string, coreConfig, subConfig *rest.Config) {
	subEventBroadcaster := record.NewBroadcaster()
	subEventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: subClient.CoreV1().Events("")})
	subEventRecorder := subEventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "astrolet"})

	coreEventBroadcaster := record.NewBroadcaster()
	coreEventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{
		Interface: coreClient.CoreV1().Events(""),
	})
	coreEventRecorder := coreEventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "astrolet"})

	SubSharedInformerFactory = informers.NewSharedInformerFactory(subClient, informerResyncPeriod)
	CoreSharedInformerFactory = informers.NewSharedInformerFactory(coreClient, informerResyncPeriod)
	CorePodInformerFactory = informers.NewSharedInformerFactoryWithOptions(coreClient, informerResyncPeriod, informers.WithTweakListOptions(func(options *metav1.ListOptions) {
		/*options.FieldSelector = fields.OneTermEqualSelector("spec.location.clusterName", clusterName).String()*/
	}))
	SubAstroSharedInformerFactory = astroInformer.NewSharedInformerFactory(subAsClient, informerResyncPeriod)
	CoreAstroSharedInformerFactory = astroInformer.NewSharedInformerFactory(coreAsClient, informerResyncPeriod)
	klog.Info("starting sharedInformerFactory informers")
	subPodInformer := SubSharedInformerFactory.Core().V1().Pods().Informer()
	//subNodeInformer := SubSharedInformerFactory.Core().V1().Nodes().Informer()
	//subSvcInformer := SubSharedInformerFactory.Core().V1().Services().Informer()
	////subSecretInformer := SubSharedInformerFactory.Core().V1().Secrets().Informer()
	//subConfigMapInformer := SubSharedInformerFactory.Core().V1().ConfigMaps().Informer()
	//subIngressInformer := SubSharedInformerFactory.Networking().V1().Ingresses().Informer()
	//subIngressClassInformer := SubSharedInformerFactory.Networking().V1().IngressClasses().Informer()
	//subNsInformer := SubSharedInformerFactory.Core().V1().Namespaces().Informer()
	//subStatefulSetInformer := SubSharedInformerFactory.Apps().V1().StatefulSets().Informer()
	//subReplicaSetInformer := SubSharedInformerFactory.Apps().V1().ReplicaSets().Informer()
	//subDeploymentInformer := SubSharedInformerFactory.Apps().V1().Deployments().Informer()
	//subDaemonsetInformer := SubSharedInformerFactory.Apps().V1().DaemonSets().Informer()
	//subEventInformer := SubSharedInformerFactory.Core().V1().Events().Informer()
	//subPvcInformer := SubSharedInformerFactory.Core().V1().PersistentVolumeClaims().Informer()
	//subPvInformer := SubSharedInformerFactory.Core().V1().PersistentVolumes().Informer()
	//subServiceAccountInformer := SubSharedInformerFactory.Core().V1().ServiceAccounts().Informer()
	subClusterInformer := SubAstroSharedInformerFactory.Astro().V1().Clusters().Informer()

	corePodInformer := CorePodInformerFactory.Core().V1().Pods().Informer()
	//coreNodeInformer := CoreSharedInformerFactory.Core().V1().Nodes().Informer()
	//coreSvcInformer := CoreSharedInformerFactory.Core().V1().Services().Informer()
	//coreSecretInformer := CoreSharedInformerFactory.Core().V1().Secrets().Informer()
	//coreStatefulSetInformer := CoreSharedInformerFactory.Apps().V1().StatefulSets().Informer()
	//coreCmInformer := CoreSharedInformerFactory.Core().V1().ConfigMaps().Informer()
	//coreIngressInformer := CoreSharedInformerFactory.Networking().V1().Ingresses().Informer()
	//coreIngressClassInformer := CoreSharedInformerFactory.Networking().V1().IngressClasses().Informer()
	//coreNsInformer := CoreSharedInformerFactory.Core().V1().Namespaces().Informer()
	//corePvcInformer := CoreSharedInformerFactory.Core().V1().PersistentVolumeClaims().Informer()
	//corePvInformer := CoreSharedInformerFactory.Core().V1().PersistentVolumes().Informer()
	//coreServiceAccountInformer := CoreSharedInformerFactory.Core().V1().ServiceAccounts().Informer()
	coreClusterInformer := CoreAstroSharedInformerFactory.Astro().V1().Clusters().Informer()

	go SubSharedInformerFactory.Start(ctx.Done())
	go CoreSharedInformerFactory.Start(ctx.Done())
	go CorePodInformerFactory.Start(ctx.Done())
	go CoreAstroSharedInformerFactory.Start(ctx.Done())
	go SubAstroSharedInformerFactory.Start(ctx.Done())

	if !cache.WaitForCacheSync(ctx.Done(),
		subPodInformer.HasSynced,
		//subNodeInformer.HasSynced,
		//subSvcInformer.HasSynced,
		////subSecretInformer.HasSynced,
		//subConfigMapInformer.HasSynced,
		//subIngressInformer.HasSynced,
		//subIngressClassInformer.HasSynced,
		//subNsInformer.HasSynced,
		//subStatefulSetInformer.HasSynced,
		//subReplicaSetInformer.HasSynced,
		//subDeploymentInformer.HasSynced,
		//subDaemonsetInformer.HasSynced,
		//subEventInformer.HasSynced,
		//subPvcInformer.HasSynced,
		//subPvInformer.HasSynced,
		//subServiceAccountInformer.HasSynced,
		subClusterInformer.HasSynced,
	) {
		runtime.HandleError(fmt.Errorf("WaitForSubClusterCacheSync failed"))
		klog.Fatal("WaitForPlanetCacheSync failed")
		return
	}
	klog.Info("finish sub cache sync ")
	if !cache.WaitForCacheSync(ctx.Done(),
		corePodInformer.HasSynced,
		//coreNodeInformer.HasSynced,
		//coreSvcInformer.HasSynced,
		//coreSecretInformer.HasSynced,
		//coreStatefulSetInformer.HasSynced,
		//coreCmInformer.HasSynced,
		//coreIngressInformer.HasSynced,
		//coreIngressClassInformer.HasSynced,
		//coreNsInformer.HasSynced,
		//corePvcInformer.HasSynced,
		//corePvInformer.HasSynced,
		//coreServiceAccountInformer.HasSynced,
		coreClusterInformer.HasSynced,
	) {
		runtime.HandleError(fmt.Errorf("WaitForCoreClusterCacheSync failed"))
		klog.Fatal("WaitForCoreClusterCacheSync failed")
		return
	}

	klog.Info("finish core cache sync ")

	CoreClientCache = &clientCache{
		Config:      coreConfig,
		Client:      coreClient,
		AstroClient: coreAsClient,
		//NodeLister:      CoreSharedInformerFactory.Core().V1().Nodes().Lister(),
		//NamespaceLister: CoreSharedInformerFactory.Core().V1().Namespaces().Lister(),
		PodLister: CorePodInformerFactory.Core().V1().Pods().Lister(),
		//CmLister:        CoreSharedInformerFactory.Core().V1().ConfigMaps().Lister(),
		//SecretLister:    CoreSharedInformerFactory.Core().V1().Secrets().Lister(),
		//ServiceLister:   CoreSharedInformerFactory.Core().V1().Services().Lister(),
		//PvcLister:       CoreSharedInformerFactory.Core().V1().PersistentVolumeClaims().Lister(),
		//PvLister:        CoreSharedInformerFactory.Core().V1().PersistentVolumes().Lister(),
		ClusterLister: CoreAstroSharedInformerFactory.Astro().V1().Clusters().Lister(),
		EventRecorder: coreEventRecorder,
	}

	SubClientCache = &clientCache{
		Config:      subConfig,
		Client:      subClient,
		AstroClient: subAsClient,
		//NodeLister:      SubSharedInformerFactory.Core().V1().Nodes().Lister(),
		//NamespaceLister: SubSharedInformerFactory.Core().V1().Namespaces().Lister(),
		PodLister: SubSharedInformerFactory.Core().V1().Pods().Lister(),
		//CmLister:        SubSharedInformerFactory.Core().V1().ConfigMaps().Lister(),
		////SecretLister:         SubSharedInformerFactory.Core().V1().Secrets().Lister(),
		//PvcLister:            SubSharedInformerFactory.Core().V1().PersistentVolumeClaims().Lister(),
		//ServiceLister:        SubSharedInformerFactory.Core().V1().Services().Lister(),
		//StatefulSetLister:    SubSharedInformerFactory.Apps().V1().StatefulSets().Lister(),
		//ReplicaSetLister:     SubSharedInformerFactory.Apps().V1().ReplicaSets().Lister(),
		//DeploymentLister:     SubSharedInformerFactory.Apps().V1().Deployments().Lister(),
		//DaemonSetLister:      SubSharedInformerFactory.Apps().V1().DaemonSets().Lister(),
		//EventLister:          SubSharedInformerFactory.Core().V1().Events().Lister(),
		//PvLister:             SubSharedInformerFactory.Core().V1().PersistentVolumes().Lister(),
		//ServiceAccountLister: SubSharedInformerFactory.Core().V1().ServiceAccounts().Lister(),
		ClusterLister: SubAstroSharedInformerFactory.Astro().V1().Clusters().Lister(),
		EventRecorder: subEventRecorder,
		ResyncPeriod:  informerResyncPeriod,
	}

}
