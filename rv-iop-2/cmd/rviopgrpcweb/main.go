package main

import (
	debugapi "github.com/Meesho/dag-debugger/debug/api/grpc"
	debugproto "github.com/Meesho/dag-debugger/debug/proto"
	"github.com/Meesho/go-client/ab"
	"github.com/Meesho/go-client/ab/simplehttp"
	"github.com/Meesho/go-core/circuitbreaker"
	config3 "github.com/Meesho/go-core/config"
	grpcServer "github.com/Meesho/go-core/grpc"
	"github.com/Meesho/go-core/logger"
	"github.com/Meesho/go-core/metric"
	"github.com/Meesho/go-core/prismlogger"
	crossRE "github.com/Meesho/iop-component/cross-re"
	crossSellExploit "github.com/Meesho/iop-component/cross-sell-exploit"
	rvInteractions "github.com/Meesho/iop-component/rv-interactions"
	configAd "github.com/Meesho/iop-component/rv-substitute-ad"
	config4 "github.com/Meesho/iop-component/rv-substitute-exploit"
	configexplore "github.com/Meesho/iop-component/rv-substitute-explore"
	"github.com/Meesho/iop-starter/cohort/config/stdiop"
	config2 "github.com/Meesho/iop-starter/config"
	memcoilClient "github.com/Meesho/memcoil/v2/pkg"
	"github.com/Meesho/memcoil/v2/pkg/memcoil"
	crossSellOnlineCG "github.com/Meesho/online-cg/client/component/cross_sell/init"
	rvOnlineCG "github.com/Meesho/online-cg/client/component/rv_substitute/init"
	rviop "github.com/Meesho/rv-iop/client/grpc/rv-iop"
	"github.com/Meesho/rv-iop/internal/api/grpc"
	"github.com/Meesho/rv-iop/internal/api/handler/ad"
	"github.com/Meesho/rv-iop/internal/api/handler/crosssell"
	"github.com/Meesho/rv-iop/internal/api/handler/exploit"
	explorervsubstitute "github.com/Meesho/rv-iop/internal/api/handler/explore"
	rvinteractionhandler "github.com/Meesho/rv-iop/internal/api/handler/rv-interaction"
	"github.com/Meesho/rv-iop/internal/config"
	segmentStoreClient "github.com/Meesho/segment-store/client/go-client"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	memcoilProvider "github.com/Meesho/memcoil/v2/pkg/provider"
)

var appConfig config.AppConfig

var (
	zkConf         *config.Zk
	serviceConf    *config.Service
	iopStarterConf *config2.IopStarter
)

// main is the entry point of the application
// It initializes the config, logger, http framework and metric
// It also registers all the APIs
func main() {

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
	prismlogger.InitV1()

	rvOnlineCG.Init()
	crossSellOnlineCG.Init()

	grpcServer.Init()

	// Obtain MemCoil client after grpcServer.Init()
	memcoilClient.Init()
	clientProvider := memcoilProvider.GetInstance()
	mCacheId := getMcacheIdFromEnv("MCACHE_CATALOG-VALIDITY_MCACHE-ID")
	catalogValidityMemCoilClient := clientProvider.GetMemCoilClient(mCacheId, memcoil.NewStringConverter())

	initDependencies(catalogValidityMemCoilClient)

	conf := getABConf()
	simpleHttp := simplehttp.NewSimpleHttp(conf)
	newSegmentStoreClient := segmentStoreClient.NewSegmentStoreClient()

	/*
		TODO: Pass through the env prefix, once circuit breaker is implemented
		When empty string is passed, circuit breaker will be disabled
	*/
	cbManager := circuitbreaker.NewManager("")
	standardIopConfigHandler := stdiop.NewStandardIopConfigHandler(iopStarterConf, simpleHttp, newSegmentStoreClient, cbManager)

	// RV Substitute handlers
	standardRvSubstituteFeedImpl := exploit.NewStandardRvSubstituteFeedImpl(serviceConf, standardIopConfigHandler)
	standardRvSubstituteAdFeedImpl := ad.NewStandardRvSubstituteFeedImpl(serviceConf, standardIopConfigHandler)
	standardRvSubstituteExploreFeedImpl := explorervsubstitute.NewStandardRvSubstituteFeedImpl(serviceConf, standardIopConfigHandler)

	// Cross Sell handlers
	standardCrossSellWidgetImpl := crosssell.NewStandardCrossSellWidgetImpl(serviceConf, standardIopConfigHandler)
	standardCrossSellFeedImpl := crosssell.NewStandardCrossSellFeedImpl(serviceConf, standardIopConfigHandler)

	// RV Interactions handlers
	standardRvInteractionsFeedImpl := rvinteractionhandler.NewStandardRvInteractionsFeedImpl(serviceConf, standardIopConfigHandler)
	standardRvInteractionsWidgetImpl := rvinteractionhandler.NewStandardRvInteractionsWidgetImpl(serviceConf, standardIopConfigHandler)

	// Create and register the gRPC server
	rvIopServer := grpc.NewRvIopGrpcServer(
		standardRvSubstituteFeedImpl,
		standardRvSubstituteExploreFeedImpl,
		standardRvSubstituteAdFeedImpl,
		standardCrossSellWidgetImpl,
		standardCrossSellFeedImpl,
		standardRvInteractionsFeedImpl,
		standardRvInteractionsWidgetImpl,
	)
	dagDebugSvc := debugapi.Init(iopStarterConf, standardIopConfigHandler)
	debugproto.RegisterDAGDebugServiceServer(grpcServer.Instance().GRPCServer, dagDebugSvc)
	rviop.RegisterRvIopServiceServer(grpcServer.Instance().GRPCServer, rvIopServer)

	log.Info().Msg("Service started successfully")
	grpcServer.Instance().Run()
}

func initDependencies(memCoilClient memcoil.Client) {
	// rvOnlineCG and crossSellOnlineCG already inited before grpcServer.Init()
	crossRE.Init(&zkConf.CrossRE, memCoilClient)
	config4.Init(&zkConf.RvExploit)
	configexplore.Init(&zkConf.RvExplore)
	configAd.Init(&zkConf.RvAd)
	crossSellExploit.Init(&zkConf.CrossSellExploit)
	rvInteractions.Init(&zkConf.RvInteractions)
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

// getMcacheIdFromEnv fetches int values from env
func getMcacheIdFromEnv(key string) int {
	if !viper.IsSet(key) {
		log.Panic().Msgf("required env not set - %s", key)
	}
	return viper.GetInt(key)
}
