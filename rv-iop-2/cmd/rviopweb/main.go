package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/Meesho/go-client/ab"
	"github.com/Meesho/go-client/ab/simplehttp"
	"github.com/Meesho/go-core/circuitbreaker"
	config3 "github.com/Meesho/go-core/config"
	"github.com/Meesho/go-core/httpframework"
	"github.com/Meesho/go-core/logger"
	"github.com/Meesho/go-core/metric"
	"github.com/Meesho/go-core/prismlogger"
	crossRE "github.com/Meesho/iop-component/cross-re"
	configAd "github.com/Meesho/iop-component/rv-substitute-ad"
	config4 "github.com/Meesho/iop-component/rv-substitute-exploit"
	configexplore "github.com/Meesho/iop-component/rv-substitute-explore"
	"github.com/Meesho/iop-starter/cohort/config/stdiop"
	config2 "github.com/Meesho/iop-starter/config"
	rvOnlineCG "github.com/Meesho/online-cg/client/component/rv_substitute/init"
	"github.com/Meesho/rv-iop/internal/api/controller"
	"github.com/Meesho/rv-iop/internal/api/handler/ad"
	"github.com/Meesho/rv-iop/internal/api/handler/exploit"
	explorervsubstitute "github.com/Meesho/rv-iop/internal/api/handler/explore"
	"github.com/Meesho/rv-iop/internal/api/router"
	"github.com/Meesho/rv-iop/internal/config"
	"github.com/Meesho/rv-iop/internal/dependency"
	segmentStoreClient "github.com/Meesho/segment-store/client/go-client"
	"github.com/rs/zerolog/log"
)

func dependencies() *dependency.Dependencies {
	service := getServiceConfInstance()
	iopStarter := getIopStarterConf()
	conf := getABConf()
	simpleHttp := simplehttp.NewSimpleHttp(conf)
	newSegmentStoreClient := segmentStoreClient.NewSegmentStoreClient()

	/*
		TODO: Pass through the env prefix, once circuit breaker is implemented
		When empty string is passed, circuit breaker will be disabled
	*/
	cbManager := circuitbreaker.NewManager("")
	standardIopConfigHandler := stdiop.NewStandardIopConfigHandler(iopStarter, simpleHttp, newSegmentStoreClient, cbManager)
	standardRvSubstituteFeedImpl := exploit.NewStandardRvSubstituteFeedImpl(service, standardIopConfigHandler)
	standardRvSubstituteAdFeedImpl := ad.NewStandardRvSubstituteFeedImpl(service, standardIopConfigHandler)
	standardRvSubstituteExploreFeedImpl := explorervsubstitute.NewStandardRvSubstituteFeedImpl(service, standardIopConfigHandler)
	controllerStandardRecentlyViewedFeedImpl := controller.NewStandardRecentlyViewedFeedImpl(standardRvSubstituteFeedImpl, standardRvSubstituteExploreFeedImpl, standardRvSubstituteAdFeedImpl)

	// rviopweb deployable is deprecated; RV interactions are only wired in rviopgrpcweb
	dependencyDependencies := dependency.NewDependencies(controllerStandardRecentlyViewedFeedImpl, nil)
	return dependencyDependencies
}

var (
	appConfig config.AppConfig
)

var (
	zkConf         *config.Zk
	serviceConf    *config.Service
	iopStarterConf *config2.IopStarter
)

// main is the entry point of the application
// It initializes the config, logger, http framework and metric
// It also registers all the APIs
func main() {
	go func() {
		fmt.Print(http.ListenAndServe(":8000", nil))
	}()

	// Initialize configuration first
	config3.InitGlobalConfig(&appConfig)
	zkConf = getZkConf()
	serviceConf = getServiceConf()
	iopStarterConf = &zkConf.IopStarter

	// Log configuration for debugging
	log.Info().Msgf("zk configs - %#v", zkConf)
	log.Info().Msgf("service config - %#v", serviceConf)

	logger.Init()
	metric.Init()
	httpframework.Init()
	prismlogger.InitV1()
	rvOnlineCG.Init()

	crossRE.Init(&zkConf.CrossRE, nil)
	config4.Init(&zkConf.RvExploit)
	configexplore.Init(&zkConf.RvExplore)
	configAd.Init(&zkConf.RvAd)

	router.Init(dependencies())
	log.Info().Msg("Service started successfully")
	httpframework.Instance().Run(":" + strconv.Itoa(serviceConf.App.Port))
}
func getServiceConf() *config.Service {
	return appConfig.ServiceConf
}
func getZkConf() *config.Zk {
	return &appConfig.ZkConf
}

func getABConf() *ab.Conf {
	return &ab.Conf{
		ExternalHttp: ab.ExternalHttp{
			Host:                serviceConf.ExternalService.Ab.Host,
			Port:                serviceConf.ExternalService.Ab.Port,
			Auth:                serviceConf.ExternalService.Ab.Auth,
			ReadTimeoutInMs:     serviceConf.ExternalService.Ab.ReadTimeOutInMs,
			DialTimeoutInMs:     serviceConf.ExternalService.Ab.DialTimeoutInMs,
			MaxIdleConns:        serviceConf.ExternalService.Ab.MaxIdleConns,
			MaxIdleConnsPerHost: serviceConf.ExternalService.Ab.MaxIdleConnsPerHost,
			IdleConnTimeoutInMs: serviceConf.ExternalService.Ab.IdleConnTimeoutInMs,
		},
	}
}

func getServiceConfInstance() *config.Service {
	return serviceConf
}

func getIopStarterConf() *config2.IopStarter {
	return iopStarterConf
}
