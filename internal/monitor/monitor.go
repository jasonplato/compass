package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/compass/chains"
	"github.com/mapprotocol/compass/internal/bsc"
	"github.com/mapprotocol/compass/internal/chain"
	"github.com/mapprotocol/compass/internal/eth2"
	"github.com/mapprotocol/compass/internal/klaytn"
	"github.com/mapprotocol/compass/internal/matic"
	"github.com/mapprotocol/compass/mapprotocol"
	"github.com/mapprotocol/compass/msg"
	"github.com/mapprotocol/compass/pkg/ethclient"
	utils "github.com/mapprotocol/compass/shared/ethereum"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Monitor struct {
	*chain.CommonSync
}

func New(cs *chain.CommonSync) *Monitor {
	return &Monitor{
		CommonSync: cs,
	}
}

func (m *Monitor) Sync() error {
	go func() {
		m.sync()
	}()
	return nil
}

// sync function of Monitor will poll for the latest block and listen the log information of transactions in the block
// Polling begins at the block defined in `m.Cfg.startBlock`. Failed attempts to fetch the latest block or parse
// a block will be retried up to BlockRetryLimit times before continuing to the next block.
// However，an error in synchronizing the log will cause the entire program to block
func (m *Monitor) sync() {
	for {
		time.Sleep(time.Hour)
	}
}

type Req struct {
	ChainId int64  `json:"chain_id"`
	Tx      string `json:"tx"`
}

