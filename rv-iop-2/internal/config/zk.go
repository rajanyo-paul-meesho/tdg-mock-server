package config

import (
	crossREComp "github.com/Meesho/iop-component/cross-re/config"
	crossSellExploitConf "github.com/Meesho/iop-component/cross-sell-exploit/config"
	adConf "github.com/Meesho/iop-component/rv-substitute-ad/config"
	exploitConf "github.com/Meesho/iop-component/rv-substitute-exploit/config"
	exploreConf "github.com/Meesho/iop-component/rv-substitute-explore/config"
	rvInteractionsConf "github.com/Meesho/iop-component/rv-interactions/config"
	"github.com/Meesho/iop-starter/config"
)

type Zk struct {
	crossREComp.CrossRE
	exploitConf.RvExploit
	exploreConf.RvExplore
	adConf.RvAd
	crossSellExploitConf.CrossSellExploit
	rvInteractionsConf.RvInteractions
	config.IopStarter
}
