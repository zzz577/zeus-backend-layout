package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/go-kratos/kratos/contrib/config/kubernetes/v2"
	kuberegistry "github.com/go-kratos/kratos/contrib/registry/kubernetes/v2"
	"github.com/go-kratos/kratos/v2/registry"
	k8sclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/utils/env"
	"os"
	"path/filepath"

	"zeus-backend-layout/internal/conf"

	kzerolog "github.com/go-kratos/kratos/contrib/log/zerolog/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/rs/zerolog"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	NameSpace = "zeus"
	// Name is the name of the compiled software.
	Name = "zeus-backend-layout"
	// Version is the version of the compiled software.
	Version string
	id, _   = os.Hostname()
)

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
	kubeConf, err := rest.InClusterConfig()
	switch {
	case errors.Is(err, rest.ErrNotInCluster):
		kubeConf, err = clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
		if err != nil {
			log.Fatal(err)
		}
	case err != nil:
		log.Fatal(err)
	}
	clientSet, err := k8sclient.NewForConfig(kubeConf)
	if err != nil {
		log.Fatal(err)
	}

	var reg registry.Registrar
	if debug, _ := env.GetBool("DEBUG", false); !debug {
		reg = kuberegistry.NewRegistry(clientSet)
	}
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
		kratos.Registrar(reg),
	)
}

func main() {
	flag.Parse()
	// log
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05.000"
	zl := zerolog.New(os.Stdout)
	logger := log.With(kzerolog.NewLogger(&zl),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)
	log.SetLogger(logger)

	// config
	var kubeConfig string
	_, err := rest.InClusterConfig()
	switch {
	case errors.Is(err, rest.ErrNotInCluster):
		kubeConfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
	case err != nil:
		log.Fatal(err)
	}
	c := config.New(
		config.WithSource(
			kubernetes.NewSource(
				kubernetes.Namespace(NameSpace),
				kubernetes.LabelSelector(fmt.Sprintf("app=%s", Name)),
				kubernetes.KubeConfig(kubeConfig),
			),
		),
	)
	defer func(c config.Config) {
		err := c.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(c)
	if err := c.Load(); err != nil {
		log.Fatal(err)
	}
	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		log.Fatal(err)
	}

	app, cleanup, err := wireApp(bc.Server, bc.Data, logger)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