func Handler(resp http.ResponseWriter, req *http.Request) {
	bytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte("Server Internal Error"))
		return
	}

	r := Req{}
	err = json.Unmarshal(bytes, &r)
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte("Server Internal Error"))
		return
	}

	if strings.ToLower(mapprotocol.OnlineChaId[msg.ChainId(r.ChainId)]) == chains.Near {
		ret := map[string]interface{}{
			"proof": "0x0xd33a28a20000000000000000000000000000000000000000000000004d415001000000010000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000218000000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000f800000000000000000000000000000000000000000000000000000000000000f13a4d7910ce4815e597f64fc26f51497a52700b15e689514f6cd573ed771fd9fca7189cae4525a3af326c84e0067b8f0b84077302b392fbb8eaccf9ecf03608d766f15110700000000a04b0f940c48f62846a059aedf0d6f3e1a312d3c106e9cbce14a72ba003f3e02b459dd600d0fea5e0db9520e0146d7454cfbab25639cf6012d4bc9079499a8ba164cf66f029be3d1375000c2962dd475a46c501dddc28c484bc96bed85cb44e946ff2d66481c6d316cc4887be3f1866ca0c7bef5aa2b978f02b6369aeba103e97b5926f95ff34717879eb7d55ad5c83168b92f09991bef3f495f081eb9d1cdbde39ead1b39c16ebd8dd06643e55d40818b4f62e92f55618a4f2d7833a9735c7ac17b42180c0632114cc803e9d1675c3a94c7f6965ff408bf8af1a16696144e7a0bc1df6a7358316f012100000000050000006e6f64653200e9f1e1f48f966f3cb086e238c8f2d9d6a54063a8e757cf5587bf429a5d2292c20924e6b8a162606c8379cebadc01000000050000006e6f646531004d7a80636b24694ba9c928394118cd539bf863f5e5773a0de27c7677d3eb0a87f36048a1f26dde289dfafa9cdc01000000050000006e6f646533000e8202a3ff22a3df1bf2221f5079005e0bc1438437a2811d513e795d797a5824843932d4fe7200e17b0a4888dc01000000160000006175726f72612e706f6f6c2e663836333937332e6d30007fdc7a529b631deb0c6c75e1893696781b3ff4d3e4a8fadee70dce85b80113c88a7ca14371c187634c9d48f3c5000000001600000030316e6f64652e706f6f6c2e663836333937332e6d30002850e20b8ce610aee3d913441ff7ade411d5524f445dfd8c2fc7784d0e5155ad7d008ebc1ba38ec386981e4e8d00000000170000006c6567656e64732e706f6f6c2e663836333937332e6d30009012890b00163561ec8bd99da8715c61c32b5be92f3dc00acd7cd9479f31929832a502dc160bcfc412299c46730000000019000000657665727374616b652e706f6f6c2e663836333937332e6d3000317f139e9df2e803cee46652ac7701aec09154c3f8cfd13acb24180069066fa297f87edc811262b3cabb3c4c68000000001900000063686f7275736f6e652e706f6f6c2e663836333937332e6d30002491a39951692ea9824a9821c6ffadb15aed6801ebbb587d225d4bb404110decfe006cf8977b1700332bb07a6700000000120000006e692e706f6f6c2e663836333937332e6d3000e8a88ee26c0c29b07da2ed2bea825cf08acbb57ee98f534aa7473eb27faa5d75397cd429f74f9df31fd18e985e000000001a0000007374616b656c795f76322e706f6f6c2e663836333937332e6d30005bdc205a212c4573d32375b2f93ad215c21a34508ef89123a3ac5f8492552fd7096ed2bea224fdbd24c764815e000000001a000000666f756e6472797573612e706f6f6c2e663836333937332e6d3000885ad5cbf24be2f85e01da691f16a0beb7d012fdeeb1a1d2f7efa8a14a48daba6b0b7c84da670375656f0bb41900000000190000006c756e616e6f7661322e706f6f6c2e663836333937332e6d30007b74680ad4633707d819b83f542232bcc88e66a28f23e6b0bf0d0c85d63502f83041b02cb9fbadd2a9bbd32417000000001f00000070617468726f636b6e6574776f726b2e706f6f6c2e663836333937332e6d3000a7891814823af0fce30a3e5037534803fed58c0ab7e0de2aea9e8ef5982c3e31400db3b793f50fac784f6404100000000019000000626565317374616b652e706f6f6c2e663836333937332e6d300096360a2d46e98ba26fe18f88cfba3f225abe1986e413b7862d74c618e52082a5f7072cab712fa00ae032a7640e000000001b0000007374616b657373746f6e652e706f6f6c2e663836333937332e6d300026366f84610fdd95923c16f9575e609ac5537adc7d5beb7b472ee7fea4db0379abff64637347b945611a43e00d00000000180000006c6561646e6f64652e706f6f6c2e663836333937332e6d3000acc27128aea3731cad4f045b0d011892f85857867a89305393b9bce796a316d51f38f929392562e0b44f97ca0d000000001a000000626c6f636b73636f70652e706f6f6c2e663836333937332e6d30004eed9f8ec11536f0c3c8f5171365b2524afb37d07a2052e433654a2c514e3e39e36f26ed2a1aa8efaaa8c1a90d0000000018000000647372766c6162732e706f6f6c2e663836333937332e6d30004a75120cf200659da75ee453d629fc0a411e5764652e441a3119d67097e20a98f5e6349ae488a98bd028afb30c000000001e000000626173696c69736b2d7374616b652e706f6f6c2e663836333937332e6d3000a73adabf811046d06cb097d6d6d967af363bf9485b8f3c63f968522ea2ff7b75df3021ce7227b9592653b4180b000000001c000000696e66696e6974656c6f6f702e706f6f6c2e663836333937332e6d300018bf01eeff2dd463487016ee068143b22832530485783a7449f92769ab2fd422b301dfacc770ff66a467f5bd0800000000140000006b696c6e2e706f6f6c2e663836333937332e6d3000a0e96a90b0e21954c40e5e2fc7d56c412f1d52af0a8ea453d088a9c3659b8abd214308416f73f8a205c82e480700000000170000006d6f6f6e6c65742e706f6f6c2e663836333937332e6d30002732c023279285046e52b946e9c8f96b0ee01fb29403ba714e4b8b5f2a746bca78db2da5db94b9f7e24560f505000000001b00000067657474696e676e6561722e706f6f6c2e663836333937332e6d30004193e2cd0630b71f1af526515f63863aad964e08eb451d2dd9454ec6041a4e7d36f0618fa24ffdf7b9279b3901000000001400000062672d302e706f6f6c2e663836333937332e6d30009f99bca8fd4c126fc1f9ade6ca38c0b308dfa8924e068a3666bb507b8764fc25278e0ff6ed674877b67be33301000000001a000000737373757279616e73682e706f6f6c2e663836333937332e6d3000ad776676e1e8d4204591ac62dc61e453feeb6e2ad8d2b43d9e3c14b97241836a82d955a034bbf5f7927d530d0100000000150000006d6f6e646c696368742e706f6f6c2e6465766e657400aebdd9da72e4fe7cb226c712c19ad3c584cfe8c8c08dba1dbc753d3b06c448a9001d78d69567a91071b9ecd70000000000190000007477696e7374616b652e706f6f6c2e663836333937332e6d3000149f4176da42a9c0791e0392f54b25e31a26c02579c603637dd71a397b7515425b11641c394370e8a6270bad000000000018000000676172676f796c652e706f6f6c2e663836333937332e6d3000261c1e846197c8df135e9949089d0b4735aaa2a3ccf639ac4a6aec59ff398a9a92f2bd12879a3f545ee598ac00000000001400000062672d312e706f6f6c2e663836333937332e6d30009c283a0348f37eab35769b210712958f3a6ce63aae5bb8f33f3e0167c6841d48ad7d295d938cd2e9efbef68d0000000000180000006a7374616b696e672e706f6f6c2e663836333937332e6d300009f7aaacc1dedaf1ccbe4402ed3d78458f3e39ccce3993c561d85f1312a718e3635cee596b30107f30a990880000000000160000006267706e74782e706f6f6c2e663836333937332e6d3000bb5052c67b1cfcd47bd9610ab4b42b7895f7c7525d67e6848cd4dfcf6045dfeca8f03fe50f359dfa97cd8d850000000000180000006c6173746e6f64652e706f6f6c2e663836333937332e6d3000680294de4a16790d40c41de999602a7cf9f2a9383fd793a42defd1788cba8a982bbe6142eba6f14b0b4d058100000000001b00000063727970746f6c696f6e732e706f6f6c2e663836333937332e6d300093630e39084f59480edce26296f4c225e729ba327b91741cd0a9055c7a993f41305dc02cb7049a655bd67d7c000000002c0000000100b4d36cd65e33061d2c223d8c487f9f40577c759005bd73665a28ebab1156e5919f4645a924e0c2b74c03817d29f686cf14b5143465c864d703f1750afc3cf60c0100bddf91b3c1b06d1d8ffa28556c219c93f2c78ee73542b202b56faa34761050bec129952b17f22a7e748cc18168d80043e8888f882b8b3827d3043ff95b37610c0100c0621d5230d8c031d30d2ea1ab545d19c12e62c0ca77d4f10344c32f9509c43cd3532ef7a6f92898ad483b5fa4ab2e70e267515d89d7cb8c77f24fee4b1b870e010012f215bd4a3010404d108ebdd4e7bbefcf005b4553d10063c87849d59d3f2c5453f04715c4e4efec76a930025b5a9acdb1b8b1f0b33a67c51f6fc939774300090100a3a11ec0b9c3e628cd76d9d438f2723345a77db260a1fac40ff18758c8215b2f316a192bf38f0f586ff38526533d7d181a1db7c404cb34f09634d5e10f12cd07010032468d288db91d87f7487fd7253559caeb95474bc8fb96596f2fa186be88ce942220cb2ab757d4d8af221debdf8bdf9744bf46e5e6db5a19f6234b08ea3ef9010001009c4248083b0820c0ecc192c73b5dd159d4f6631ac45b34f184ee0663f1c048d0086b3986d84a2b5bd5ece9c9dc6ce6f3b1bbc6495be3940ba8c6dc4e218dcf010000000000000100c1ebab369403712da31ab77c85ce36c8246c79fe4e953cb761084c3f4b5ba556df5ebdf1fe944b63a694b55397885eb229f5ad3b897d825ae6b68e1dc62832090001007e11911eb6b689d8998623cc00b1fad93cced0a6d8e234764e9f5b779c239207525b80c7f141988f1895cfd86ba6288ad2f4717bfaee0198229165104b6c5b090100fb9e78b6550161a3be10f8899814dd5c3cb690cfef071905589438f83e153db908565e71ec589841897c81e01443eccbac67549995c331fce958242820920f0700000100bd48605fdf377391a62688e7a7bde3db8f43500f51184c923c43c945edd88434cbdb5db52948657f1d50093156b4bce1ad5dc7e7ba206799433086481a91ec07000000000000000000000100f54719f857b8aac4fdd4d466d4575209bc7bff354b389cfd512883af8b10fc4ff44068bacc5627d9e49ff9018c0e2a4561e08637f18f89a8431d7c6544f371060100806002b7322ec818e8aeb098674da6a9ef134247aa2ea884a6b356e2bd274e4670886b437a575b1bdccbb6815f05d86bcf3786a0bf31c702dc8540c296d1f20f000000000000010065dc8c9eeab50adb3638db48d6a9aacf7364897438a104b93f239539937365b1bd268b25c03402628972d2418c4de2d29a234ad296176e8b3c1d6b65e999f30f0000010099120902ccf22d466c8bf921178a8b7a828fe523141b2e791a40d82eb06c4b4f5fa83f481167d125f60bad2bf005b35357ad513fbb402c3864397d4acd4b6b0a000000000000000000000000000000000000000000000000000000000000000000000000000000000000000011d702000000d01ae8395e8dee5fc97776d71ed6d71e9d66671e54c46382f3a6a96c5533eb6e00e7f93ffc31e9f41cc6784fcb6bf9ec60c5558b516ff9fef459a6a753ff28b7fd00285bd566667e07284a18c007f03e2ce865903144aa91b9dc328534a7ac01615398637018beeb9a82c873aea35c5b69040a7ada366b3693d8028a2d7e293916a9020000001f08000073776170206f75743a207b2266726f6d5f636861696e223a2235353636383138353739363331383333303839222c22746f5f636861696e223a223937222c226f726465725f6964223a22307833313437616634636234303639666231353163353836346437663162346433623133366561656230333237633731646439623265383336613037393733376664222c22746f6b656e223a5b3131372c3131352c3130302c39392c34362c3130392c39372c3131322c34382c34382c35352c34362c3131362c3130312c3131352c3131362c3131302c3130312c3131365d2c2266726f6d223a5b3130362c3131352c3131312c3131302c3131392c3131312c3131312c3130302c34362c3131362c3130312c3131352c3131362c3131302c3130312c3131365d2c22746f223a5b3231322c33322c3130342c37332c3134322c3232382c34312c3233362c3133362c3133332c3131362c3233392c3233322c35342c33352c3231352c3131332c3232362c3234362c3134375d2c22616d6f756e74223a2231303030303030303030222c22737761705f64617461223a22307830303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303630303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303163303030303030303030303030303030303030303030303030306432386331313837313638646139646631623766366362383439356536353933323264323763396630303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303031303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303032303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303033363161303834303565386664383030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303031323363303633376165356335313834343363303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303038303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303830303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303032303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303230303030303030303030303030303030303030303030303064386636396531663130306462363535643435303335343563336262333038636161623461336236303030303030303030303030303030303030303030303030623434333838326563373465366632363331323637666634666330653034633036663030303038393030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030313462343433383832656337346536663236333132363766663466633065303463303666303030303839303030303030303030303030303030303030303030303030222c227261775f737761705f64617461223a7b22737761705f706172616d223a5b7b22616d6f756e745f696e223a22393938303030303030303030303030303030303030222c226d696e5f616d6f756e745f6f7574223a2235333831383635353834363534373238383430323532222c2270617468223a22307830303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303230303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030323030303030303030303030303030303030303030303030306438663639653166313030646236353564343530333534356333626233303863616162346133623630303030303030303030303030303030303030303030303062343433383832656337346536663236333132363766663466633065303463303666303030303839222c22726f757465725f696e646578223a2230227d5d2c227461726765745f746f6b656e223a22307862343433383832656337346536663236333132363766663466633065303463303666303030303839222c226d61705f7461726765745f746f6b656e223a22307864323863313138373136386461396466316237663663623834393565363539333232643237633966227d2c227372635f746f6b656e223a22757364632e6d61703030372e746573746e6574222c227372635f616d6f756e74223a2231303030303030303030222c226473745f746f6b656e223a22307862343433383832656337346536663236333132363766663466633065303463303666303030303839227d200500006361316366386365626638383439393432396363613866383763626361313561623864616664303637303232353961353334346464636538396566336633613566393032366438383464343135303031303030303030303136316130333134376166346362343036396662313531633538363464376631623464336231333665616562303332376337316464396232653833366130373937333766643933373537333634363332653664363137303330333033373265373436353733373436653635373439303661373336663665373736663666363432653734363537333734366536353734393464343230363834393865653432396563383838353734656665383336323364373731653266363933383433623961636130306239303230303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030363030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030316330303030303030303030303030303030303030303030303030643238633131383731363864613964663162376636636238343935653635393332326432376339663030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303130303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303230303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303336316130383430356538666438303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303132336330363337616535633531383434336330303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303830303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030383030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303230303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030323030303030303030303030303030303030303030303030306438663639653166313030646236353564343530333534356333626233303863616162346133623630303030303030303030303030303030303030303030303062343433383832656337346536663236333132363766663466633065303463303666303030303839303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303031346234343338383265633734653666323633313236376666346663306530346330366630303030383930303030303030303030303030303030303030303030303001000000635c861b92d1f6222e12a8dccc65fab1d5126da48d9d9191c6c56d1bc6ca30d5fa69f4718305000000ba661f235cdedc2000000000000000120000006d6f732e6d61703030372e746573746e6574020300000022302202000000407894aba3ab65d1a9b57a5dcd22e5aca2d4010264bcdf5e54d93889922a6b0b002eeb74a6177f588d80c0c752b99556902ddf9682d0b906f5aa2adbaf8466a4e900e4e8448855f4ea3efe64cf9d4d0d74beaae725474e8a893fab560760eee9f1d00584523ef8deed6dbf99f754b5a2d02cbf56403f6f6d32b8c7588ba5a2429e3a6a15110700000000a04b0f940c48f62846a059aedf0d6f3e1a312d3c106e9cbce14a72ba003f3e02b459dd600d0fea5e0db9520e0146d7454cfbab25639cf6012d4bc9079499a8ba50a6382bf2a246bc6d595ef7cdad28e264b519a42d84ff159ff509668d739901e0ff3a04bdcfc0c01a782d7cd1f9d1dfe5ad5804187df69ff5a12a1d68d1446d54d68a3d5ff34717879eb7d55ad5c83168b92f09991bef3f495f081eb9d1cdbde39ead1b39c16ebdd60cf7ca8ce071402084d2532079ff0d5a83044cdd779fc4bf73cafc7dd6112f12000000e4e8448855f4ea3efe64cf9d4d0d74beaae725474e8a893fab560760eee9f1d000a9d1232e1712d966579eb988fc537aecb139b9586a2256922586ef852d1eff2801d1a2c05535fb71a9973dd4048e5a889047398d3be1b16a659168aabbbf3e120e01a2b54de3d2f5a86a41869233e385b8bf71034b1add01bc063178cc5b7f2bfc1300221264381ec8ec61cb40b9d4480d563fe91a6534c8c775ff2a3f3319c00290050046097e478943771c06dd4360eb0436b7cbe2d299e8988b5a10e9c510adf816a1008bb45bc470c9ace2778c08e654a2ef6554f4c62bb704547ec0dc537f251e0adb000d1e3ee9f9d235b0085a333d9dfbe804752bc7b30e35b97a28579c1b4d2c2d7b007c0a40b5a0df44a210604f9b4a3431bfebf89a86befe8a4d2a96634f14e874970064ed93218b34d62cbc36757ff110f49e73ce501446e2c4d2dba1d8ae5cf4e0040072572826ce3c187f3ca16c53a3804fe92780f3c7af2adbcc5c642eb68a3f550f00b81d69b2366459133cfa2e14074218543f3be6f0f05068c3d3f93f3d43dba70a002864b6bdc8bf86961a086007b8a7d4ec0c104191f7f54ab0415348ec9b56093700adf1a676082b3f06a0c63b79c24993c312d6810062e425448ff8b0f71797bd0a00e27353c84ab067183cb21921378f6211821e54fd57c288cdd884dbdbaad154bb0057261bb221fecd3121fcdb345cb99aace4afe5f90fb90f95a6fbf4552f1a4aec001b17b83832de71bfb37b02ec0aba5dfc8f6d3e6ecb1e8f387614d352edcd25b6008e4aeae42d6aadf15ad68f3196d948afbb14c452c1165de6e63d4fc2b324748b00000000000000000000",
		}

		d, _ := json.Marshal(ret)
		_, _ = resp.Write(d)
		return
	}

	cfg, ok := mapprotocol.OnlineChainCfg[msg.ChainId(r.ChainId)]
	if !ok {
		log.Info("Found a log that is not the current task ", "toChainID", r.ChainId)
		resp.WriteHeader(404)
		resp.Write([]byte(fmt.Sprintf("This ChainId(%d) Not Support", r.ChainId)))
		return
	}
	client, err := ethclient.Dial(cfg.Endpoint)
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte("Server Internal Error"))
		return
	}
	receipt, err := client.TransactionReceipt(context.Background(), common.HexToHash(r.Tx))
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte("Server Internal Error"))
		return
	}
	find := false
	method := mapprotocol.MethodOfTransferIn
	logParam := receipt.Logs[0]
	for _, l := range receipt.Logs {
		if mtd, ok := mapprotocol.Event[l.Topics[0]]; ok {
			logParam = l
			method = mtd
			find = true
			break
		}
	}
	if !find {
		resp.WriteHeader(400)
		resp.Write([]byte("This Tx Not Match"))
		return
	}
	var data []byte
	switch strings.ToLower(mapprotocol.OnlineChaId[msg.ChainId(r.ChainId)]) {
	case chains.Bsc:
		data, err = bsc.GetProof(client, receipt.BlockNumber, logParam, method, msg.ChainId(r.ChainId))
	case chains.Map:
		data, err = utils.GetProof(client, receipt.BlockNumber, logParam, method, msg.ChainId(r.ChainId))
	case chains.Matic:
		data, err = matic.GetProof(client, receipt.BlockNumber, logParam, method, msg.ChainId(r.ChainId))
	case chains.Klaytn:
		kc, err := klaytn.DialHttp(cfg.Endpoint, true)
		if err != nil {
			resp.WriteHeader(500)
			resp.Write([]byte("Klaytn InitConn Failed, Server Internal Error"))
			return
		}
		data, err = klaytn.GetProof(client, kc, receipt.BlockNumber, logParam, method, msg.ChainId(r.ChainId))
	case chains.Eth2:
		data, err = eth2.GetProof(client, receipt.BlockNumber, logParam, method, msg.ChainId(r.ChainId))
	default:
	}
	client.Close()
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte("Server Internal Error"))
		return
	}
	ret := map[string]interface{}{
		"proof": "0x" + common.Bytes2Hex(data),
	}

	d, _ := json.Marshal(ret)
	_, _ = resp.Write(d)
}
