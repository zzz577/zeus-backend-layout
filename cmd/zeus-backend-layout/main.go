package main

import (
	"flag"
	"fmt"
	"github.com/go-kratos/kratos/contrib/config/kubernetes/v2"
	kuberegistry "github.com/go-kratos/kratos/contrib/registry/kubernetes/v2"
	k8sclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
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
	restConfig, err := rest.InClusterConfig()
	home := homedir.HomeDir()
	if err != nil {
		kubeConfig := filepath.Join(home, ".kube", "config")
		restConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
		if err != nil {
			log.Fatal(err)
		}
	}
	clientSet, err := k8sclient.NewForConfig(restConfig)
	if err != nil {
		log.Fatal(err)
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
		kratos.Registrar(kuberegistry.NewRegistry(clientSet)),
	)
}

func main() {
	flag.Parse()
	// log
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05.000"
	zlogger := zerolog.New(os.Stdout)
	logger := log.With(kzerolog.NewLogger(&zlogger),
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
	c := config.New(
		config.WithSource(
			kubernetes.NewSource(
				kubernetes.Namespace(NameSpace),
				kubernetes.LabelSelector(fmt.Sprintf("app=%s", Name)),
				// kubernetes.KubeConfig(filepath.Join(homedir.HomeDir(), ".kube", "config")),
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
