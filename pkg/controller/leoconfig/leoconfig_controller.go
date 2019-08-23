package leoconfig

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"

	"github.com/go-acme/lego/certcrypto"
	"github.com/go-acme/lego/lego"
	"github.com/go-acme/lego/registration"

	operatorv1alpha1 "github.com/webner/leo-operator/pkg/apis/operator/v1alpha1"
	"github.com/webner/leo-operator/pkg/controller/acme"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_leoconfig")

// Add creates a new LeoConfig Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileLeoConfig{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("leoconfig-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource LeoConfig
	err = c.Watch(&source.Kind{Type: &operatorv1alpha1.LeoConfig{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner LeoConfig
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.LeoConfig{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileLeoConfig implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileLeoConfig{}

// ReconcileLeoConfig reconciles a LeoConfig object
type ReconcileLeoConfig struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a LeoConfig object and makes changes based on the state read
// and what is in the LeoConfig.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileLeoConfig) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	namespace := request.Namespace
	reqLogger := log.WithValues("Request.Namespace", namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling LegoConfig")

	// Fetch the LegoConfig instance
	instance := &operatorv1alpha1.LeoConfig{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Check acme account
	accountSecretName := "acme-account-" + request.Name

	accountSecret := &corev1.Secret{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: accountSecretName, Namespace: namespace}, accountSecret)
	if err != nil && errors.IsNotFound(err) {

		// Create a user. New accounts need an email and private key to start.
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return reconcile.Result{}, err
		}

		user := acme.User{
			Email: instance.Spec.Account.Email,
			Key:   privateKey,
		}

		config := lego.NewConfig(&user)

		if instance.Spec.Production {
			config.CADirURL = lego.LEDirectoryProduction
		} else {
			config.CADirURL = lego.LEDirectoryStaging
		}

		config.Certificate.KeyType = certcrypto.RSA2048

		// A client facilitates communication with the CA server.
		client, err := lego.NewClient(config)
		if err != nil {
			return reconcile.Result{}, err
		}

		// New users will need to register
		reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			return reconcile.Result{}, err
		}

		keyPemEncoded := certcrypto.PEMEncode(privateKey)
		regEncoded, err := json.Marshal(reg)

		accountSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      accountSecretName,
				Namespace: namespace,
			},
			StringData: map[string]string{
				"email":        user.Email,
				"key":          string(keyPemEncoded),
				"registration": string(regEncoded),
			},
		}

		if err := controllerutil.SetControllerReference(instance, accountSecret, r.scheme); err != nil {
			return reconcile.Result{}, err
		}

		err = r.client.Create(context.TODO(), accountSecret)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}
