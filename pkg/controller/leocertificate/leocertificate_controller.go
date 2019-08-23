package leocertificate

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"time"

	operatorv1alpha1 "github.com/webner/leo-operator/pkg/apis/operator/v1alpha1"
	"github.com/webner/leo-operator/pkg/controller/acme"
	mylog "github.com/webner/leo-operator/pkg/log"
	"github.com/webner/leo-operator/pkg/utils"
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

	"github.com/go-acme/lego/certcrypto"
	"github.com/go-acme/lego/certificate"
	"github.com/go-acme/lego/lego"
	legolog "github.com/go-acme/lego/log"
	"github.com/go-acme/lego/providers/dns/acmedns"
	"github.com/go-acme/lego/registration"

	"github.com/cpu/goacmedns"
)

var log = logf.Log.WithName("controller_leocertificate")

// Add creates a new LeoCertificate Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileLeoCertificate{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("leocertificate-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource LeoCertificate
	err = c.Watch(&source.Kind{Type: &operatorv1alpha1.LeoCertificate{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner LeoCertificate
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.LeoCertificate{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileLeoCertificate implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileLeoCertificate{}

// ReconcileLeoCertificate reconciles a LeoCertificate object
type ReconcileLeoCertificate struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

type secretStorage struct {
	namespace string
	instance  *operatorv1alpha1.LeoCertificate
	r         *ReconcileLeoCertificate
}

func (ss secretStorage) Save() error {
	return nil
}

func (ss secretStorage) Put(domain string, acct goacmedns.Account) error {

	credentialsName := "acme-dns-" + domain
	data, err := json.Marshal(acct)
	if err != nil {
		return err
	}

	credentials := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      credentialsName,
			Namespace: ss.namespace,
		},
		StringData: map[string]string{
			"account": string(data),
		},
	}

	if err := controllerutil.SetControllerReference(ss.instance, credentials, ss.r.scheme); err != nil {
		return err
	}

	return ss.r.client.Create(context.TODO(), credentials)
}

func (ss secretStorage) Fetch(domain string) (goacmedns.Account, error) {

	credentialsName := "acme-dns-" + domain

	credentials := &corev1.Secret{}
	err := ss.r.client.Get(context.TODO(), types.NamespacedName{Name: credentialsName, Namespace: ss.namespace}, credentials)
	if err != nil && errors.IsNotFound(err) {
		return goacmedns.Account{}, goacmedns.ErrDomainNotFound
	} else if err != nil {
		return goacmedns.Account{}, err
	}

	account := goacmedns.Account{}
	err = json.Unmarshal(credentials.Data["account"], &account)
	return account, err
}

func (r *ReconcileLeoCertificate) getAcmeUserFromSecret(namespace, secretName string) (acme.User, error) {
	if secretName == "" {
		secretName = "acme-account-default"
	}

	accountSecret := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: namespace}, accountSecret)
	if err != nil {
		return acme.User{}, err
	}

	reg := registration.Resource{}
	err = json.Unmarshal(accountSecret.Data["registration"], &reg)
	if err != nil {
		return acme.User{}, err
	}

	block, _ := pem.Decode(accountSecret.Data["key"])
	x509Encoded := block.Bytes
	privateKey, _ := x509.ParseECPrivateKey(x509Encoded)

	return acme.User{
		Email:        string(accountSecret.Data["email"]),
		Key:          privateKey,
		Registration: &reg,
	}, nil
}

// Reconcile reads that state of the cluster for a LeoCertificate object and makes changes based on the state read
// and what is in the LeoCertificate.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileLeoCertificate) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	namespace := request.Namespace

	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling LeoCertificate")

	// Fetch the LetsEncryptCertificate instance
	instance := &operatorv1alpha1.LeoCertificate{}
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

	leoConfig := &operatorv1alpha1.LeoConfig{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "default", Namespace: request.Namespace}, leoConfig)
	if err != nil {
		if errors.IsNotFound(err) {
			// TODO: write errorMessage config not found
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	reqLogger.Info("get acme user form secret")
	acmeUser, err := r.getAcmeUserFromSecret(namespace, "")
	if err != nil {
		return reconcile.Result{}, err
	}

	config := lego.NewConfig(&acmeUser)
	legolog.Logger = mylog.StdLoggerAdapter{
		ReqLogger: reqLogger,
	}

	if leoConfig.Spec.Production {
		config.CADirURL = lego.LEDirectoryProduction
	} else {
		config.CADirURL = lego.LEDirectoryStaging
	}
	config.Certificate.KeyType = certcrypto.RSA2048

	reqLogger.Info("create lego client")
	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(config)
	if err != nil {
		return reconcile.Result{}, err
	}

	dnsClient := goacmedns.NewClient(utils.ValueOrDefault(leoConfig.Spec.Provider.AcmeDNS.URL, "https://auth.acme-dns.io"))
	storage := secretStorage{
		namespace: namespace,
		instance:  instance,
		r:         r,
	}

	provider, err := acmedns.NewDNSProviderClient(dnsClient, storage)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = client.Challenge.SetDNS01Provider(provider)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Check certificates
	certificateName := "acme-certificate-" + instance.Spec.Domain

	reqLogger.Info("get current certificate")
	certificateSecret := &corev1.Secret{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: certificateName, Namespace: request.Namespace}, certificateSecret)
	if err != nil && errors.IsNotFound(err) {

		reqLogger.Info("creating new certificate")

		certRequest := certificate.ObtainRequest{
			Domains: []string{"*." + instance.Spec.Domain},
			Bundle:  true,
		}

		certificates, err := client.Certificate.Obtain(certRequest)
		if err != nil {
			return reconcile.Result{Requeue: true, RequeueAfter: time.Minute * 2}, err
		}

		certificateSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      certificateName,
				Namespace: request.Namespace,
			},
			StringData: map[string]string{
				"tls.crt": string(certificates.Certificate),
				"tls.key": string(certificates.PrivateKey),
			},
		}
		if err := controllerutil.SetControllerReference(instance, certificateSecret, r.scheme); err != nil {
			return reconcile.Result{}, err
		}

		err = r.client.Create(context.TODO(), certificateSecret)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err == nil {

		reqLogger.Info("renew certificate")

		certRes := certificate.Resource{}
		certRes.Certificate = certificateSecret.Data["tls.crt"]
		certRes.PrivateKey = certificateSecret.Data["tls.key"]

		certificates, err := certcrypto.ParsePEMBundle(certRes.Certificate)
		if err != nil {
			return reconcile.Result{}, err
		}
		x509Cert := certificates[0]
		timeLeft := x509Cert.NotAfter.Sub(time.Now().UTC())

		if int(timeLeft.Hours()) < 30*24 {

			newCertificates, err := client.Certificate.Renew(certRes, true, false)
			if err != nil {
				return reconcile.Result{}, err
			}
			certificateSecret.StringData = map[string]string{
				"tls.crt": string(newCertificates.Certificate),
				"tls.key": string(newCertificates.PrivateKey),
			}
			err = r.client.Update(context.TODO(), certificateSecret)
			if err != nil {
				return reconcile.Result{}, err
			}
		} else {
			reqLogger.Info(fmt.Sprintf("Skipping renew, certificate is still valid for %d hours", int(timeLeft.Hours())))
		}
	}

	return reconcile.Result{Requeue: true, RequeueAfter: time.Hour * 24}, nil
}
