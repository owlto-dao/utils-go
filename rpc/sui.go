package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/sui"
	_ "github.com/gagliardetto/solana-go"
	"github.com/owlto-dao/utils-go/loader"
	"github.com/owlto-dao/utils-go/util"
)

type SuiRpc struct {
	tokenInfoMgr *loader.TokenInfoManager
	chainInfo    *loader.ChainInfo
	client       sui.ISuiAPI
	legTokens    map[string]interface{}
}

func NewSuiRpc(chainInfo *loader.ChainInfo) *SuiRpc {
	var legs map[string]interface{}
	json.Unmarshal([]byte(legTokenStr), &legs)
	return &SuiRpc{
		client:       chainInfo.Client.(sui.ISuiAPI),
		chainInfo:    chainInfo,
		tokenInfoMgr: loader.NewTokenInfoManager(nil, nil),
		legTokens:    legs,
	}
}

func (w *SuiRpc) IsAddressValid(addr string) bool {
	return strings.HasPrefix(addr, "0x") && len(addr) == 66 && util.IsHex(addr[2:])
}

func (w *SuiRpc) GetChecksumAddress(addr string) string {
	return addr
}

func (w *SuiRpc) GetBalanceAtBlockNumber(ctx context.Context, ownerAddr string, tokenAddr string, blockNumber int64) (*big.Int, error) {
	return w.GetBalance(ctx, ownerAddr, tokenAddr)
}

func (w *SuiRpc) GetTokenInfo(ctx context.Context, tokenAddr string) (*loader.TokenInfo, error) {
	tokenAddr = strings.TrimSpace(tokenAddr)
	if legToken, ok := w.legTokens[tokenAddr]; ok {
		data := legToken.(map[string]interface{})
		return &loader.TokenInfo{
			TokenName:    data["symbol"].(string),
			ChainName:    w.chainInfo.Name,
			TokenAddress: tokenAddr,
			Decimals:     int32(data["decimals"].(float64)),
			FullName:     data["name"].(string),
			Icon:         data["iconUrl"].(string),
			TotalSupply:  big.NewInt(1000000000),
		}, nil
	}

	if util.IsHexStringZero(tokenAddr) {
		tokenAddr = "0x2::sui::SUI"
	}

	tokenInfo, ok := w.tokenInfoMgr.GetByChainNameTokenAddr(w.chainInfo.Name, tokenAddr)
	if ok {
		return tokenInfo, nil
	}

	rsp, err := w.client.SuiXGetCoinMetadata(ctx, models.SuiXGetCoinMetadataRequest{
		CoinType: tokenAddr,
	})
	if err != nil {
		return nil, err
	}

	ti := &loader.TokenInfo{
		TokenName:    rsp.Symbol,
		ChainName:    w.chainInfo.Name,
		TokenAddress: tokenAddr,
		Decimals:     int32(rsp.Decimals),
		FullName:     rsp.Name,
		Icon:         rsp.IconUrl,
	}
	w.tokenInfoMgr.AddTokenInfo(ti)

	trsp, err := w.client.SuiXGetTotalSupply(ctx, models.SuiXGetTotalSupplyRequest{
		CoinType: tokenAddr,
	})
	if err == nil {
		totalSupply, ok := big.NewInt(0).SetString(trsp.Value, 0)
		if ok {
			ti.TotalSupply = totalSupply
		}
	}

	return ti, nil
}

func (w *SuiRpc) GetBalance(ctx context.Context, ownerAddr string, tokenAddr string) (*big.Int, error) {
	ownerAddr = strings.TrimSpace(ownerAddr)
	tokenAddr = strings.TrimSpace(tokenAddr)
	if util.IsHexStringZero(tokenAddr) {
		tokenAddr = "0x2::sui::SUI"
	}
	rsp, err := w.client.SuiXGetBalance(ctx, models.SuiXGetBalanceRequest{
		Owner:    ownerAddr,
		CoinType: tokenAddr,
	})
	if err != nil {
		return nil, err
	}
	num, ok := big.NewInt(0).SetString(rsp.TotalBalance, 0)
	if !ok {
		return nil, fmt.Errorf("sui balance invalid %s", rsp.TotalBalance)
	}
	return num, nil

}

func (w *SuiRpc) GetAllowance(ctx context.Context, ownerAddr string, tokenAddr string, spenderAddr string) (*big.Int, error) {
	return big.NewInt(0), fmt.Errorf("not impl")
}

func (w *SuiRpc) IsTxSuccess(ctx context.Context, hash string) (bool, int64, error) {
	return false, 0, fmt.Errorf("not impl")
}

func (w *SuiRpc) GetClient() sui.ISuiAPI {
	return w.client
}

func (w *SuiRpc) Client() interface{} {
	return w.chainInfo.Client
}

func (w *SuiRpc) Backend() int32 {
	return 9
}

func (w *SuiRpc) GetLatestBlockNumber(ctx context.Context) (int64, error) {
	return 0, fmt.Errorf("not impl")
}

var legTokenStr string = `{
  "0x43e9045850072b10168c565ca7c57060a420015343023a49e87e6e47d3a74231::hoppy::HOPPY": {
    "name": "HOP BUNNY",
    "symbol": "HOPPY",
    "description": "HOP BUNNY",
    "iconUrl": "https://images.hop.ag/ipfs/QmdSDb1nWDfZFnfrExfkYXvk7R3AMGiWcJQw2cAN354CCy",
    "decimals": 6
  },
  "0x0d7acc8e91ea02582f3111dd6809f4f65e2ca2c1a0d13c966af3a673603ced42::orcie::ORCIE": {
    "name": "Orcie",
    "symbol": "ORCIE",
    "description": "Here to make a splash on Sui!",
    "iconUrl": "https://images.hop.ag/ipfs/QmZUt7qM1GH1aTggoNMykKhP4F8kffagsrfkkEACav83KL",
    "decimals": 6
  },
  "0xfd1d9f625036109fc4c5be1e0a0485abc77470e9ea7fc475b2e73f07251e234a::one::ONE": {
    "name": "1",
    "symbol": "1",
    "description": "1",
    "iconUrl": "https://images.hop.ag/ipfs/QmYzrog6p5VNME4Q1BoLH1JWSJBotsdZ2vJedszyRHEm18",
    "decimals": 6
  },
  "0x870c62977acdcf7cfba198eb7a92854ee65ad89a42cf731fbe0eac8947e3f6b5::hoppy::HOPPY": {
    "name": "HOPPY",
    "symbol": "HOPPY",
    "description": "HOPPY",
    "iconUrl": "https://images.hop.ag/ipfs/QmNvPcfqycMYz5ggfdizLLjDkUDW89Jc1DND85EdBo5sg5",
    "decimals": 6
  },
  "0x32f310486e71202187848902efd49726277edff6d0269bf08891dccc059aeeb2::yeti::YETI": {
    "name": "Yeti",
    "symbol": "YETI",
    "description": "Yeti",
    "iconUrl": "https://images.hop.ag/ipfs/QmTPcNsitTYs1hbs2jYKJUmUMCAm5FZHUcsv8bJNrYGYf8",
    "decimals": 6
  },
  "0x35ec45f1927663cbe36b07555b30bd25ccce7b0b938f2564ea185ffd879be4a2::hopeless::HOPELESS": {
    "name": "HOPeless",
    "symbol": "HOPELESS",
    "description": "Hang tight lads few minutes",
    "iconUrl": "https://images.hop.ag/ipfs/QmUSUvz6LgTdwnsTauQf5u6kPex8NwaMV2QVxgprTqc5Si",
    "decimals": 6
  },
  "0xeedb38544f4a0d264ff9ed9738768b6d8096d70499bdfa47c605d74fd2de1294::siu::SIU": {
    "name": "Siuuuuu",
    "symbol": "SIU",
    "description": "Siuuuuu",
    "iconUrl": "https://images.hop.ag/ipfs/QmWTYpGXt5vSa2Cb1Db2Lui7XkCypkSwur22btHymr1Jwm",
    "decimals": 6
  },
  "0x20883dfae09b5dd4a5433a1207bdd9fa945c9be8da24ae64f8af4145ca9af0e6::rex::REX": {
    "name": "Suirex",
    "symbol": "$REX",
    "description": "Hop like Rex. Rex like Hop",
    "iconUrl": "https://images.hop.ag/ipfs/QmdHT3ngjCqfn7wq8n668cnrqkaJ8YFQwt5GAghiQ9STuq",
    "decimals": 6
  },
  "0xe74876860836a4e2f2aed6cf8620f2652d8efc0d3c6e0255047b7c63c0e194fa::swoop::SWOOP": {
    "name": "Swoop",
    "symbol": "Swoop",
    "description": "🐤💧 Woop Woop Woop 💧🐤",
    "iconUrl": "https://images.hop.ag/ipfs/Qmb9SzPXE58pxXy8fXkGae31uSGfo9cPuhTHRNeS95FhWz",
    "decimals": 6
  },
  "0x02be491c1755a14830f0702825069b3c7c8b373ab44dd3c191a8113205084c23::finger::FINGER": {
    "name": "Fingerprint",
    "symbol": "Finger",
    "description": "Leave your mark on the SUI ecosystem with Fingerprint. One touch at a time!",
    "iconUrl": "https://images.hop.ag/ipfs/QmNjshAEe9bKQ3ixcDqgZ7XrDxXbGEANsAzVJUMKhroYQt",
    "decimals": 6
  },
  "0x8088eeb68375927b42154f929c6d2f71df197547d03007b0f3b2a800e127479a::suilady::SUILADY": {
    "name": "The Queen Of Sui Ocean",
    "symbol": "SUILADY",
    "description": "The Queen Of Sui Ocean",
    "iconUrl": "https://images.hop.ag/ipfs/QmQeHxQXeguR8oeqf1zwciNn51oKRfkCszfBLwraBBhhB5",
    "decimals": 6
  },
  "0xc8961689a6a478438aa39dc5745cfe519a8df98586c7e82a87a6db608a21730e::suinut::SUINUT": {
    "name": "suinut",
    "symbol": "SUINUT",
    "description": "SUI was here",
    "iconUrl": "https://images.hop.ag/ipfs/QmYThJBWxVHUC4tLg9UkLKMQPvebeb6j8wQEEcxUJraXqo",
    "decimals": 6
  },
  "0x660edaa68c5553d962d99770153e6acf564c7bc2ba7b85b856e41f3df39333e6::unihop::UNIHOP": {
    "name": "UNIHOP",
    "symbol": "UNIHOP",
    "description": "The dog of sui co-founder, Evan Cheng.",
    "iconUrl": "https://images.hop.ag/ipfs/QmTLVZuQqm4A1UQaktUSyVeGcA7hzba4qq6ZL4oM421nfs",
    "decimals": 6
  },
  "0xc9cf8d39041f284877f0384deb5b7ba2aee691cf1aa7b817f844a5af0095351e::trump_2024::TRUMP_2024": {
    "name": "☑ TRUMP2024",
    "symbol": "TRUMP2024",
    "description": "The official TRUMP2024 token rallies Trump supporters as election day nears. If Donald Trump wins, $100.000 worth of SOL will be rewarded to moonshot holders based on their contribution.",
    "iconUrl": "https://images.hop.ag/ipfs/QmcDtgNM1PpTea6zxb6jkSFhkmjefMUpazqAYwq91E4aE6",
    "decimals": 6
  },
  "0x068b7840d85c99d273085f5366052d1eca3684d3f2df567a98b8423c237bb954::boi::BOI": {
    "name": "BoiOnSui",
    "symbol": "BOI",
    "description": "BOI good boi of SUI",
    "iconUrl": "https://images.hop.ag/ipfs/QmVKNL7YQh94J1pUzwUqg8hdg84bkLjr2HXi6CEFk1zpK4",
    "decimals": 6
  },
  "0x5217def12dd5c456244d6cd643f4360736ba281c7b3cda1d7a7ed45a49119606::hopdog::HOPDOG": {
    "name": "HOPDOG",
    "symbol": "HOPDOG",
    "description": "Meet Hopdog, the dog of SUI and loves to hop!",
    "iconUrl": "https://images.hop.ag/ipfs/QmRKatJ6xRmwJ3NaK24UT4WwU7YqCXTWQbTf4AZGnHPn8R",
    "decimals": 6
  },
  "0x19bb4ac89056993bd6f76ddfcd4b152b41c0fda25d3f01b343e98af29756b150::cally::CALLY": {
    "name": "Pepecally",
    "symbol": "CALLY",
    "description": "PepeCally is a meme coin on the Sui blockchain, inspired by the iconic internet frog Pepe and built around a community-driven spirit. It’s designed for those who appreciate humor and want to join a trending crypto community with a playful, bold style.\n\nThe project offers holders unique rewards and an engaging experience. With Sui’s low fees and high speed, PepeCally is perfect for fast, accessible trading, allowing everyone to be part of the latest crypto phenomenon.",
    "iconUrl": "https://images.hop.ag/ipfs/QmSS6gMkGkoSXCZkN2F7LmxeeN3gsbLyPV5ugbVDYozNRr",
    "decimals": 6
  },
  "0xb1b7ecfe0ae36fc078560a30212e5f5ee86459ebce027762bf74972d4415232d::hopdog::HOPDOG": {
    "name": "HopDog",
    "symbol": "HopDog",
    "description": "HopDog",
    "iconUrl": "https://images.hop.ag/ipfs/QmSD5znzbTdpfxBFaUqLG95hHAmzV9WqbPut1vSNee8YJv",
    "decimals": 6
  },
  "0xc9146981ef058c72263570fa818a75fc7a6d26b85c9095a30fe0354920ddd0c8::cat::CAT": {
    "name": "HOP CAT",
    "symbol": "CAT",
    "description": "First  CAT on http://hop.fun",
    "iconUrl": "https://images.hop.ag/ipfs/Qmen8LhMSiPa1Sy4VX3UjzvDY3XuRGaR2vzsQFWAHYKoyF",
    "decimals": 6
  },
  "0x42d373163f9e4caa762bbb83350ae581b5781aa7bd27c34345455b2b2758db0c::wdog::WDOG": {
    "name": "wrapped dog",
    "symbol": "wDOG",
    "description": ">>>.<<",
    "iconUrl": "https://images.hop.ag/ipfs/QmVAQDZHkqzaWotRwut3fjw1WAgRByiyqPVFMdWExzFExy",
    "decimals": 6
  },
  "0x165814c3a252d8512a9572efebd600bbd83ce984fabdadbd2e869b6eb642cf54::hopcat::HOPCAT": {
    "name": "HopCat",
    "symbol": "HopCat",
    "description": "First cat on Hop Fun",
    "iconUrl": "https://images.hop.ag/ipfs/QmRa3XL5YZQhDzQQyyEzCH1Q9aGaDEGe6N1TXkwyeNDRtP",
    "decimals": 6
  },
  "0x129c9d98e9354dd2210d4c8c1be0ca5c8e0e8a24945ed66422a0492309db3d4d::swoop::SWOOP": {
    "name": "swoop",
    "symbol": "swoop",
    "description": "swoop",
    "iconUrl": "https://images.hop.ag/ipfs/QmYGcjMQEVqv5DnLNKKH4gYuLi457vheuadKYpgbFhCN4a",
    "decimals": 6
  },
  "0x22059dd5a5fbd6dcb19f8fe5c3497277a764b71ce778f89b78bead4a881389e4::dragon::DRAGON": {
    "name": "Totem-Dragon",
    "symbol": "Dragon",
    "description": "Dragon is a mythological creature with great symbolic significance, which is often regarded as a symbol of power, dignity and auspiciousness. Unlike Western dragons, Chinese dragons are usually depicted as long, snake-shaped, four-clawed and capable of controlling water, rain and wind.",
    "iconUrl": "https://images.hop.ag/ipfs/Qme4VYBTkrRKRD2qt29QTK5rg5h5jx9BGAcHzE2CgoUAWc",
    "decimals": 6
  },
  "0x5c931e75a2aebbed6b314a80e722cd9ba8696c5a59c2db8143ed6845a69dd476::hopcat::HOPCAT": {
    "name": "HopCat",
    "symbol": "HopCat",
    "description": "First cat on Hop Fun",
    "iconUrl": "https://images.hop.ag/ipfs/QmWBx34wFTPcn8u2RbT74MgrKZFcbssptPYZWZhHQM3oVt",
    "decimals": 6
  },
  "0xc4f18516431908968571361a85d3273a08183fef69ee735f5068d83c94bbb0c4::bubl::BUBL": {
    "name": "BuBL",
    "symbol": "BuBL",
    "description": "Bubbling on SuiNetwork to make frens",
    "iconUrl": "https://images.hop.ag/ipfs/Qmbf752EhgUbtLCph7R9XuGYiWQZKdyCaA8Bo5afpwuNiY",
    "decimals": 6
  },
  "0xfea33f6aa98650530e61347fe4e34b31d1cf4b3fbb02b0c09f110c950d873205::horse::HORSE": {
    "name": "Suihorse",
    "symbol": "HORSE",
    "description": "The chillest horse in the sui",
    "iconUrl": "https://images.hop.ag/ipfs/QmfLLvTxyZuCXGujr532hX89vK5rxqkyD5DDZxrwaDkThm",
    "decimals": 6
  },
  "0xda2b55386cd1124a1d2da3dbc721496a5135a67480be9aab26da0b9579884d69::hopfrog::HOPFROG": {
    "name": "HOP FROG",
    "symbol": "HOPFROG",
    "description": "\"Victorious warriors win first and then go to war, while defeated warriors go to war first and then seek to win.\" Sun Tzu, The Art of War.",
    "iconUrl": "https://images.hop.ag/ipfs/QmdqJVpHZhhdnRxEDZtdwnJJyCXv6cpssE3RnxpYyeAQXN",
    "decimals": 6
  },
  "0x90fa5fa5f284536c80f01e909d46e0c9fdedc39aa921251ac5abc17f74bfdb1f::pnut::PNUT": {
    "name": "Peanut",
    "symbol": "pnut",
    "description": "Peanut - Favorite squirrel",
    "iconUrl": "https://images.hop.ag/ipfs/QmPv1TvN9U3fT1WHzueqouhcSu8oTPiHnmA7tFtGEuDztw",
    "decimals": 6
  },
  "0x22c03493c98b3b2c9c462610577e4fa5cda55283a771f74acd2c01e02388244a::hopcat::HOPCAT": {
    "name": "HOPCAT",
    "symbol": "HOPCAT",
    "description": "VERY RICH FATTY CAT",
    "iconUrl": "https://images.hop.ag/ipfs/QmTUTNJb9XaqLn8cduhndAGsDKQK1qYNsSQ3ZXLZNp21kQ",
    "decimals": 6
  },
  "0xbf0c4bab1e98b8e608f9abf54604aecc4f2b5daf7b3fa7c2a042e794e0d725e7::hosu::HOSU": {
    "name": "HOSU",
    "symbol": "HOSU",
    "description": "HOP X SUI = $HOSU💧",
    "iconUrl": "https://images.hop.ag/ipfs/QmVrHtwSLeNZB7oWKVqtwMMEunGcgkTwFCeEhbfWZPrD26",
    "decimals": 6
  },
  "0x8bab77656031938088a8e039b21748a6cf90e3b8f66c68d374d248a190b7ec14::swoop::SWOOP": {
    "name": "Swoop",
    "symbol": "SWOOP",
    "description": "Swoop da poop",
    "iconUrl": "https://images.hop.ag/ipfs/QmYGcjMQEVqv5DnLNKKH4gYuLi457vheuadKYpgbFhCN4a",
    "decimals": 6
  },
  "0xdcd3cfccf1c84154a50e0079035a34aefc409df8239f5cfe7f275821911a6d03::rex::REX": {
    "name": "Suirex",
    "symbol": "REX",
    "description": "Suirex is a cute, fun-loving, dino on the Sui blockchain. Suirex doesn't stay cute and small forever though. Suirex grows, and he grows fast. Nobody has ever seen just how big rex can grow, which begs the question. Just how big can he get?",
    "iconUrl": "https://images.hop.ag/ipfs/QmdHT3ngjCqfn7wq8n668cnrqkaJ8YFQwt5GAghiQ9STuq",
    "decimals": 6
  },
  "0xff6d855f2876cd5b5bfaa7fcbbb9721f447290606699a6207b4a7757e3eea0ae::sgf::SGF": {
    "name": "Smoking Sui Giga Fish",
    "symbol": "SGF",
    "description": "No soy larps. No jeets. Only $SGF. What are you waiting for? Join us.",
    "iconUrl": "https://images.hop.ag/ipfs/QmdF5cs4EQtKrFzaWw6HJjmn3Xg35N5xxK6vDS25F39e1j",
    "decimals": 6
  },
  "0xd5fcf6a2947411e145a01e31cf97d43d13d9cd37b7cac2bb3296f7539ebaaf4a::rex::REX": {
    "name": "suirex",
    "symbol": "REX",
    "description": "I'm Rex. W're all Rex",
    "iconUrl": "https://images.hop.ag/ipfs/QmUzqQG6zVXzQ3moLjrFhb3MnrAG2ufmKS6emzd7L2xNqn",
    "decimals": 6
  },
  "0xf3bbbbe2104872a5258932fbfc11b0c6fc929c54347455ec91ab4d5d734ef570::pnut::PNUT": {
    "name": "Pnut",
    "symbol": "Pnut",
    "description": "Pnut",
    "iconUrl": "https://images.hop.ag/ipfs/QmPMsM4mssGg77fgKnE9mW58f6xaDFiPnm9YLfgc9HGay9",
    "decimals": 6
  },
  "0x0fbaa586d87caa19c90fd1678ec5d3e37c88fee76e1e09f4f08f5a9207735da0::fpnut::FPNUT": {
    "name": "First Pnut on Sui",
    "symbol": "FPNUT",
    "description": "First Pnut on Sui",
    "iconUrl": "https://images.hop.ag/ipfs/QmUkmJW32gqRkKoybQgH4h6XH5y1SvsotRUA3JhiGDwJkH",
    "decimals": 6
  },
  "0xf4a61fc117400c8456a90ff6f5a2d3064a8082c7c675f35f0c662e5dc56fbf92::rex::REX": {
    "name": "SUIREX",
    "symbol": "REX",
    "description": "Something big and blue",
    "iconUrl": "https://images.hop.ag/ipfs/QmdHT3ngjCqfn7wq8n668cnrqkaJ8YFQwt5GAghiQ9STuq",
    "decimals": 6
  },
  "0x8ff9243d747f3248ceee3e55f916561e8e3d5e29394ba3e608d9b2329d93d1ec::pnut::PNUT": {
    "name": "Peanut",
    "symbol": "PNUT",
    "description": "The squirl that died :(",
    "iconUrl": "https://images.hop.ag/ipfs/QmbaiLBSQCXtxLAh6hPj1PJNFVLvsCo7w2Mhtp2AF4HBEH",
    "decimals": 6
  },
  "0xf56a4b3873046e4629e06a2a30b7a26842906fbf5b0d3e1c03daca94fec0afe7::hopdoge::HOPDOGE": {
    "name": "HOPDoge",
    "symbol": "HOPDoge",
    "description": "First Doge ON HOP",
    "iconUrl": "https://images.hop.ag/ipfs/QmdZP6M6z641fz4jynew9UjTVjrC4B1mBo8M4ZuA3hyL85",
    "decimals": 6
  },
  "0xaf14aa0eff34d85466b95a4e05c305755acf1d8a6bf5d46496823d5056906682::suirex::SUIREX": {
    "name": "Suirex",
    "symbol": "SUIREX",
    "description": "I’m Rex, we’re all Rex….",
    "iconUrl": "https://images.hop.ag/ipfs/QmdHT3ngjCqfn7wq8n668cnrqkaJ8YFQwt5GAghiQ9STuq",
    "decimals": 6
  },
  "0xd4ca972d2b8e1df1067ad2846cc8814d96c04fadf836d985c935f8e177f3fcf0::frog::FROG": {
    "name": "Frogger the FROG",
    "symbol": "FROG",
    "description": "Coolest FROG on SUI",
    "iconUrl": "https://images.hop.ag/ipfs/QmeqJeebVwbaen94yQy5cKp66w5a2XD8bRYHt99psyqf3E",
    "decimals": 6
  },
  "0x2be6e50bb0552ffaafc67b402620450ec892080a71557b9c1addf7ee8d11ad09::poring::PORING": {
    "name": "Poring",
    "symbol": "PORING",
    "description": "$PORING - The true meme token on $SUI blockchain with fair distribution",
    "iconUrl": "https://images.hop.ag/ipfs/QmRqMXgfJPNEJ2S4uXsAyAGDmcseXYeJ7CGqfKABZoBddR",
    "decimals": 6
  },
  "0x2b8f5741cd94da15593cdaedb950af922abc5f4e60eae008e188da0835716b8e::rex::REX": {
    "name": "suirex",
    "symbol": "REX",
    "description": "RREX",
    "iconUrl": "https://images.hop.ag/ipfs/QmYdBcFuBEcfycWePmayw4Tb3fwDa11GXYY389MNTpmh8P",
    "decimals": 6
  },
  "0xf11dc1c9b737aa040ef42f521cb79306a51330156e42852949433c5214ec10a7::fckh::FCKH": {
    "name": "FUCK HOP",
    "symbol": "FCKH",
    "description": "FUCK HOP",
    "iconUrl": "https://images.hop.ag/ipfs/QmSt9Demj7SpA3zPL5wRc5Y4GKgd3xnFXAjyEE9uNGRnMy",
    "decimals": 6
  },
  "0x530e0e47ae51c083277e85cbf8e8a9f6ccdf9b3a5afb9f8d494659bb5f0727fe::aaa::AAA": {
    "name": "aaa Hop",
    "symbol": "AAA",
    "description": "the first token on hop",
    "iconUrl": "https://images.hop.ag/ipfs/QmXC4qFVCAXdFqfmbMJVqsttrw6A6MMetmKZ1MZecedNH4",
    "decimals": 6
  },
  "0xa17ff626e4d5fc136fad6b3511c9a27254e68b5040223136a06a22f537902e3d::seanut::SEANUT": {
    "name": "Seanut The Suirrel",
    "symbol": "SEANUT",
    "description": "RIP Peanut",
    "iconUrl": "https://images.hop.ag/ipfs/QmPncTECroMSa7Jxth4R69QNWwssKFFLrQ5Nvh3rFqBiJP",
    "decimals": 6
  },
  "0xa0c60c3f0c88eb7b09de07b74bd75581ff86ee9b176863d1d3fd617ebba51e57::fuckhop::FUCKHOP": {
    "name": "fuckhopfun",
    "symbol": "FUCKHOP",
    "description": "FUCK HOP FUN!",
    "iconUrl": "https://images.hop.ag/ipfs/QmQvXjD3avEQMvYHXK9qGPA4ixefgNqQXLfyd5PiLz3y35",
    "decimals": 6
  },
  "0xbd8ca372ed85a5b15ceac2ab771ea0ea65940855282edbd5aba52acd746e70ff::suin::SUIN": {
    "name": "SUI NIGGA",
    "symbol": "SUIN",
    "description": "DEXes over CEXes",
    "iconUrl": "https://images.hop.ag/ipfs/QmUH69sCoZCiubx9NyHQjgu787tPk44oVDZeE4VPbUCb6B",
    "decimals": 6
  },
  "0xc2716c2d45c8e2f0ad108224907e92709a964220cf3ef46fb81f4f7ed3d4dad1::cat::CAT": {
    "name": "HopCat",
    "symbol": "cat",
    "description": "Did you see who knocked on the door?\nThis is the HopFun \n@HopAggregator\n that will launch on October 12 and we're going with it",
    "iconUrl": "https://images.hop.ag/ipfs/QmWBx34wFTPcn8u2RbT74MgrKZFcbssptPYZWZhHQM3oVt",
    "decimals": 6
  },
  "0xaab1d64c945163c29aa38ae2a9e565d542fee155d7ae4c7f6666e7e8f85fadab::swoop::SWOOP": {
    "name": "Swoop",
    "symbol": "SWOOP",
    "description": "🐤💧 Woop Woop Woop 💧🐤",
    "iconUrl": "https://images.hop.ag/ipfs/QmYGcjMQEVqv5DnLNKKH4gYuLi457vheuadKYpgbFhCN4a",
    "decimals": 6
  },
  "0xb47d4300fd7955b471bccf622e94efe6275f423c1a9b8d9ffea5b670b0d12fd1::hopdog::HOPDOG": {
    "name": "HopDog",
    "symbol": "HopDog",
    "description": "HopDog",
    "iconUrl": "https://images.hop.ag/ipfs/QmSD5znzbTdpfxBFaUqLG95hHAmzV9WqbPut1vSNee8YJv",
    "decimals": 6
  },
  "0x930294bc1273f445091e127dbdd02b5992e936e35bae7245f57528a579a2ac4e::pnut::PNUT": {
    "name": "Peanut The Squirrel",
    "symbol": "PNUT ",
    "description": "PEANUT FOREVER. FUCK THE NYSDEC",
    "iconUrl": "https://images.hop.ag/ipfs/QmSfHV3Gx7MHA6Lkvbuu8KTrNe6YmDU2kGUxfc5FhdGY3B",
    "decimals": 6
  },
  "0xc73ba7938f026985aba56513621d70a7e046ad327db687466e47b6c336faae0d::hopdog::HOPDOG": {
    "name": "HopDog",
    "symbol": "HopDog",
    "description": "HopDog Official",
    "iconUrl": "https://images.hop.ag/ipfs/QmSfGqkPkqN5dqAY3Ds26wtXNLpJ66LgZXHwd8PZuizxTa",
    "decimals": 6
  },
  "0xd08ae76aabdc499721cf64b200885054cd4f30e6a77daf285dfa362f6e100f21::nut_4_trump::NUT_4_TRUMP": {
    "name": "Nut4Trump",
    "symbol": "Nut4Trump",
    "description": "Nut4Trump",
    "iconUrl": "https://images.hop.ag/ipfs/Qmd1PepZLWW7pLiHxNgHhHBU2Me35Bkh7sueumZfLLYYcL",
    "decimals": 6
  },
  "0x8fb494ab5423bff8de3772c1c0b87550c6e59c8ece274d4367c97e3481255df7::suiyan::SUIYAN": {
    "name": "Super Suiyan",
    "symbol": "Suiyan",
    "description": "SUIYAN",
    "iconUrl": "https://images.hop.ag/ipfs/QmQ6qeoySt1gDZoYWNWr7WZAehT3NaaAMVVbXLnPkaW4NF",
    "decimals": 6
  },
  "0xc6984cb3f4a1fb638d9c09bbe4997f3a8c909f9fa9b25adfc6c6b36525a77ebf::dory::DORY": {
    "name": "DORY",
    "symbol": "DORY",
    "description": "Dory is a Regal Blue Tang feesh",
    "iconUrl": "https://images.hop.ag/ipfs/QmeYftaYw3vhjxrmvDJ5YhpvqPqsNH2t6tPmdsRcmMdtYP",
    "decimals": 6
  },
  "0xb62c5dd33f051481d7c6b48c21378d0d7d1ddc5769e6a1f5d43bbf9fde4e06a2::peanut::PEANUT": {
    "name": "PEANUT",
    "symbol": "PEANUT",
    "description": "RIP PEANUT",
    "iconUrl": "https://images.hop.ag/ipfs/QmQzRSgaAbP6Hpk9qwnxzYTJdRpKKJ6HUCo3ENaxuqXCob",
    "decimals": 6
  },
  "0x4831b3f34faee0ea88fcce4564d8bcdb427d4cb7d4641a594272a95d285d1a71::hopper::HOPPER": {
    "name": "HOPPER THE FUN",
    "symbol": "HOPPER",
    "description": "HOPPER CAME TO HOPFUN THE MOON AND BEYOND",
    "iconUrl": "https://images.hop.ag/ipfs/QmRbRssWarD6BpGu7HgMvvKrR8zwwuFMFgFp1ZS4sQx2VM",
    "decimals": 6
  },
  "0xb53b0ada678c2b0e722ee96231cbd2c34e79829dafeca6bd84a4c460b6a8e275::elphnt::ELPHNT": {
    "name": "Suilephant",
    "symbol": "ELPHNT",
    "description": "A Blue Elephant Casually Pleasuring Themself on Sui",
    "iconUrl": "https://images.hop.ag/ipfs/QmcB9C7uDJon9eeq4Wx1w3Lp7e9BWuvKfEFe2H2zMMHk37",
    "decimals": 6
  },
  "0x9ebc222dcb29977b3c7f059eca080af133f37a41cbd55733825eb36c3819f73f::onepepe::ONEPEPE": {
    "name": "One Pepe",
    "symbol": "onepepe",
    "description": "Epic adventures of One pepe crew",
    "iconUrl": "https://images.hop.ag/ipfs/QmdEPyZRxpvRig8qNvyrCnSQcf3Hvadq8TvnX7sJt1Jts9",
    "decimals": 6
  },
  "0x55ff96aba7ffab8eac592f4fb550fa5d969adb351cda8905f30f69d287a9876c::mew::MEW": {
    "name": "MeowWooF",
    "symbol": "$MEW",
    "description": "Join the MEW Pack to be apart of the fastest growing community on SUI! \nTO DA MEWNNNN!",
    "iconUrl": "https://images.hop.ag/ipfs/QmQzsgpFsSLA1XV3yeAS2j2MxfdLUt4qR9CaGnHgVWAiRA",
    "decimals": 6
  },
  "0x8462c4e7093d258c5f7a12def6ae44b95de8ba2f105412c3497c8b0c18d4febd::hopper::HOPPER": {
    "name": "The Rabbit",
    "symbol": "HOPPER",
    "description": "Hopper the rabbit from Moon, bouncing high to a cosmic tune!",
    "iconUrl": "https://images.hop.ag/ipfs/QmZ6rdWkRLMFGzceL6YtgabAVo3c33BkxrKWThoQX3necS",
    "decimals": 6
  },
  "0x79d047d9d174022088cc6067c442c76f6ea45b8e5b59f0826aa89f37d294f74d::bocat::BOCAT": {
    "name": "Book of Cats",
    "symbol": "BOCAT",
    "description": "Book of Cats is a unique token for cat lovers, featuring a collection of cats securely stored on our site.",
    "iconUrl": "https://images.hop.ag/ipfs/QmNpqTpCjdkZXwo7JhTF1wffLUgsX2xJrt6pioHTus5Bz9",
    "decimals": 6
  },
  "0x51f8c180d805433bdb165af77e6ce6fbbadb6f06b3a1ec17859ca6395bf1ce2c::gfbf::GFBF": {
    "name": "My Girlfriend's Boyfriend",
    "symbol": "GFBF",
    "description": "How do I please my girlfriend? Find her a boyfriend!",
    "iconUrl": "https://images.hop.ag/ipfs/QmWGnNh9RorEvoJZTPJN56qwzYo1ywAoKaVgjbfoAFR5jf",
    "decimals": 6
  },
  "0x0a86b64277925b83fbb339572c33a672b493bd3c9d6bde1f602ed48b04f71299::hpepe::HPEPE": {
    "name": "HopPEPE",
    "symbol": "HPEPE",
    "description": "PEPE IS COMING IN DA HOUSE. READY TO RUMBLE?",
    "iconUrl": "https://images.hop.ag/ipfs/QmZMTnCNYincJTTtNvxptHGEnB36F336C544Re5Zjo2QLj",
    "decimals": 6
  },
  "0x840bc6e0ace6985a570ed3794e96877c7dd476437b73b9ea28b3bbc54b7b5eeb::sngui::SNGUI": {
    "name": "SnugglyUi",
    "symbol": "SNGUI",
    "description": "\"Meet SnugglyUi! 🐱\"\nStarting as a tiny kitten in the crypto jungle, SnugglyUi has BIG dreams—to grow into a huge, hype-worthy feline! With a heart full of fluff and a paw on the pulse of the meme world, SnugglyUi’s only wish is to be adopted by kind-hearted folks like you. Stick around, and watch this kitty climb to purrfection! 🐾",
    "iconUrl": "https://images.hop.ag/ipfs/QmZ45Hb6wKvJ9S11e11HFfSCDymT5ayLehoLQFcuKhm9R4",
    "decimals": 6
  },
  "0x1da081b11b3d8ed1ee3b02ae405cee765ba7b8b5ae8929a4551ad1c891c8971d::devin::DEVIN": {
    "name": "Devin",
    "symbol": "DEVIN",
    "description": "i am devin.",
    "iconUrl": "https://images.hop.ag/ipfs/QmQvnd1wRvUpra4xuFm3j8xwz68tTwc2XaSFUzxJnggUT3",
    "decimals": 6
  },
  "0x18bdb214d95837aa8e66da069cf2685db34cc7b01dea311829c1464b420484e5::pnut::PNUT": {
    "name": "RIP Pnut",
    "symbol": "Pnut",
    "description": "Pnut coin",
    "iconUrl": "https://images.hop.ag/ipfs/QmSnehJ1Ldn5XQjcxC5veGDVGrd5bG9bvLBz33UvdhveoE",
    "decimals": 6
  },
  "0x93f4f805b3ff4d9432e6902786c69d1eedd77a80ee13e6ae6631759350e257ed::bvr::BVR": {
    "name": "BeaverSUI",
    "symbol": "BVR",
    "description": "$BVR — The sharpest marine architects, the Beavers, now thrive on the most efficient blockchain: SUI.",
    "iconUrl": "https://images.hop.ag/ipfs/QmNRrzi7j2Lw4hBZfkFrg17jymuZqSMQa637Bz81xFE7Y3",
    "decimals": 6
  },
  "0x99b8701ebe1f1969477c01fac0f6a08583313f78601d30b85d84a016732b97e3::pnut::PNUT": {
    "name": "Peanut the Squirrel",
    "symbol": "Pnut",
    "description": "It’s our job to help save peanut the squirrel! Sign the petition!",
    "iconUrl": "https://images.hop.ag/ipfs/QmcizBAsLA1EiRKKYGveXjZLGSfDHWBad5jJr6SZ6bBEtz",
    "decimals": 6
  },
  "0x76fae3529733c9ee7617e2cbe1371d6334c53ebdbdb8106e4bcbad98cb385464::siu::SIU": {
    "name": "CR7 ON Siiiuuuuu",
    "symbol": "siu",
    "description": "CR7 LFG",
    "iconUrl": "https://images.hop.ag/ipfs/QmRPntw8tKpYPqtcyAe3df2vLNWX5Wb7gCCZZz7CrDEBi8",
    "decimals": 6
  },
  "0x42078609783b30f9324ea771e26bd460f8efd1fb2691144e8ba41d85329a7020::aibear::AIBEAR": {
    "name": "AI Bear",
    "symbol": "AIBEAR",
    "description": "\"Bears love AI too.\"",
    "iconUrl": "https://images.hop.ag/ipfs/QmWDKkhGNZqAopFdXJWvjFrhvdy8sqYg7WspWCd3jrekoK",
    "decimals": 6
  },
  "0xf36a8f2dd17c98ea8e084ed0a45b68b55a21e5df5822fbd0d32ad54aa51ab0c2::ghost::GHOST": {
    "name": "GHOSTBUSTERS",
    "symbol": "GHOST",
    "description": "Whatcha Gonna Buy?",
    "iconUrl": "https://images.hop.ag/ipfs/QmdYVstP8gYJ9qcT6LvuNtaXUXd1oemkKPfvxi8XNBNQm9",
    "decimals": 6
  },
  "0x6dd9c6cab894913c129207f49890e4250054230bdb680e6a594118f303ef615a::nut::NUT": {
    "name": "PNUT",
    "symbol": "NUT",
    "description": "R.I.P.",
    "iconUrl": "https://images.hop.ag/ipfs/QmcwCF29pUqfQWZ4su5QAKWTFP8Hd6f4cczUSck6heZMC7",
    "decimals": 6
  },
  "0xd94502d3c1f942d0adc3b3d3a5a39fd5d788f3ab75b2a2a228b7b61269697bc1::luce::LUCE": {
    "name": "Luce",
    "symbol": "Luce",
    "description": "The Vatican has unveiled the official mascot of the Holy Year 2025: Luce (Italian for Light). Archbishop Fisichella says the mascot was inspired by the Church's desire \"to live even within the pop culture so beloved by our youth.\"",
    "iconUrl": "https://images.hop.ag/ipfs/QmZDVFcDL7L4uHZ5iARsnAxdqew8XGas2DirU6Pc4iB9jw",
    "decimals": 6
  },
  "0x4b91de2b2cb2ee7ea9c673680c78a0ca452a154283badd485bfd9ccea25e585a::hep::HEP": {
    "name": "Hep",
    "symbol": "Hep",
    "description": "0x43e9045850072b10168c565ca7c57060a420015343023a49e87e6e47d3a74231%3A%3Ahoppy%3A%3AHOPPY",
    "iconUrl": "https://images.hop.ag/ipfs/QmTiWn4JR3nJGz7P2FhJBQW1B6A3NMy8KeSxmE857dnRJK",
    "decimals": 6
  },
  "0x38ab5fbb5a9ddb0e4f5d441a12742ae6425dd3060b1ef80c53d22f24f6e0ede1::pnut::PNUT": {
    "name": "Peanut the Squirrel",
    "symbol": "pnut",
    "description": "It’s our job to help save peanut the squirrel! Sign the petition!",
    "iconUrl": "https://images.hop.ag/ipfs/QmYmE7pTAefxPr8mT15MD9X9seZvSTPWDM2sdFfNiKECHU",
    "decimals": 6
  },
  "0x8a814a826e9e29777a4874abc5007560052ef39a236373795cf741113c8a228c::uni::UNI": {
    "name": "Sui Founders Dog",
    "symbol": "UNI",
    "description": "Sui Founders Dog",
    "iconUrl": "https://images.hop.ag/ipfs/QmTtssCS6nqc6Rr7oyuqCjtNPthe5RF3jeugVase7uQw59",
    "decimals": 6
  },
  "0x4b39a93c6f46a876d86f5e967fc8f680fade10293f970aea998fe13946d39bb6::jugo::JUGO": {
    "name": "Jugo",
    "symbol": "JUGO",
    "description": "Jugo Para Todos, the meme token that brings a burst of flavor and creativity to the SUI network! Join a vibrant community that celebrates laughter, fun, and the spirit of sharing. With every transaction, you’re tasting the joy of meme culture and making waves in the digital landscape!",
    "iconUrl": "https://images.hop.ag/ipfs/Qmf1J2N3PkmFogopZZK8UYiTw5TFUd4uzrpcKSmRgnFMrQ",
    "decimals": 6
  },
  "0xf77b745efcdbb2f266b3552637c93f050be22a3eb186c450cb4861ae8b684f63::suicat::SUICAT": {
    "name": "suicat",
    "symbol": "suicat",
    "description": "sui memes",
    "iconUrl": "https://images.hop.ag/ipfs/QmQnbVQ6ft5A9xvRzbcuBHnvvtqbi5V7HpvpFU9zygLvSa",
    "decimals": 6
  },
  "0xcca2f80561ed91585313340a7bf5a9c59ff6c73c6e610e8fbfbaf40b1f7ec8e0::hippo::HIPPO": {
    "name": "sudeng",
    "symbol": "HIPPO",
    "description": "suideng",
    "iconUrl": "https://images.hop.ag/ipfs/QmZSsgvbAXhXFBHo4wj9jchdfNhFmLWuHMjkFUcSRXbqup",
    "decimals": 6
  },
  "0x312bb4ede6409a35027f0f6313b0b58fb91086a92b57183e689bc5e7edc2d733::color::COLOR": {
    "name": "Sui Color",
    "symbol": "Color",
    "description": "Color on sui，give sui more color,welcom to sui community",
    "iconUrl": "https://images.hop.ag/ipfs/QmP5okeHmsKCLkWAbWA3C34BPcqUDRWvPYDyLXVTPuEFWL",
    "decimals": 6
  },
  "0xeb44d070a535e3d2e3daf467b60cfbdc5782d5d3cff823f3596a83162c8c3946::peanut::PEANUT": {
    "name": "peanut",
    "symbol": "peanut ",
    "description": "peanut",
    "iconUrl": "https://images.hop.ag/ipfs/QmPAo1rx8PpMhqe1RjoK8J9TNUUn9NYyxVKsVzZWMmE4bP",
    "decimals": 6
  },
  "0x2178b6489dea4f059da3d1c0f442317cff0091f024cb5ef10add592153151eb6::pnut::PNUT": {
    "name": "Peanut the Squirrel",
    "symbol": "Pnut",
    "description": "Peanut the Squirrel",
    "iconUrl": "https://images.hop.ag/ipfs/QmbHZYAZM57G6qQz38YJaGhL8NsF3xvMEt7u3hacEJC8Jq",
    "decimals": 6
  },
  "0x0f66680756e9ca86549d3906e4094408f5a9a0abdf9c7d9371feef986d7cd015::hippo::HIPPO": {
    "name": "sudeng",
    "symbol": "HIPPO",
    "description": "HIPPO",
    "iconUrl": "https://images.hop.ag/ipfs/QmTQsPbCirNbwYkMho3V32nkzA6Utxs5rFTrEbipmBnhft",
    "decimals": 6
  },
  "0xcbf8738afe79f682023a8b86e6fc0be1f80f9f04dc172f5f6cb963637d70fca7::lmsh::LMSH": {
    "name": "Lets Make Sui Hob Again",
    "symbol": "LMSH",
    "description": "$LMSH - Making Sui Great, One Meme at a Time! 🚀🤣 Join the movement, bring the laughs, and let’s send Sui to the moon (or at least the meme hall of fame)!",
    "iconUrl": "https://images.hop.ag/ipfs/QmUw5T6gKUTAuZu8seyWHAQtdWZJ3hdBEQKqy6AzcgHRvt",
    "decimals": 6
  },
  "0xeba0a0e113fee87d3f2fca4c1ac5cc863eaee1e80b20c7d4fb31274ef78b93f4::trump::TRUMP": {
    "name": "Donald trump",
    "symbol": "$Trump",
    "description": "Donald J. Trump For President 2024. America's comeback starts right now. Join our movement to Make America Great Again!",
    "iconUrl": "https://images.hop.ag/ipfs/QmTQoF7XcM3tHPjMnRwd7FnBTCHwqWSpkrnrpaASthxFU2",
    "decimals": 6
  },
  "0xc1ef630eea371735ebe514e634ba9e97a2a41ed163f2e3d9726747d3e5efac68::migo::MIGO": {
    "name": "Amigos",
    "symbol": "Migo",
    "description": "Welcome to Amigos, the meme coin that's so laid-back, it's practically horizontal.",
    "iconUrl": "https://images.hop.ag/ipfs/QmXQJRsvAxiLHmv21kdwCEF6vyzWXiVJv1oV84cPUUpa9P",
    "decimals": 6
  },
  "0x0220450ff31fff6bf50c19d44910ad75a07e5f5baad2e2386d03dff06acdaa44::hopium::HOPIUM": {
    "name": "HOPIUM",
    "symbol": "HOPIUM",
    "description": "hopium",
    "iconUrl": "https://images.hop.ag/ipfs/QmeSyknJVii7EtLft5ZnQEgiQAM9TvZvwFHgHLp54dmHtR",
    "decimals": 6
  },
  "0x5c4033c2c6012e53f35de934e6a8485d971b2838c35ef8cd9500bce0aaac3d21::siu::SIU": {
    "name": "SIU",
    "symbol": "SIU",
    "description": "SIU SIUUU SIUUUUU SIUUUU SIUUU SIU",
    "iconUrl": "https://images.hop.ag/ipfs/QmWtAVBvGGHAyDq8Lyb2tTSBptFAdqRRsvxBNsFJMsoAPA",
    "decimals": 6
  },
  "0xb9ae259693607f2f5f15331f144afc62434282a69e95288e63eb4a1b67a8d865::swoop::SWOOP": {
    "name": "Swoop",
    "symbol": "$Swoop",
    "description": "$Swoop is here",
    "iconUrl": "https://images.hop.ag/ipfs/QmXHzRikJ9VzUBwjr61nJaDsSXQyAnhuwuZgoKSVGaK1wE",
    "decimals": 6
  },
  "0xc40a834debb557bff3654011b651564dff76d0d9b76d1e811fd339c3dea747df::test::TEST": {
    "name": "TEST",
    "symbol": "TEST",
    "description": "test",
    "iconUrl": "https://images.hop.ag/ipfs/QmcSDW1ETVsNnDoH9rNv4cCXUKNNFkwY9aum9C7DwHEqKi",
    "decimals": 6
  },
  "0x3a2fc16ace41d8e8f7e06cafcc14247a280c0733f88f84d171fba2da404e379f::hop::HOP": {
    "name": "HOP",
    "symbol": "HOP",
    "description": "Swap on @SuiNetwork for the best price with zero fees.The first token on hop.fun",
    "iconUrl": "https://images.hop.ag/ipfs/Qmb3jiE6CwdhupbcP1W2CSiRETFYVXofRk8ULypq3BCnRX",
    "decimals": 6
  },
  "0x6187f8506f3a0f7aa95190fedbdc6a79b55e34921fba07bfca165d920489f718::kitty::KITTY": {
    "name": "KittyCat",
    "symbol": "Kitty",
    "description": "The cutest Kitty Cat on Sui",
    "iconUrl": "https://images.hop.ag/ipfs/QmPNf7yRRH3tRysSe3r3teM6P64USuk9LPbq523xiihR29",
    "decimals": 6
  },
  "0xf35ef74d98e181946b275c28fe2aa7a93c8afca0e520f45538d5a273eef468de::happy::HAPPY": {
    "name": "HAPPY",
    "symbol": "HAPPY",
    "description": "Step into the Future with Happy’s Meme Coin Web 3 Rewards, Gaming, and NFT Mysteries Await!",
    "iconUrl": "https://images.hop.ag/ipfs/QmXF6bpU6CdL8BFKY9XnCX8x6tqJNWrsRSnbgNN1wZuAx7",
    "decimals": 6
  },
  "0x25a1db6fb543db366097aca1bf5a949536a58835837262cfa79ae4068ede8a72::bubl::BUBL": {
    "name": "BUBL",
    "symbol": "BUBL",
    "description": "BUBL\nin the world of Sui, love is simple. bubl loves blub, and together, they make everything a little more bubbly. this duo is all about bringing joy, fun and a splash of something special. because at the heart of bubl, there's always a little blub",
    "iconUrl": "https://images.hop.ag/ipfs/QmbFzA2pwoSSjs1hr9a5z8p44eb2yAUF47TFpARHUu1orZ",
    "decimals": 6
  },
  "0xf395645ba076c43970796fefe1d22cef496bf714d077844d1034c9c71c770041::pnut::PNUT": {
    "name": "Peanut",
    "symbol": "PNUT",
    "description": "This is for you Peanut 🥜❤️",
    "iconUrl": "https://images.hop.ag/ipfs/QmPAo1rx8PpMhqe1RjoK8J9TNUUn9NYyxVKsVzZWMmE4bP",
    "decimals": 6
  },
  "0x704c08fd53e76585fed40defca833bea8c56bda177152f3687cc104425e29093::swoop::SWOOP": {
    "name": "Swoop",
    "symbol": "Swoop",
    "description": "🐤💧 Woop Woop Woop 💧🐤",
    "iconUrl": "https://images.hop.ag/ipfs/QmYGcjMQEVqv5DnLNKKH4gYuLi457vheuadKYpgbFhCN4a",
    "decimals": 6
  },
  "0x7d006a3536790e555d96b4cb3639c8fcb1088a83846890829fa5aaa55f320326::pnut::PNUT": {
    "name": "Peanut the Squirrel",
    "symbol": "Pnut",
    "description": "It’s our job to help save peanut the squirrel! Sign the petition!",
    "iconUrl": "https://images.hop.ag/ipfs/QmPncTECroMSa7Jxth4R69QNWwssKFFLrQ5Nvh3rFqBiJP",
    "decimals": 6
  },
  "0xc46fe7c7642ab0135543528fe2f483e97e825a72e1e38c4a625ce827034256df::bubl::BUBL": {
    "name": "BUBL",
    "symbol": "BUBL",
    "description": "BUBL  BLUB\nin the world of Sui, love is simple. bubl loves blub, and together, they make everything a little more bubbly. this duo is all about bringing joy, fun and a splash of something special. because at the heart of bubl, there's always a little blub",
    "iconUrl": "https://images.hop.ag/ipfs/Qmbf752EhgUbtLCph7R9XuGYiWQZKdyCaA8Bo5afpwuNiY",
    "decimals": 6
  },
  "0x0298ecface4c8c7c2561ff251e9c6c1c0f3a6cfa7ecc3d16c929d1717d7980f2::trm::TRM": {
    "name": "trump",
    "symbol": "trm",
    "description": "7MnmNHEHgs1gJ9q477xrCpQfL4PDSWfzTZocZdNfpump",
    "iconUrl": "https://images.hop.ag/ipfs/QmVU9jhfdyAVZkchT6Kb4bfT4jdjehHAfsYzmpAR3ZSt7w",
    "decimals": 6
  },
  "0xdd4e4cbca55e0100a614a0bb8ac54e1c84866c3547a23d6b04299797c83b635f::papa::PAPA": {
    "name": "papa",
    "symbol": "papa",
    "description": "the godfather of memes",
    "iconUrl": "https://images.hop.ag/ipfs/QmQLg1baonXESU5sGwQX7BV3BsertZn9JfASGNKhSTwRQJ",
    "decimals": 6
  },
  "0x68efdc04bcbc10712447c359280bfab4fa7ed97fe2bc112621e32309b5e088b6::pnut::PNUT": {
    "name": "Peanut",
    "symbol": "PNUT",
    "description": "Fuck government",
    "iconUrl": "https://images.hop.ag/ipfs/QmT8Zmj2bxeoHmJWKfcZnPGBvL1mVBJtvF58TU9ymuph2A",
    "decimals": 6
  },
  "0x2bb012eac3a6aae4964f4a96fd09687172b98d96338aed127fc406bb460a06b1::milady::MILADY": {
    "name": "milady",
    "symbol": "milady",
    "description": "milady coin",
    "iconUrl": "https://images.hop.ag/ipfs/QmS8tJUtwo8WPzpVSvazPuvDBxNvcHmQPQgfZYKAPZzALE",
    "decimals": 6
  },
  "0xa866581c68a353b8112510c43f86bb67856c850b488819aea6cf0c2f2b0a6834::toast::TOAST": {
    "name": "HOP TOAST CAT",
    "symbol": "TOAST",
    "description": "firts toast on hop",
    "iconUrl": "https://images.hop.ag/ipfs/QmVLxXBdm3ijhZ3mEQZtaX8H3H6SgZdRyCfDfjurndQmHP",
    "decimals": 6
  },
  "0xe634e37136cc9e14469f099b469725912cfbfe7b551fb29fea6735e9a48720f0::hopcat::HOPCAT": {
    "name": "HOPCAT",
    "symbol": "HOPCAT",
    "description": "HOPCAT",
    "iconUrl": "https://images.hop.ag/ipfs/QmWBx34wFTPcn8u2RbT74MgrKZFcbssptPYZWZhHQM3oVt",
    "decimals": 6
  },
  "0x90956b47b4f600dedbc75b0396256b65bb498778871b1533cac14ab440995536::nymph::NYMPH": {
    "name": "Sui Nymph",
    "symbol": "NYMPH",
    "description": "Hopfun is so broken i had to use claude to make the picture",
    "iconUrl": "https://images.hop.ag/ipfs/QmPcVkn5cuLsExwbyBbsBbzZBtjoPLAM6QKDECenQAv2hF",
    "decimals": 6
  },
  "0xac8ebfc7f4176e4805c763f00d7a0bb3fdbaf22c8a665bf503b5ac003b99b3f4::feet::FEET": {
    "name": "Feet JPG",
    "symbol": "Feet",
    "description": "vindictas feet",
    "iconUrl": "https://images.hop.ag/ipfs/QmV7ssAAsbY28aLQPh4kb3rDCKAMYCFEGbM4F1LavmNsuU",
    "decimals": 6
  },
  "0xbe54b6f79a8925c501006e2d68221d799d615baf8ea0296db60654f9adada8f4::piggy::PIGGY": {
    "name": "Piggy Bank",
    "symbol": "Piggy",
    "description": "Piggy Bank (PIGGY) - The Meme Coin for Your Digital Savings!\n\nIntroducing Piggy Bank, the fun and engaging meme coin set to launch on the Sui blockchain via hop.fun! Inspired by the nostalgic charm of childhood savings, Piggy Bank aims to combine community-driven humor with the excitement of cryptocurrency.\n\nPiggy Bank (PIGGY) isn’t just another token; it’s a playful symbol of saving and growth in the crypto world. With its adorable piggy mascot and vibrant community, PIGGY seeks to create a fun and accessible way for users to engage with the blockchain while enjoying the benefits of meme culture.\n\nKey Features:\n\nCommunity-Driven: PIGGY thrives on the creativity and involvement of its community, encouraging users to share memes, ideas, and support one another.\nTokenomics: Designed with a unique reward system to incentivize holding and participation, ensuring everyone can benefit as the community grows.\nSeamless Transactions: Built on the Sui blockchain, Piggy Bank offers fast and secure transactions, making it easy for users to trade and use their tokens.\nJoin the Piggy Bank movement today, and let’s save, laugh, and grow together in the vibrant world of cryptocurrency! 🐷💰",
    "iconUrl": "https://images.hop.ag/ipfs/QmXtBrMNyovsXVKUi2LFEiUPYzNAJ5DsG3fetQ7HWuZzhy",
    "decimals": 6
  },
  "0x3daebd8e13018302a9017602980f7e512c7082f61aa018170464361d4750152d::hopunk::HOPUNK": {
    "name": "HoPunk",
    "symbol": "HoPunk",
    "description": "$HoPunk brings fun with adorable pet cuties.",
    "iconUrl": "https://images.hop.ag/ipfs/QmQdEkttLHPfvBx7ptBhRaPKoTzAG1CUxgBqW9e8GZQWWc",
    "decimals": 6
  },
  "0x456c14348c770ec978a3da32ca1d4f19694b012230378c7c3c0948f23e134159::suiyan::SUIYAN": {
    "name": "SUIYANCOIN",
    "symbol": "SUIYAN",
    "description": "Memecoin supercycle. SUI supercycle. Suiyan supercycle.",
    "iconUrl": "https://images.hop.ag/ipfs/QmWEV2uVXRBB4tBVoWd5s7qbeoZ6S5aVHzDoGdWpytSrBT",
    "decimals": 6
  },
  "0x39dbfaa18163c6586493ce7f7f1bdc75190ef98f64a00a69046d0bbd45b1f916::peanut::PEANUT": {
    "name": "Peanut the Squirrel",
    "symbol": "PEANUT",
    "description": "Peanut (also known as P'Nut) was a pet Eastern grey squirrel that had a popular Instagram account devoted to it. In October 2024, it was seized from its owners home and later euthanized.",
    "iconUrl": "https://images.hop.ag/ipfs/QmNWi2Hk4sPioqDx1MrkSEHVk6hLFtKnkk9Fk38UgwyN8v",
    "decimals": 6
  },
  "0xb121f052e505c0d4aca093fda59d61b3a23694871d38dd2335804e4971485e20::hdog::HDOG": {
    "name": "HOP dog",
    "symbol": "HDOG",
    "description": "First HDOG on SUI",
    "iconUrl": "https://images.hop.ag/ipfs/QmVAQDZHkqzaWotRwut3fjw1WAgRByiyqPVFMdWExzFExy",
    "decimals": 6
  },
  "0x5c76d06037c5f23af4b9bf9eb814364ee9e02bb1122ede5e2b2e58804c1f33b0::bonkman::BONKMAN": {
    "name": "bonkman",
    "symbol": "bonkman",
    "description": "Hello, I am Bonkman.",
    "iconUrl": "https://images.hop.ag/ipfs/QmWn6rXpcFnNjaZzGBaPKDz6yYqnW2H6AE82DHnUPDP4J6",
    "decimals": 6
  },
  "0x7f08e7a88a1e31c34b556d96a74851116e25fc3f97f9fec40c5a80264f218e76::ghost::GHOST": {
    "name": "GHOSTBUSTERS",
    "symbol": "GHOST",
    "description": "Whatcha gonna buy?",
    "iconUrl": "https://images.hop.ag/ipfs/QmdYVstP8gYJ9qcT6LvuNtaXUXd1oemkKPfvxi8XNBNQm9",
    "decimals": 6
  },
  "0xf05381d5c15833ddec7c4060e3b9bf3dded29cfa81d034ab0f693e68bcd9f9db::rachop::RACHOP": {
    "name": "Rachop",
    "symbol": "RACHOP",
    "description": "Rachop is one of the first projects of Hopfun. Our lead characters, Rachop the pirate raccoon and his loyal friend, HoP the Hopfun mascot İnvite you to join them on their thrilling adventures",
    "iconUrl": "https://images.hop.ag/ipfs/QmcU72eDV2E2SnH9kbTTcNunjXrRUqDJBKMteLuxtJjXPt",
    "decimals": 6
  },
  "0x6cd11a84dd2806694370d2ff0c9a2f01da6e0a7e3d35d1ef14a28bb2d9a5059c::filigranio::FILIGRANIO": {
    "name": "FILIGRANIO",
    "symbol": "FILIGRANIO",
    "description": "FILIGRANIO FOR THE CULTURE",
    "iconUrl": "https://images.hop.ag/ipfs/QmZP4hHqtgp37qrmcv5tPPAcBS4Mef9siLv6KxpQtMBtB8",
    "decimals": 6
  },
  "0x9ef51c8fe5b7f98331d5c7638c334ba0778b6e07ab030c4c52dbf7ddf16bc0ec::chop::CHOP": {
    "name": "CHOP SUI",
    "symbol": "CHOP",
    "description": "WAKE UP",
    "iconUrl": "https://images.hop.ag/ipfs/QmR6qQCq4Zd5fmkLZt9uiaS9CauFVVya5iwkCGpdHRBFMA",
    "decimals": 6
  },
  "0x52f00e0d15688d1ed29dcc74bccde9181115e8420c92ae67d870c0cee01c7930::suitama::SUITAMA": {
    "name": "Suitama",
    "symbol": "SUITAMA",
    "description": "$SUITAMA is the main protagonist of the series and the titular One-Punch Man on Sui Chain.",
    "iconUrl": "https://images.hop.ag/ipfs/QmXDUpjaVwf6PLX4xtTK8j11ZS7CWBmSDRYgQM4PLrJLa6",
    "decimals": 6
  },
  "0xe185b86cec4fbbae7b91344eaf0035fd3496988b7b25c85b778b316e2d073c7a::ollie::OLLIE": {
    "name": "Rockstar Otter",
    "symbol": "Ollie",
    "description": "Ollie To A Dolly $$$",
    "iconUrl": "https://images.hop.ag/ipfs/QmYfPDSwukxiDGbfiKBm1Tg7qLYJvR76Am5yUAHF2vk58Z",
    "decimals": 6
  },
  "0x01915890c2bb3ee12539b59a4a60ef051b339fc27fd27f866bb6d8ad828562ca::suno::SUNO": {
    "name": "Sui Suno",
    "symbol": "SUNO",
    "description": "Suno - the rebel coin with zero chill. Powered by pure rage and headbangin’ energy, Suno’s here to give the crypto space a roar it won't forget. Forget being nice—this coin’s here to smash, stack, and scream its way to the moon. Ready to rage?",
    "iconUrl": "https://images.hop.ag/ipfs/QmTSvUQgiHa3HbDRN4BxSA1Jfi1LzTMCpdPMj2BB7bhKjz",
    "decimals": 6
  },
  "0xe1c47393d16dcb2a300ad4675c4c1c24f8f5a85840c0313a61191b4fca491a1a::pnut::PNUT": {
    "name": "Penut",
    "symbol": "PNUT",
    "description": "Peanut lives on",
    "iconUrl": "https://images.hop.ag/ipfs/QmNg6ALyfsiLZPnkFBjhwN8LGxv9gEanrZTHwvSc9EdwPL",
    "decimals": 6
  },
  "0x40e5745dc262b3fad8de972eb34d1deeacd77c020469a0eff154cd92c61818eb::suilander::SUILANDER": {
    "name": "SuiLander",
    "symbol": "SUILANDER",
    "description": "Lander family is arriving to SUI network...meet 'em all!",
    "iconUrl": "https://images.hop.ag/ipfs/QmRLjAX2Fdue4K95hT6EBpBwMpZHZx5qP77upYS9K9W6Ci",
    "decimals": 6
  },
  "0x8e24e6a99f1d7a7ad03dd17e2cc564cfea222f2c54fbbd8cbf119e587a523896::cult::CULT": {
    "name": "NOT A CULT",
    "symbol": "CULT",
    "description": "is it bad to start a cult ?",
    "iconUrl": "https://images.hop.ag/ipfs/QmT2tzsVVnU6uMYHQL5GHLZfkp9Yw3rB51agq5VkXr9Cpa",
    "decimals": 6
  },
  "0xee95161bd81f6476a9bac0c92da54c14e208ec8356893c2e24e23df08883e2c2::hopdog::HOPDOG": {
    "name": "Hop Dog",
    "symbol": "HOPDOG",
    "description": "Sui's and Hop's best dog!",
    "iconUrl": "https://images.hop.ag/ipfs/QmRKatJ6xRmwJ3NaK24UT4WwU7YqCXTWQbTf4AZGnHPn8R",
    "decimals": 6
  },
  "0x959f6228db998ee11e2f11c494f506a70bff6aa60d7cbdcda3176f6eaee069f1::suipepe::SUIPEPE": {
    "name": "PEPEONSUI",
    "symbol": "$SUIPEPE",
    "description": "I’m just a first $PEPE on Sui. That’s all.",
    "iconUrl": "https://images.hop.ag/ipfs/QmYbHdJKxc9PH64HkFeidzwDGpCRXECLPS1kwLYRQoXBQF",
    "decimals": 6
  },
  "0x4c0b496f07ba0ead086698069e32e31b249b5c2e671ef1067a19f945717e7b73::hopium::HOPIUM": {
    "name": "hopium",
    "symbol": "hopium",
    "description": "hopium!",
    "iconUrl": "https://images.hop.ag/ipfs/QmTqZ1Kw82amoD575LSGhVf5sXtdpYhxs7HrBDV96Ld249",
    "decimals": 6
  },
  "0x47f6e1cfd76de9161ac323e1ef0ef40dde538c2b3b42bab41b2b280d0cffc49d::hopdog::HOPDOG": {
    "name": "HOPDOG",
    "symbol": "HOPDOG",
    "description": "HOP First Dog",
    "iconUrl": "https://images.hop.ag/ipfs/QmcQeLSr4LME24GRTpS5kpt5hpwPi6vEHVYeZegkFdU232",
    "decimals": 6
  },
  "0x5f0d9413a0efe838ddac77b2134d0aef73b61446f388b89885aa47df88f7be19::pnut::PNUT": {
    "name": "Peanut the Squirrel",
    "symbol": "PNUT",
    "description": "It’s our job to help save peanut the squirrel! Sign the petition!",
    "iconUrl": "https://images.hop.ag/ipfs/QmPncTECroMSa7Jxth4R69QNWwssKFFLrQ5Nvh3rFqBiJP",
    "decimals": 6
  },
  "0x0ee2ddff21c6171cbdd8b26a547f8cfc4442797e2aa14132735df4b82f27be0d::owl::OWL": {
    "name": "HOPOWL",
    "symbol": "OWL",
    "description": "The First OWL on Hop.fun",
    "iconUrl": "https://images.hop.ag/ipfs/QmSZ7xkgG9V8bVAr1pC4uwSBvgrSMxdSNAHaw1uoN3KNKd",
    "decimals": 6
  },
  "0xac2b974b379896f8429f5911391408b5a85b2438293c8c96a60071ef82839ad9::spup::SPUP": {
    "name": "seal pup",
    "symbol": "spup",
    "description": "In the deep blue depths of the Sui ecosystem, a promising memecoin has emerged: Seal Pup. Poised for a bright journey ahead, SEAL PUP",
    "iconUrl": "https://images.hop.ag/ipfs/QmSwqazGiAVe4BfFqbmDEq4ZRABgjK41VB9LXioY8BEDVA",
    "decimals": 6
  },
  "0x17583d7125bf5825256af92564f3c28fe80d09257c18805abbe4ec08afb5f3c3::rabbit::RABBIT": {
    "name": "Rabbit",
    "symbol": "Rabbit",
    "description": "Welcome to the Rabbit World, a new era awaits you! 🐰",
    "iconUrl": "https://images.hop.ag/ipfs/QmTjuY6xdMyp96JdLEuU9MgvmUfSf34p5zhEXAbVrW1Z2S",
    "decimals": 6
  },
  "0x0fc34bb1abad3da4433e7450d7723e61e28905b801dede8abf8c28d120ff5e61::ghost::GHOST": {
    "name": "GHOSTBUSTERS",
    "symbol": "GHOST",
    "description": "Whatcha gonna buy?",
    "iconUrl": "https://images.hop.ag/ipfs/QmdYVstP8gYJ9qcT6LvuNtaXUXd1oemkKPfvxi8XNBNQm9",
    "decimals": 6
  },
  "0xba25c11b32159ecc215bd67a85ef2ae00edf1d278ad04198ffa6aae2c09bc12e::xai::XAI": {
    "name": "Xai Memes",
    "symbol": "XAI",
    "description": "Meme",
    "iconUrl": "https://images.hop.ag/ipfs/QmXVRmyotXfz5zSNmSM6Ao4qmmsry9CF5fsowYzZHxgSvP",
    "decimals": 6
  },
  "0x8b7aa115daaa980ac5476944d3971bce9731bc02380d2039f1bfb402fffa0fce::beluga::BELUGA": {
    "name": "BELUGA",
    "symbol": "$BELUGA",
    "description": "THE FIRST WHALE HOPPING IN THE SUI WATERS",
    "iconUrl": "https://images.hop.ag/ipfs/QmenaDAB4AqadRUbwnigf7oJA7A5e9bS3YnZY6gfc1XUX3",
    "decimals": 6
  },
  "0x607e010aa30291b1f6e96d53293b76c8a081488ca0d18cbdfa9bbe5a1720ff3d::suipe::SUIPE": {
    "name": "PEPE ON SUI",
    "symbol": "SUIPE",
    "description": "PEPE ON SUI",
    "iconUrl": "https://images.hop.ag/ipfs/QmRxbiTGcMUYKGZaKPZSLFSSKUNebRcQ5PEvQ1oqtkUGBF",
    "decimals": 6
  },
  "0x01de2b217167ab938ffd90692d9829273a080f035a5a27483ee12035691b85b8::hopdog::HOPDOG": {
    "name": "HOPDOG",
    "symbol": "HOPDOG",
    "description": "Doogy, The Hopper",
    "iconUrl": "https://images.hop.ag/ipfs/QmQcWxSv4KMRKz5uhZgGAuiB3iK6VKcTio9r1wzAiaZk99",
    "decimals": 6
  },
  "0x0282df8fb7bd82114c70112b4716e77f8ba7012ce448ee561f12b07fc063f55a::swoop::SWOOP": {
    "name": "Swoop 💧🐤",
    "symbol": "SWOOP",
    "description": "🐤💧 Woop Woop Woop 💧🐤",
    "iconUrl": "https://images.hop.ag/ipfs/QmYGcjMQEVqv5DnLNKKH4gYuLi457vheuadKYpgbFhCN4a",
    "decimals": 6
  },
  "0xc6df34d6e9c054900db2bd959e153d9e138fe5e8933692703ef81800d50f24ee::bubl::BUBL": {
    "name": "BUBL",
    "symbol": "BUBL",
    "description": "BUBL  BLUB\nin the world of Sui, love is simple. bubl loves blub, and together, they make everything a little more bubbly. this duo is all about bringing joy, fun and a splash of something special. because at the heart of bubl, there's always a little blub",
    "iconUrl": "https://images.hop.ag/ipfs/Qmbf752EhgUbtLCph7R9XuGYiWQZKdyCaA8Bo5afpwuNiY",
    "decimals": 6
  },
  "0xc121465afb9596e63b35c6a90a97fcd3c4d3a2545c5e656200b42dd304734a62::hopwif::HOPWIF": {
    "name": "HOP WIF",
    "symbol": "HOPWIF",
    "description": "hop wif token",
    "iconUrl": "https://images.hop.ag/ipfs/QmQvXjD3avEQMvYHXK9qGPA4ixefgNqQXLfyd5PiLz3y35",
    "decimals": 6
  },
  "0x32f8e88b2a8167a29c2ad2fb6e26c68d58b50b525111657baa981e3f00e7444f::fun::FUN": {
    "name": "HFM",
    "symbol": "fun",
    "description": "meme",
    "iconUrl": "https://images.hop.ag/ipfs/QmSyYkjBEBJc2orzpMBQyUPTTu7yNceZTYYCZU5BuQQRvN",
    "decimals": 6
  },
  "0xcd049bd0d295a26f74244e16b66927e2a76b60577beb01eb0a13668382a4f1e9::migo::MIGO": {
    "name": "Amigos",
    "symbol": "Migo",
    "description": "Welcome to Amigos, the meme coin that's so laid-back, it's practically horizontal.",
    "iconUrl": "https://images.hop.ag/ipfs/QmbX9NveFjtYhDuetVw1RAXQbQAMHpB7W5ygZ1RTTjYHed",
    "decimals": 6
  },
  "0x751b0e929c0d1dd77548dc31f2e2ccca736a00c0de3d1327f193e4d7521f3f0a::face::FACE": {
    "name": "Suiface",
    "symbol": "FACE",
    "description": "This is my face.",
    "iconUrl": "https://images.hop.ag/ipfs/QmUX2VTdjCJig67wWrH5t3MVxV521cv5w1CX9jTXt6wWqs",
    "decimals": 6
  },
  "0x1ca876cee2560a929907c09892b7c413e8eaea9ca402339631b3a4f08c355d0e::hdog::HDOG": {
    "name": "Hop Dog",
    "symbol": "HDOG",
    "description": "a",
    "iconUrl": "https://images.hop.ag/ipfs/QmT1PLJWZTfZQx8qA25EgA2oRqn4ANKDdK6pseWr1Z6RXt",
    "decimals": 6
  },
  "0x8d3b159836d9e9e83766c297429bd54864010fbaaf9118d293735cb25dae3206::squi::SQUI": {
    "name": "SQUI",
    "symbol": "SQUI",
    "description": "SQUI is the cutest mollusc to ever grace the depths of the Sui ocean.",
    "iconUrl": "https://images.hop.ag/ipfs/QmQzzAfNYtgLQbGEaGruS5dQnAFqGmioht1WEFVfz3rQgW",
    "decimals": 6
  },
  "0xd709e7e941036e2c641da126d1a9097cf065bafdf23d3ea9f162121bba04ccb9::dodo::DODO": {
    "name": "dodo",
    "symbol": "dodo",
    "description": "$dodo is your rabbit friend.",
    "iconUrl": "https://images.hop.ag/ipfs/QmR9jrTj3kUoXBRQj8i4UxxtN72hRcuWiGXRZxWUg9Tife",
    "decimals": 6
  },
  "0xba2ad31440c90386d3fc18d7449e530bbeefcdef335039d23be14b324873dc6f::bubl::BUBL": {
    "name": "BUBL",
    "symbol": "BUBL",
    "description": "BUBL\nin the world of Sui, love is simple. bubl loves blub, and together, they make everything a little more bubbly. this duo is all about bringing joy, fun and a splash of something special. because at the heart of bubl, there's always a little blub",
    "iconUrl": "https://images.hop.ag/ipfs/QmbFzA2pwoSSjs1hr9a5z8p44eb2yAUF47TFpARHUu1orZ",
    "decimals": 6
  },
  "0xb28dd39e55f54e81a817dd704b29eb19117f13a6084821a8073663bd894fc775::ghost::GHOST": {
    "name": "Ghostbusters",
    "symbol": "GHOST",
    "description": "Whatcha gonna buy?",
    "iconUrl": "https://images.hop.ag/ipfs/QmdYVstP8gYJ9qcT6LvuNtaXUXd1oemkKPfvxi8XNBNQm9",
    "decimals": 6
  },
  "0xc8380a4398e7fbba8a583768a03a8326ab89b1ede0c77064ad7faf5eab4ab15d::squide::SQUIDE": {
    "name": "Squide",
    "symbol": "Squide",
    "description": "Squide The Pearl OF SUI, The cryptocurrency market on the SUI network.",
    "iconUrl": "https://images.hop.ag/ipfs/QmSdzf5XBe6TV93XX9RX19KJqiBii5VSiNozwVqbMwzzip",
    "decimals": 6
  },
  "0xb2fcc6b75c35a75ad1ffdce3aa54e9109dbc4539d027580338cd5da51d8427ef::hopcat::HOPCAT": {
    "name": "HopCat",
    "symbol": "HOPCAT",
    "description": "first ever deployed cat on hopfun",
    "iconUrl": "https://images.hop.ag/ipfs/QmaGmGXfXxwfUnn9B3HjJPdzdXPft8LhmQX18PQkk4grRb",
    "decimals": 6
  },
  "0x48026bf948bc1b07036a5dbf65dfa091e35382734593ce21a9f854ed089cd40a::pnut::PNUT": {
    "name": "Peanut the Squirrel",
    "symbol": "Pnut",
    "description": "Pnut",
    "iconUrl": "https://images.hop.ag/ipfs/QmPWFeAXhN9rnQ2ZrpdNjmVqjJugFmu9sRbmDeom65pM9X",
    "decimals": 6
  },
  "0x7c557acc4c7de66c5f470d38d19cbda1f4244dc2b874bc8f2edbe60fbfcd52f5::suilana::SUILANA": {
    "name": "Sui + Solana",
    "symbol": "SuiLANA",
    "description": "Sui is becoming Solana",
    "iconUrl": "https://images.hop.ag/ipfs/QmWQA2U7dG7u5tz1AGabd9ZZebM6bpBvSdvpJ3yMrKtung",
    "decimals": 6
  },
  "0xa24794a04e68476bc8d56926423bbc5fd06059cbfc81ea20650cfcbbc2a420a6::hopcat::HOPCAT": {
    "name": "Hopcat",
    "symbol": "Hopcat",
    "description": "Hopcat",
    "iconUrl": "https://images.hop.ag/ipfs/QmRa3XL5YZQhDzQQyyEzCH1Q9aGaDEGe6N1TXkwyeNDRtP",
    "decimals": 6
  },
  "0x60cdfe2606bbf7f4750e241d62bf419f0a104ac0e0d990b572d427d58c4bd4ad::migo::MIGO": {
    "name": "Amigos",
    "symbol": "Migo",
    "description": "Welcome to Amigos, the meme coin that's so laid-back, it's practically horizontal.",
    "iconUrl": "https://images.hop.ag/ipfs/QmbX9NveFjtYhDuetVw1RAXQbQAMHpB7W5ygZ1RTTjYHed",
    "decimals": 6
  },
  "0x7e8bb443f34987fb5c223cb251f4da6163d0c2ec5e1ceeaa8dbbee55165d4aea::sui_rock::SUI_ROCK": {
    "name": "SUI ROCK",
    "symbol": "SUI ROCK",
    "description": "Ape Rocks - first rocks on SUI",
    "iconUrl": "https://images.hop.ag/ipfs/Qme8AVsruhVUK64ndGjkvQV8kETbDz6csMpxdCcVU3rqZX",
    "decimals": 6
  },
  "0x632b7b1a865fa5257b77bbb69cf624ef49a2d5a4e2df0956f4475d211aa5d0d1::hpos::HPOS": {
    "name": "Hop piece of shit",
    "symbol": "HPOS",
    "description": "Hop piece of shit",
    "iconUrl": "https://images.hop.ag/ipfs/QmZbv9UGvXyY1XSL628odoQyjvJ5y7bfmwRpBCxA9T8JJD",
    "decimals": 6
  },
  "0x8808a2959c2ba0204a7b55f8eeee5a9a3235ea1869e0250caa6a79a1189ea1cc::fuckhop::FUCKHOP": {
    "name": "FUCK HOPFUN",
    "symbol": "FUCKHOP",
    "description": "FUCK HOPFUN TOGETHER",
    "iconUrl": "https://images.hop.ag/ipfs/QmPTYzanEi2Y2jB65g44E57BqaApb21JYB4BQpuUx5HRb4",
    "decimals": 6
  },
  "0xe348db5b3a57f6696077887308d774652ede5c997a20b4a6efb882f39a6e0470::rebyata::REBYATA": {
    "name": "REBYATA",
    "symbol": "REBYATA",
    "description": "Rebyata is the crypto token that looks harmless and friendly, just like Leopold the Cat—but don't be fooled! With spiked fists and a smile, it's here to shake things up in the market. Join the wild side of crypto with Leopold Coin and make gains that won't just scratch the surface!",
    "iconUrl": "https://images.hop.ag/ipfs/QmSTAter1mysVmiA9S5hkL7S7EnXqXZLVhM7C8zCL4zWkK",
    "decimals": 6
  },
  "0xa2b587f73b9425ba1edd034e00a4b6c044b4f836f6ffff9ad5cff9061befb26b::rex::REX": {
    "name": "suirex",
    "symbol": "REX",
    "description": "I'm Rex. W're all Rex",
    "iconUrl": "https://images.hop.ag/ipfs/QmUzqQG6zVXzQ3moLjrFhb3MnrAG2ufmKS6emzd7L2xNqn",
    "decimals": 6
  },
  "0x54add2d0cf25d62c87b9c4487a385d832a5b4f84c5f879f7b84cd0a072701c6e::ghost::GHOST": {
    "name": "Ghostbusters",
    "symbol": "GHOST",
    "description": "Whatcha gonna buy?",
    "iconUrl": "https://images.hop.ag/ipfs/QmdYVstP8gYJ9qcT6LvuNtaXUXd1oemkKPfvxi8XNBNQm9",
    "decimals": 6
  },
  "0x3b7deccf9a10af8dd2445a93be9786ccde9f219a58110c7b21c4b1f38962dafa::frog::FROG": {
    "name": "HopFrog",
    "symbol": "$FROG",
    "description": "From the ashes, a new community rose: not just a Frog, but a more grounded, united Shinobi Dojo.",
    "iconUrl": "https://images.hop.ag/ipfs/QmUxugtFwfRN2WVLSW9BTrKqM9c6udo7bM7aU81NkhTcUr",
    "decimals": 6
  },
  "0x50cd383d11a0f68ff337ef7515e80b04deca653baa3f331f1daca1234fe14e02::suilander::SUILANDER": {
    "name": "SuiLander",
    "symbol": "SUILANDER",
    "description": "Lander family is arriving to SUI network...meet 'em all!",
    "iconUrl": "https://images.hop.ag/ipfs/QmRLjAX2Fdue4K95hT6EBpBwMpZHZx5qP77upYS9K9W6Ci",
    "decimals": 6
  },
  "0x72fdf062285dba36e3fd398098d6a1dc10d3ac08ae6eeddbc2aff71417bf36e0::hopcat::HOPCAT": {
    "name": "HopCat",
    "symbol": "HOPCAT",
    "description": "first ever deployed cat on hopfun",
    "iconUrl": "https://images.hop.ag/ipfs/QmaGmGXfXxwfUnn9B3HjJPdzdXPft8LhmQX18PQkk4grRb",
    "decimals": 6
  },
  "0x9a00d43703b50a88728cb3cfc43f50425d925d53c2aef92f5a581bfe30103d36::red::RED": {
    "name": "REDACTOOR",
    "symbol": "RED",
    "description": "Redactoor: Turning ‘redacted’ hype into meme-fueled mayhem!",
    "iconUrl": "https://images.hop.ag/ipfs/QmSzRJngYETPu6CcWCDgMTNJ8ryUiLzgoh6hSei7rjHDWK",
    "decimals": 6
  },
  "0xdd65c7896270f98f58c0857ed248b747712144986b6093427adfbc636a6a26eb::hdog::HDOG": {
    "name": "Hop dog",
    "symbol": "Hdog",
    "description": "First dog on hop",
    "iconUrl": "https://images.hop.ag/ipfs/QmVNLk4MXko3CAvq5nzH6BWteAcG45vQrY5zyrkDqHWPxs",
    "decimals": 6
  },
  "0x59a3502c812ba9fb0489949ee9ec75112afdf932ed0a38f41567e297439c4af9::first::FIRST": {
    "name": "The First",
    "symbol": "FIRST",
    "description": "The first of its kind. The creator of all. The first SUI coin made.",
    "iconUrl": "https://images.hop.ag/ipfs/QmUNzFRPW4z3yzPS1A2KX6QdEZz27GN4aH3cQZoGSvxgAA",
    "decimals": 6
  },
  "0xbe4e75216625482efa98f67747ab5ccf60997aeca7273224967b7526dcd88a93::paw::PAW": {
    "name": "PawPaw",
    "symbol": "PAW",
    "description": "hop hop, grrr grr, woof woof",
    "iconUrl": "https://images.hop.ag/ipfs/QmVcgt1NRcAetTi3jHvrKscfaNQ5K5jJ83CzShbFk5qtRL",
    "decimals": 6
  },
  "0xb3147a972e6c5a32822be8b816b13b0a153c4d3c3e9912d80136b21cb5e7881a::ghost::GHOST": {
    "name": "Ghostbusters",
    "symbol": "GHOST",
    "description": "Whatcha gonna buy?",
    "iconUrl": "https://images.hop.ag/ipfs/QmdYVstP8gYJ9qcT6LvuNtaXUXd1oemkKPfvxi8XNBNQm9",
    "decimals": 6
  },
  "0x6860961d69cca084fb334c7aaf7b83b796810d96e7bf7612a414b2b257e01829::qui::QUI": {
    "name": "DeSQuill",
    "symbol": "Qui",
    "description": "Patriot",
    "iconUrl": "https://images.hop.ag/ipfs/QmTkVX5LCn7iCvHJTTfMQr6kb2JCJ53xGfgmChbUBdg3gG",
    "decimals": 6
  },
  "0x4a1f4ff537d159b7f5c9245a62948976c8b24769cf0665ab3a8043b03c67939d::pnut::PNUT": {
    "name": "Pnut",
    "symbol": "Pnut",
    "description": "Pnut",
    "iconUrl": "https://images.hop.ag/ipfs/QmbsG6iwzy3ykkrRhaqd1NKFao1QK33qTG9Jt6rf1ov8sR",
    "decimals": 6
  },
  "0x76d23d1b2d8f4f6dc3b32accae541810ab508d891ad2a30a2cd7ac157fd97bb7::suiyan::SUIYAN": {
    "name": "SUPER SUIYAN",
    "symbol": "SUIYAN",
    "description": "It's the Super Suiyan cycle",
    "iconUrl": "https://images.hop.ag/ipfs/QmbnM9L7FdGNLYgNpA6kmJw3sSf4QFRm36QEaqwmzk5Zn8",
    "decimals": 6
  },
  "0xbb5e262d8174fd15d31bc7aa98e2aafee11d11fedabf97973325893e4f0d06c9::hopdog::HOPDOG": {
    "name": "hopdog",
    "symbol": "hopdog",
    "description": "hopdog test",
    "iconUrl": "https://images.hop.ag/ipfs/QmRKatJ6xRmwJ3NaK24UT4WwU7YqCXTWQbTf4AZGnHPn8R",
    "decimals": 6
  },
  "0xe938675eca6a01b5a3f20ec9c5d537a3c8986438b38cc8cbc3e9ea1c9db320a9::trump::TRUMP": {
    "name": "TRUMP",
    "symbol": "TRUMP",
    "description": "RULE THE WORLD",
    "iconUrl": "https://images.hop.ag/ipfs/Qmem6Q1rCNjf32qUWmRhENBP5e86Y3prNawHSubAiUJDXj",
    "decimals": 6
  },
  "0x93024aeaabac6f4144f0eb6956768decfa0619d8459057db7de2b0369f4698d0::ocean::OCEAN": {
    "name": "DEEP OCEAN",
    "symbol": "OCEAN",
    "description": "DEEP OCEAN is the official color of the SUI blockchain logo.",
    "iconUrl": "https://images.hop.ag/ipfs/QmR3XipksKVPboVVBCsGp29BVm3ieXZCNaDjhMXCiMEUTA",
    "decimals": 6
  },
  "0x2b043acefa5f2228a7cdcbe6a3f3e0a1738c31e85ee018dc9917626875eecd30::hoppy::HOPPY": {
    "name": "Hoppy",
    "symbol": "Hoppy",
    "description": "The 1st Kangaroo on Hop Fun",
    "iconUrl": "https://images.hop.ag/ipfs/QmRA9w7AGKzny9WTuMnXEWzZGzPv3yxjqNEVRgBpdGKtyF",
    "decimals": 6
  },
  "0x764ed8bbf6bc48562964aae62d951c0de71488f9897d31cefa43232019494f01::hopless::HOPLESS": {
    "name": "HOPLESS",
    "symbol": "HOPLESS",
    "description": "this is community token!\nif you believe HOP than LESS you're not our part",
    "iconUrl": "https://images.hop.ag/ipfs/QmY1Zr9Wbvevjs1KC8q31PsNKUeQL3ayc5Nx63ZUZ5GFid",
    "decimals": 6
  },
  "0x0228d4f0062b8815e3eaa0c9227ab832c4bcf1ff19332c5c354670d5138a2138::rtm::RTM": {
    "name": "RAMBA THE MAMBA",
    "symbol": "RTM",
    "description": "Ramba has been stolen by the state of New York and it’s our job to spread the message and get him released! Do your part. Once free we’ll continue to support and celebrate Ramba!",
    "iconUrl": "https://images.hop.ag/ipfs/QmQxL6ZDNgaBLoAej8qHwi6yqtTazmXmf8J7JEYj7Pp5SP",
    "decimals": 6
  },
  "0x7e1affb0f1b07585e41f1580ad78c81c71553495a12dfbec0b96db3218d2e9bd::nsui::NSUI": {
    "name": "First Nigga To Launch",
    "symbol": "nSUI",
    "description": "FIRST NIGGA TO LAUNCH ON HOP",
    "iconUrl": "https://images.hop.ag/ipfs/QmWCKj4rcA3wktrD2dJ6edJBmGhvBSAxhjDsqvDbZ7FefK",
    "decimals": 6
  },
  "0xeb8a76fac7b4df641d8d6a3670396cc9fcc8aec02af5507c0cb73fdba9ae0755::hehe::HEHE": {
    "name": "HEHE",
    "symbol": "HEHE",
    "description": "HEHE",
    "iconUrl": "https://images.hop.ag/ipfs/QmR1Rhbs6rBpoyHsnLvmdia7JfYF4TvdfT2uL3SWJxSZWb",
    "decimals": 6
  },
  "0xc554d030985ac0306b93807343918bbd9324342505970c683ce1844c195d4890::hoppo::HOPPO": {
    "name": "Hoppopotamus",
    "symbol": "Hoppo",
    "description": "First Hippopotamus on SUI! Hoppopotamus",
    "iconUrl": "https://images.hop.ag/ipfs/QmbyWg4TzUd5SXjEXCtiuednDfKjNRSgaR9kDP4L9ZQiKs",
    "decimals": 6
  },
  "0x9058bc23103ea6acbaab9a76b023a07361896d78610dade9068bc8a0c7089ef8::shark::SHARK": {
    "name": "SHARK",
    "symbol": "SHARK",
    "description": "$SHARK surges from the $SUI ocean and unleashing a fresh wave",
    "iconUrl": "https://images.hop.ag/ipfs/Qmek5ZyUAektrzmq5VynwRFirMBQUgU1StKhuDLacGgk1b",
    "decimals": 6
  },
  "0x1bd1a24892906bda8af25997246c265ab22e345f4d5a9bbb051fd9496db1f4ee::hopdog::HOPDOG": {
    "name": "HopDog",
    "symbol": "HOPDOG",
    "description": "HopDog was a cheerful dog and the mascot of the SUI network. He helped developers and became famous for his playful spirit.\n\nUsers created \"HopCoin\" in his honor, which was generated with each transaction to reward good deeds. HopDog became a symbol of friendship in the community, showing that fun and collaboration are the greatest treasures of the digital world.",
    "iconUrl": "https://images.hop.ag/ipfs/QmdnryzzbKJTEYVfeaEoeXM8dBrqciyP6ec5zKKGLQVMkT",
    "decimals": 6
  },
  "0xad060f63bbfbc3db561d6c3fc10d6d97cfa0ef3984bef1aca881cdd1fbe6dfb9::hopcat::HOPCAT": {
    "name": "HOPCAT",
    "symbol": "HOPCAT",
    "description": "We are CAT !",
    "iconUrl": "https://images.hop.ag/ipfs/QmWBx34wFTPcn8u2RbT74MgrKZFcbssptPYZWZhHQM3oVt",
    "decimals": 6
  },
  "0xbde214cfe38436b3b6ddab43a34467284a63828e486ebf1f7ba1dae1f0c3ab8c::hops::HOPS": {
    "name": "HopSui",
    "symbol": "HOPS",
    "description": "In a world where blockchains are racing toward the future, HopSui is the cheerful guide, born in the heart of Sui’s vibrant digital realm. With floppy ears and a radar for the hottest trends, HopSui bounds across decentralized networks, bridging gaps between communities, spreading opportunities, and energizing the ecosystem with its relentless drive. HopSui was created to empower every user, from seasoned investors to curious newcomers, to move quickly and confidently in the crypto space.\n\nHopSui is not just here to run laps around the market but to inspire others to join the journey with a token that’s as transparent and community-focused as the network it calls home. With each hop, HopSui is setting new standards of connectivity and speed, offering unique staking rewards, fun community events, and a roadmap packed with innovative features.",
    "iconUrl": "https://images.hop.ag/ipfs/Qma1AMoKVKmhqhGoWA5zqrYxojKvQUtmkE2q3YxnP1cM1k",
    "decimals": 6
  },
  "0x6a201a57e85c335ebacd8333d47e1ef824316dc1cb3193fe912a5860d9f9c727::billy::BILLY": {
    "name": "BILLY",
    "symbol": "BILLY",
    "description": "The community is bullish about Billy (BILLY) today.",
    "iconUrl": "https://images.hop.ag/ipfs/QmTf9tGmZXAuDG5BbYNG2iDomHudQqDVdArTwhsXS4cVqU",
    "decimals": 6
  },
  "0x220d50c0397bc50533557942c89dd172b99fd18bcd89e8bda85db911be900863::hopcat::HOPCAT": {
    "name": "Hopcat",
    "symbol": "Hopcat ",
    "description": "First",
    "iconUrl": "https://images.hop.ag/ipfs/QmRa3XL5YZQhDzQQyyEzCH1Q9aGaDEGe6N1TXkwyeNDRtP",
    "decimals": 6
  },
  "0xdd5178392d7747041bf0c81acc1cc87058212dc4f3848db82ec1fe074f04989f::ronado::RONADO": {
    "name": "CRISTIANO RONADO",
    "symbol": "RONADO",
    "description": "I AM RONADO\nI WIN EUROS\nI AM PORTUGA\nSUUIIIII",
    "iconUrl": "https://images.hop.ag/ipfs/QmUoLjvYgpAtStgNSJW1pyR8pwFboTAjrGqeGoGvCWdbai",
    "decimals": 6
  },
  "0x928d76f7c44b451762a5d297ea52acfc0269c12af68d0f6ca302c18b79196456::dodo::DODO": {
    "name": "dodo",
    "symbol": "dodo",
    "description": "$dodo is your rabbit friend.",
    "iconUrl": "https://images.hop.ag/ipfs/QmR9jrTj3kUoXBRQj8i4UxxtN72hRcuWiGXRZxWUg9Tife",
    "decimals": 6
  },
  "0xc89f362e52384c3f3d0d2e888126f8a6b9a1f7a2c93ead4b5c8d3b2651aabd3e::uni::UNI": {
    "name": "UNI DOG",
    "symbol": "UNI",
    "description": "UNI",
    "iconUrl": "https://images.hop.ag/ipfs/QmaDNCGSuLUqitJNugBBCMLhdmPBNFZCdhU2eSYJzixpYK",
    "decimals": 6
  },
  "0x65491b4efcf6161237dabffabc7fedfced2ecdb9975d136b5628e859118cb8c0::pnut::PNUT": {
    "name": "Pnut",
    "symbol": "Pnut",
    "description": "punt",
    "iconUrl": "https://images.hop.ag/ipfs/QmbZT8enZzdJURSpWyxf6MipPM5UKRXdx93HbDQrfw1m9x",
    "decimals": 6
  },
  "0xf375faa0611e0ca4dcbe9926f85b81f90cef3e6f54d8fd7c9b10888408723a17::bluuuub::BLUUUUB": {
    "name": "BLUUUB",
    "symbol": "BLUUUUB",
    "description": "BLUUUUUB",
    "iconUrl": "https://images.hop.ag/ipfs/QmcY3e7vh8uMbZrjHG9whKw22vHuAntgSFDMkUaJhm4ptU",
    "decimals": 6
  },
  "0x6753612ea8530661e8b5992f218da04320d05a112be4be65c2f19ba23f588fef::bunny::BUNNY": {
    "name": "HopBunnyonsui",
    "symbol": "BUNNY",
    "description": "HOP BUNNY THE FIRST BUNNY ON hopaggregator & suinetwork",
    "iconUrl": "https://images.hop.ag/ipfs/QmbkLAsbAgonB67EgvfWjLD4RPzyW8WZM9Jjn5qArwjQe5",
    "decimals": 6
  },
  "0xdd47fe74fbc025597dd12eda75104782e9754e5b160c9065b1dc2265087c12d6::hopcat::HOPCAT": {
    "name": "HopCat",
    "symbol": "HopCat",
    "description": "HopCat",
    "iconUrl": "https://images.hop.ag/ipfs/QmRa3XL5YZQhDzQQyyEzCH1Q9aGaDEGe6N1TXkwyeNDRtP",
    "decimals": 6
  },
  "0x3c4468980e2f97fcec59d05fe7ff0ee60a5d5a03475ce20c6f74c831a5e43dda::pepepeppe::PEPEPEPPE": {
    "name": "pepe",
    "symbol": "pepepeppe",
    "description": "no.1",
    "iconUrl": "https://images.hop.ag/ipfs/QmcGJRRfktkf7tEMPbbXLYJgtLtbBvqVbnwKcCXhDBv2rb",
    "decimals": 6
  },
  "0xd3f1ad0947540c3c473aa60d3224aa0556ef2c9574e34b0d1c195fbf59160aa1::hopium::HOPIUM": {
    "name": "HOPIUM",
    "symbol": "HOPIUM",
    "description": "Once Upon a Meme-In a world where serious crypto projects ruled, a playful memecoin called $HOPIUM burst onto the scene.",
    "iconUrl": "https://images.hop.ag/ipfs/QmREEmntmUqpDp8goDBjxXK5cvRiEvvnAYinj426Kes3ro",
    "decimals": 6
  },
  "0xa950cc9fcf3e3f25f92baccb215d9f30a6a8f54b0918b4997f5eaebf8003fc7b::birds::BIRDS": {
    "name": "BIRDS",
    "symbol": "BIRDS",
    "description": "birds on sui",
    "iconUrl": "https://images.hop.ag/ipfs/QmdxvG2gowhJZZihWpPam2RHXFJE1xzyxhaDfSLRB5iH72",
    "decimals": 6
  },
  "0xb17d58e3fd2d0bd38a65e8835bb40d7f391bfb336313dac4e751fcd5d1cbc711::cb::CB": {
    "name": "t-CallBoyZ",
    "symbol": "CB",
    "description": "Governance token for the CallBoyZ sub-ecosystem on SUI chain.",
    "iconUrl": "https://images.hop.ag/ipfs/QmZRMjD582M84ZRZBMkmAmzXQV5vLTbLqbt7CL27nuTXYN",
    "decimals": 6
  },
  "0xa23874c6510cf3978068e8e71b83825c76d99d7f70ed225a5f23c7b459f0e5c6::qui::QUI": {
    "name": "DeSQuill",
    "symbol": "Qui",
    "description": "Patriot",
    "iconUrl": "https://images.hop.ag/ipfs/QmTkVX5LCn7iCvHJTTfMQr6kb2JCJ53xGfgmChbUBdg3gG",
    "decimals": 6
  },
  "0x0256fb7c8e9ea6a1ac2f7c47afb7dd27ec4921644cffec9228443e55b5390e6f::bvr::BVR": {
    "name": "Beaver",
    "symbol": "BVR",
    "description": "beaver is not just a meme its movement",
    "iconUrl": "https://images.hop.ag/ipfs/QmdASx1tdgFPMUaBim37JC9nWTmXoJg8s6Bw21px1uVDLF",
    "decimals": 6
  },
  "0xd51b25c84268b5e5ed2408ec3edef40dc88068cb346feff00d4a88574aa4635f::pepe::PEPE": {
    "name": "SUI PEPE",
    "symbol": "pepe",
    "description": "$PEPE. The most memeable memecoin in existence.",
    "iconUrl": "https://images.hop.ag/ipfs/QmWoETfzdWpdMb9MNrttKnUx4oFPBViqYdyvXidTrQAHXZ",
    "decimals": 6
  },
  "0x2282aa660202720edb18ead7df1377be5e39fc34cb49d11e7a415ff85f1902d8::rex::REX": {
    "name": "SuiRex",
    "symbol": "REX",
    "description": "Sui and Rex",
    "iconUrl": "https://images.hop.ag/ipfs/QmSdMzkyWpF9ANKt6iLFa1dYEozZfKRHcvPKHZoL3o8z5V",
    "decimals": 6
  },
  "0xc3308ee10abddeaedd049fa10a988ab617fc0ee2619f791b1df7945bfe20714a::hop::HOP": {
    "name": "HopFun",
    "symbol": "HOP",
    "description": "For the best price with zero fees.",
    "iconUrl": "https://images.hop.ag/ipfs/QmQvXjD3avEQMvYHXK9qGPA4ixefgNqQXLfyd5PiLz3y35",
    "decimals": 6
  },
  "0xd73b611673fc86567733ef272e7370fdf691091cfdc6fd1ea4bd5caca929b0d5::hopcat::HOPCAT": {
    "name": "HOPCAT",
    "symbol": "HOPCAT",
    "description": "HOPCAT",
    "iconUrl": "https://images.hop.ag/ipfs/QmQRV3UxnBtx1SPVBAFTGgMLbuTmivEBo7B8yKiwytCJbR",
    "decimals": 6
  },
  "0x98d2e64e2cc22b651fb64cd747b00229a4cb62c7c097492496e5e052f6f67e3f::bk::BK": {
    "name": "bonkman",
    "symbol": "bk",
    "description": "aa",
    "iconUrl": "https://images.hop.ag/ipfs/QmWn6rXpcFnNjaZzGBaPKDz6yYqnW2H6AE82DHnUPDP4J6",
    "decimals": 6
  },
  "0xf68f206fbdb3db58efcff3637b63bd362115e5cacbc1cfb806f674ec7dcc2af2::fred::FRED": {
    "name": "FRED",
    "symbol": "FRED",
    "description": "FRED",
    "iconUrl": "https://images.hop.ag/ipfs/QmZ9NUcGrhdW4pNz8bMeUpYNaecPQGjCuDHc7Wxs4uPpMP",
    "decimals": 6
  },
  "0x3cff6237ec738e85ec1781cc087090cbc3cc6ec166ac987bb737fceae14c66c7::hopium::HOPIUM": {
    "name": "Hopium",
    "symbol": "HOPIUM",
    "description": "sometimes you just hop",
    "iconUrl": "https://images.hop.ag/ipfs/QmaTPnY8SXLci5WsaJ41cDxsLrCWy7sjTiPUVu1prs3GP7",
    "decimals": 6
  },
  "0x422e45ada00fc49369187e68f238c6b490207a838deb2038cd341edb0184b231::peanut::PEANUT": {
    "name": "Peanut",
    "symbol": "Peanut ",
    "description": "It’s our job to help save peanut the squirrel! Sign the petition!",
    "iconUrl": "https://images.hop.ag/ipfs/QmPncTECroMSa7Jxth4R69QNWwssKFFLrQ5Nvh3rFqBiJP",
    "decimals": 6
  },
  "0xa8debf9f3ea3d26903f7f3351d0b92e2958ef75948e712e3ad30adc7fd5474e3::kenobi::KENOBI": {
    "name": "Kenobi",
    "symbol": "Kenobi",
    "description": "SUI First Kenobi",
    "iconUrl": "https://images.hop.ag/ipfs/QmP1pNBdQPGzrbyN4wUuqzSWurVGJgxhkw1KpTnsMvwGUb",
    "decimals": 6
  },
  "0xd9ce01c774e8d2cc32c37f4e032caec60fffa126ee9199030b4df2e5407305f7::srl::SRL": {
    "name": "Suirrel",
    "symbol": "SRL",
    "description": "Suirrel SUI MASCOT",
    "iconUrl": "https://images.hop.ag/ipfs/QmWNT3jGjZFRsz7ipsNyX8gwtptujcwaxgqde2yG5RsWCo",
    "decimals": 6
  },
  "0xd96e6388efed3936d7eb87e56d7e0f7d79a079f4a7b8617871850d3f3c3e9a02::hopfomo::HOPFOMO": {
    "name": "FOMOONHOPFUN",
    "symbol": "HOPFOMO",
    "description": "This is your Frist FOMO on hopfun but not the last one",
    "iconUrl": "https://images.hop.ag/ipfs/Qmatv3A9BeHuB3CZCD7saRRfeJASPfmsmhPWwuQ4KVEnQg",
    "decimals": 6
  },
  "0xc36c43fdc76a8fd5432d8437d5a61f6a800e79b26b915278b43afb395651521f::jelly::JELLY": {
    "name": "Jellyfish",
    "symbol": "JELLY",
    "description": "Jellyfish is a memecoin launched on SUI Network. The project aims to capitalize on the popularity of meme tokens such as $BLUB and $HIPPO to become a leading meme cryptocurrency itself.",
    "iconUrl": "https://images.hop.ag/ipfs/QmdR2uB1NC5gFjMMFZXVCJbnmejCXrNWCtsu4w9HnhZJxH",
    "decimals": 6
  },
  "0x58627316a410ef3b78a7f2a92952132e39113ece40b619768daa6b708c75be0e::pnut::PNUT": {
    "name": "Peanut",
    "symbol": "Pnut",
    "description": "Peanut Lives On In Our Hearts",
    "iconUrl": "https://images.hop.ag/ipfs/QmbnhjUbZ2xop1AaCY1nPcG5YAcqJy8mMFSnj9Bbu5Ewav",
    "decimals": 6
  },
  "0x636cc3fbe001601798c0d3db0c5e4993b628a2a4923e7c81970dcdae970e1456::bk::BK": {
    "name": "bonkman",
    "symbol": "bk",
    "description": "https://twitter.com/bonkman22",
    "iconUrl": "https://images.hop.ag/ipfs/QmWn6rXpcFnNjaZzGBaPKDz6yYqnW2H6AE82DHnUPDP4J6",
    "decimals": 6
  },
  "0x4020d15dba843ad8dd7d431f1537ae2baa3f1ae7cb97a3e200a2d4610eb7b2e8::cat::CAT": {
    "name": "HOP CAT",
    "symbol": "CAT",
    "description": "First  CAT on http://hop.fun",
    "iconUrl": "https://images.hop.ag/ipfs/Qmen8LhMSiPa1Sy4VX3UjzvDY3XuRGaR2vzsQFWAHYKoyF",
    "decimals": 6
  },
  "0x686a9b2610ea994c3d0db8e6d0158d079e78a2741484e2b5b7fe4c6f56713fa6::suiamese::SUIAMESE": {
    "name": "SUIAMESE",
    "symbol": "SUIAMESE",
    "description": "Meet the king cat, The SUI-perstar of SUI Chain. SUIAMESE!",
    "iconUrl": "https://images.hop.ag/ipfs/QmWH5eVPRXrDRZGESQugsshjC5FrFqAt67Kui6eVcL9G8h",
    "decimals": 6
  },
  "0x174f56009be0098866b7138e8ad610ac43342c2f676ab327d80d51ab4ddd135f::pnut::PNUT": {
    "name": "Obi PNut Kenobi",
    "symbol": "Pnut",
    "description": "“If you strike me down, I will become more powerful than you could possibly imagine” Obi PNut Kenobi",
    "iconUrl": "https://images.hop.ag/ipfs/QmbZT8enZzdJURSpWyxf6MipPM5UKRXdx93HbDQrfw1m9x",
    "decimals": 6
  },
  "0xcb0443cff8f6084c2af518994253e9c4cfc09491d15366461ddc607f20caa381::ocean::OCEAN": {
    "name": "DEEP OCEAN",
    "symbol": "OCEAN",
    "description": "DEEP OCEAN is the official color of the SUI blockchain logo.",
    "iconUrl": "https://images.hop.ag/ipfs/QmR3XipksKVPboVVBCsGp29BVm3ieXZCNaDjhMXCiMEUTA",
    "decimals": 6
  },
  "0xd9e58e83a8733c302853918b6cd875bb32f7c3ca92c8ea5f3ef6a13622e21686::hopmoon::HOPMOON": {
    "name": "HOP to the MOON",
    "symbol": "HOPMOON",
    "description": "Hop fun to the moon",
    "iconUrl": "https://images.hop.ag/ipfs/QmfVVYWYUPjqJEZS5qv9GKi1bhpxs54Cc8yCMHwdNz6jHe",
    "decimals": 6
  },
  "0xe93a37714cff07c6fc1a4d19b1124b55ab17259c6fd375c06a18569144a7942f::ius::IUS": {
    "name": "IUS",
    "symbol": "IUS",
    "description": "IUS",
    "iconUrl": "https://images.hop.ag/ipfs/QmWMp5GNdW8h9RJgkrDSMbSDuQYPMop335DKr8YPCuSuP5",
    "decimals": 6
  },
  "0x61987c04d8db9a8005f6605de52179e324fc133e2733f44984f200fc9eb88e03::hopfun::HOPFUN": {
    "name": "Hopfun",
    "symbol": "Hopfun",
    "description": "HOPFUN",
    "iconUrl": "https://images.hop.ag/ipfs/QmRsoyPX9rPxfS6WjVxBxygpJTywFMYhpt3uZ9b5wUKciZ",
    "decimals": 6
  },
  "0xa47519e90e68a36d78f175437915473826d38a2cc2f8731fc0e57c7d95aea4db::bubl::BUBL": {
    "name": "BUBL",
    "symbol": "BUBL",
    "description": "bublsui.com",
    "iconUrl": "https://images.hop.ag/ipfs/Qmbf752EhgUbtLCph7R9XuGYiWQZKdyCaA8Bo5afpwuNiY",
    "decimals": 6
  },
  "0xfd1bb9943bd6b666e1307257b1e1419f8908c0bc137a4c309e5166eaa908b0dd::hopcat::HOPCAT": {
    "name": "HopCat",
    "symbol": "HopCat",
    "description": "First cat on Hop Fun",
    "iconUrl": "https://images.hop.ag/ipfs/QmWBx34wFTPcn8u2RbT74MgrKZFcbssptPYZWZhHQM3oVt",
    "decimals": 6
  },
  "0xde0b6447ab3b091ce4639393fef174ae31c7fbfffd5b49fb0a99f80494264913::doghop::DOGHOP": {
    "name": "HappyDOG",
    "symbol": "DOGHOP",
    "description": "HOP DOG 1",
    "iconUrl": "https://images.hop.ag/ipfs/QmdPryMZ2dAt6AqskqWJZ2oNvGpwDemp26ssT4Uw7sJZ3P",
    "decimals": 6
  },
  "0x4e36ba46db5767172126bd29618744f974ac7f7554e50456904077f2c8f2a832::hopcat::HOPCAT": {
    "name": "Hop Cat",
    "symbol": "HOPCAT",
    "description": "HOPCAT ON SUI",
    "iconUrl": "https://images.hop.ag/ipfs/QmbyoV2waeruJkBGp4niEkgNnD2WGgDqjpgGgWK8dLU2RH",
    "decimals": 6
  },
  "0x7c73dccb881dec4ade395af196c4aab432d353d5d1b73accab38cdcdbdd0f11c::hops::HOPS": {
    "name": "HopSoul",
    "symbol": "HopS",
    "description": "Swap on @SuiNetwork and Pray Hard for HopSoul",
    "iconUrl": "https://images.hop.ag/ipfs/QmTDKRGW9SF1yf5ShGVdV6Hr3AbraDP2t7oxPTEgAYgwSj",
    "decimals": 6
  },
  "0xcc2fa1e12b68480419bb23b7968ff777ff8454e12265a2359dabee78983d50b3::hopcat::HOPCAT": {
    "name": "hopcat",
    "symbol": "HOPCAT",
    "description": "first cat on hop fun",
    "iconUrl": "https://images.hop.ag/ipfs/QmUYrJD3CGxmvxdn6BmkiBRht3KZnrxAqWnj7fxtVB4QyW",
    "decimals": 6
  },
  "0x82d3b9f9d886c6bf6671f8f8e19e13e3f4a3bca0c0daee1930d5ac7d52a697c1::hopcat::HOPCAT": {
    "name": "HopCat",
    "symbol": "HopCat",
    "description": "First cat on Hop Fun",
    "iconUrl": "https://images.hop.ag/ipfs/QmRa3XL5YZQhDzQQyyEzCH1Q9aGaDEGe6N1TXkwyeNDRtP",
    "decimals": 6
  },
  "0x71f21ebf33eac29b52c6e24177d58a6b8283c6cee4f5ee50ef0f5fbdf986a962::gud::GUD": {
    "name": "GUD COIN",
    "symbol": "GUD",
    "description": "DIS COIN IS GUD COIN. VERY GUD.",
    "iconUrl": "https://images.hop.ag/ipfs/QmSbyWJn388iyGLjvKbK2nBcXtqGevcZr6YHU2xLmeWuSa",
    "decimals": 6
  },
  "0x3e295b4b835496ee460a503d37f44e1696cc2d38d2c27484ae8ab542e3ce0fe1::hoprabbit::HOPRABBIT": {
    "name": "HopRabbit",
    "symbol": "HopRabbit",
    "description": "First rabbit on hopfun",
    "iconUrl": "https://images.hop.ag/ipfs/QmcjJPsYoytXCn1dmZDqXuCkv8rxn8N19iY7UGMQLsB7ya",
    "decimals": 6
  },
  "0xa8b75b04cf62cfd9656513bd74cc7abba9862272d56f144eaa9a08194844e897::pnut::PNUT": {
    "name": "Peanut Squirrel",
    "symbol": "PNUT",
    "description": "$PNUT On SUI, Paying Homage to the social media star squirrel.",
    "iconUrl": "https://images.hop.ag/ipfs/QmWf2zuxzRm2F6kVzorpcV9TKtbrJ6bPKJkoA3hZDan11D",
    "decimals": 6
  },
  "0x7909c84b318301eec0ac363d2d0a9dc4afb95e7bc48b0dfc5161aef6a22d9532::pnut::PNUT": {
    "name": "Peanut",
    "symbol": "Pnut",
    "description": "Peanut Lives on in our hearts!",
    "iconUrl": "https://images.hop.ag/ipfs/QmbnhjUbZ2xop1AaCY1nPcG5YAcqJy8mMFSnj9Bbu5Ewav",
    "decimals": 6
  },
  "0x1243640212f32afbf73890529cd404e6b0f63993cdb9636f7ac8d4e4f8fc1601::hopape::HOPAPE": {
    "name": "HopApe",
    "symbol": "HopApe",
    "description": "HopApe",
    "iconUrl": "https://images.hop.ag/ipfs/QmeCQDaeW7MJoZpHHgNU3jbyrquuugPKM2DZYv5L6wLJbc",
    "decimals": 6
  },
  "0x336fb6bb95b6601e41a46b8fb7364e740d89ad8fbbe3e063cbd5a810389ddc40::rex::REX": {
    "name": "suirex",
    "symbol": "REX",
    "description": "suirex Something big and blue is coming exclusively to @hopaggregator",
    "iconUrl": "https://images.hop.ag/ipfs/QmdHT3ngjCqfn7wq8n668cnrqkaJ8YFQwt5GAghiQ9STuq",
    "decimals": 6
  },
  "0xeddf484e1ca2179eb9523cc003f26fd88f1a12b8c16c7c8b3c0ca89df2e32428::wrz::WRZ": {
    "name": "WorlldZ",
    "symbol": "WRZ",
    "description": "just try to have fun",
    "iconUrl": "https://images.hop.ag/ipfs/QmYMUF8UshRkcepq3PByxGC9aB9Z2WvRntkDPbvtQm5fGJ",
    "decimals": 6
  },
  "0xe9af0525761f274dc1ceb24c4dbb36a95393efc8e551b6fa5a0f322d79fc0c6b::hopdog::HOPDOG": {
    "name": "HopDog",
    "symbol": "HOPDOG",
    "description": "Hi, I'm Hop, the coolest cat on sui! Yes, I'm like that - I jump like a real athlete, or almost... But every jump I make is a show! Also, between you and me, I love to eat fish.",
    "iconUrl": "https://images.hop.ag/ipfs/QmRKatJ6xRmwJ3NaK24UT4WwU7YqCXTWQbTf4AZGnHPn8R",
    "decimals": 6
  },
  "0xdb2cf357c80a0b8d8b10fff25f594e721d82548a2bcb1a1fc24d7da8a731a3a8::boi::BOI": {
    "name": "$boi",
    "symbol": "boi",
    "description": "$boi",
    "iconUrl": "https://images.hop.ag/ipfs/QmVKNL7YQh94J1pUzwUqg8hdg84bkLjr2HXi6CEFk1zpK4",
    "decimals": 6
  },
  "0x85245e7fc36ae68cab210fff39359c84fabc8b79370e71ce06d320b61f57d0c6::hopmoon::HOPMOON": {
    "name": "HOP MOON",
    "symbol": "HOPMOON",
    "description": "Hop fun to the moon",
    "iconUrl": "https://images.hop.ag/ipfs/QmfVVYWYUPjqJEZS5qv9GKi1bhpxs54Cc8yCMHwdNz6jHe",
    "decimals": 6
  },
  "0x4b49e232bdd1e193c1d3192547a6429cafb4047c8ddcbdd831451fe743ab3141::hop::HOP": {
    "name": "HOP",
    "symbol": "HOP",
    "description": "on for the best price with zero fees",
    "iconUrl": "https://images.hop.ag/ipfs/QmQvXjD3avEQMvYHXK9qGPA4ixefgNqQXLfyd5PiLz3y35",
    "decimals": 6
  },
  "0x1788c208e0f0c3a0b9119119c7bc37069012f8528d84178b1583570c34531b48::pepe::PEPE": {
    "name": "pepe",
    "symbol": "PEPE",
    "description": "first meme coin on sui",
    "iconUrl": "https://images.hop.ag/ipfs/QmZMTnCNYincJTTtNvxptHGEnB36F336C544Re5Zjo2QLj",
    "decimals": 6
  },
  "0x6e5b36fb9d6502e7350e9da4c6b18eac4e00108cb17d993248c0b2001699c6f8::rachop::RACHOP": {
    "name": "RachopSui",
    "symbol": "Rachop",
    "description": "Rachop is one of the first memetokens to launch on Hop Fun, centered around the story of a raccoon with a growing passion for piracy and his loyal rabbit companion, Hop. This project isn’t just a memetoken—Rachop will soon evolve into an NFT collection and a game that allows players to experience the adventures of Rachop and Hop firsthand. By playing through various scenarios in the game, players can earn additional Rachop tokens.",
    "iconUrl": "https://images.hop.ag/ipfs/QmSGQUYn4MkWnYhmmQCaf7rWtNrd4zcsRt1UWr9HFZ1Mgp",
    "decimals": 6
  },
  "0xf58742b8df1e378d8dc6d847c6ece15beeaa7d00cb4872d8fffc4266a489972f::bruce::BRUCE": {
    "name": "Bruce the Sharky",
    "symbol": "BRUCE",
    "description": "Hello fren, you have entered Bruce's ocean but that does not make you part of the Sea.",
    "iconUrl": "https://images.hop.ag/ipfs/QmZQFWExhtGNwLbeR4Qpv5T6FcPRQew2WbYrh8Su82xCst",
    "decimals": 6
  },
  "0x71bfbdc71303564c972a3cf666f9267c53cd27e39157acc886f42106e0c6f9b1::toad::TOAD": {
    "name": "BLUE TOAD",
    "symbol": "Toad",
    "description": "Toad",
    "iconUrl": "https://images.hop.ag/ipfs/QmUJbs9ahB2JLBZ5ADaggGDnJTZP3gU8ZhtXYFdZva97nP",
    "decimals": 6
  },
  "0x19fa591719f5d4988833e267f844ae9980bbfd228508a36b963990c91e6e1b8c::bubl::BUBL": {
    "name": "BUBL",
    "symbol": "BUBL",
    "description": "Bubbling on \n@SuiNetwork\n to make frens",
    "iconUrl": "https://images.hop.ag/ipfs/Qmbf752EhgUbtLCph7R9XuGYiWQZKdyCaA8Bo5afpwuNiY",
    "decimals": 6
  },
  "0xe39251444808ea8ff5f746cf4a7d4dfac808b699e03adf8495ca8d2df0d7c483::hdog::HDOG": {
    "name": "HopDoge",
    "symbol": "HDOG",
    "description": "First DOGE on hop.fun",
    "iconUrl": "https://images.hop.ag/ipfs/QmQ1Zdkn6oHcD84WQTSgsa4GVvnobAgnjHq5MLwT54Ubf2",
    "decimals": 6
  },
  "0x70ab8b80afef4ce52d44f8bd7936b46eb5ed7d2d12ee17bb5c4c03dcf35e7362::msi::MSI": {
    "name": "messui",
    "symbol": "msi",
    "description": "messi is the king",
    "iconUrl": "https://images.hop.ag/ipfs/QmUSNPmaR2RsHzHyeMgv4XzCrunu5RH5akyQd8nQCFdmH4",
    "decimals": 6
  },
  "0xcf3c00eaeb7e3de07075341619351c936c9d455e4f6e93ef1d35f5b1d1d8dc62::bubl::BUBL": {
    "name": "BUBL",
    "symbol": "BUBL",
    "description": "Bubbling on \n@SuiNetwork\n to make frens",
    "iconUrl": "https://images.hop.ag/ipfs/Qmbf752EhgUbtLCph7R9XuGYiWQZKdyCaA8Bo5afpwuNiY",
    "decimals": 6
  },
  "0x99192eed432b68339d1368c9fd36e6ce3172b0ac39ef49ee08157cba7b973651::blue::BLUE": {
    "name": "Sui is Blue",
    "symbol": "BLUE",
    "description": "Sui Color is Blue",
    "iconUrl": "https://images.hop.ag/ipfs/QmTXNQMd43G96w6ugbCEeaCZEeZdKyCUFGxLjRRjAECTXB",
    "decimals": 6
  },
  "0xe909e777e6cc3030858d88bfe901970d03cb6209a3445fdad61f4a3e7308c7c8::suibbu::SUIBBU": {
    "name": "Suibbu The Crab",
    "symbol": "SUIBBU",
    "description": "Meet Suibbu the Crab, a deity of sideways market actions",
    "iconUrl": "https://images.hop.ag/ipfs/QmTupQs35dErnPeKhqo7ZD5JzfWPKwv3cb4rqbpukmPTby",
    "decimals": 6
  },
  "0x4c800d296b0c419e91f5f64b3179f4210197222b9206c985da9c19f2a3b49afe::suiyan::SUIYAN": {
    "name": "Super Suiyan",
    "symbol": "SUIYAN",
    "description": "Super Suiyan the blockchain",
    "iconUrl": "https://images.hop.ag/ipfs/QmQHXY6CFB9iCvec79J5YqUMozWzcMFuEfqYK658MuBJ2S",
    "decimals": 6
  },
  "0xd5890d08f42cce8a2186d4a5b0c3469306209af6558fc545eed4f10e718cc72b::pepe::PEPE": {
    "name": "pepe",
    "symbol": "PEPE",
    "description": "first meme coin on sui",
    "iconUrl": "https://images.hop.ag/ipfs/QmZMTnCNYincJTTtNvxptHGEnB36F336C544Re5Zjo2QLj",
    "decimals": 6
  },
  "0x724bc7349fc5984bfd5159fa5c9d7680a99871bc087e02a6d4256f47b5898882::sip::SIP": {
    "name": "SUIPEPSI",
    "symbol": "SIP",
    "description": "SUI PEPSI",
    "iconUrl": "https://images.hop.ag/ipfs/QmaS9xVSRRUZAttHzzUmsy9rhnN6QD1v5QfbCpM3AR9Yas",
    "decimals": 6
  },
  "0x9a5e61995c97c54047cc5b1ba9f33ede09f8b83b067383192cc34577944323ce::pepe::PEPE": {
    "name": "PEPE",
    "symbol": "PEPE",
    "description": "firts pepe",
    "iconUrl": "https://images.hop.ag/ipfs/QmNLswqyxnmqFFsnK9nVf6x8ht5BV6t8weiZQMJHeRw8Dc",
    "decimals": 6
  },
  "0x2db4802b9bbacba5b09e8608b073822e347df90069ef521af4603c8e5fad1977::luce::LUCE": {
    "name": "Official Mascot of the Holy Year",
    "symbol": "LUCE",
    "description": "The Vatican has unveiled the official mascot of the Holy Year 2025: Luce (Italian for Light). Archbishop Fisichella says the mascot was inspired by the Church's desire \"to live even within the pop culture so beloved by our youth.\"",
    "iconUrl": "https://images.hop.ag/ipfs/QmNzBmiH4jsj8486brCKCyb5qCu1XuJAwe4z7uBPu8RgPu",
    "decimals": 6
  },
  "0x8d66961e572809bc9ebe7e5c815d245a3876f64e98ac899834d350f93e4b70e1::swojak::SWOJAK": {
    "name": "Scuba Wojak",
    "symbol": "SWOJAK",
    "description": "Wojak on Sui",
    "iconUrl": "https://images.hop.ag/ipfs/QmZyJjJi6qZDSKv8ajBF3nZxbq3AiY67jjXGkJPffaJxRF",
    "decimals": 6
  },
  "0xd66c2962485722dc0b66e5c04e5e1a04e31d38dfa385cc1c1cfb18fcd78a3a66::cat::CAT": {
    "name": "HopCat",
    "symbol": "Cat",
    "description": "HopCat",
    "iconUrl": "https://images.hop.ag/ipfs/QmWBx34wFTPcn8u2RbT74MgrKZFcbssptPYZWZhHQM3oVt",
    "decimals": 6
  },
  "0xe5f4cc66045cd848c4f03158b0fcdc3f34632821c3ceea8186e8f41e3840910e::suirex::SUIREX": {
    "name": "$REX",
    "symbol": "suirex",
    "description": "“su-rɛks”. Something big and blue is coming exclusively to \n@hopaggregator\n. I’m Rex, we’re all Rex…. trade REX using moonlite",
    "iconUrl": "https://images.hop.ag/ipfs/QmdHT3ngjCqfn7wq8n668cnrqkaJ8YFQwt5GAghiQ9STuq",
    "decimals": 6
  },
  "0x111737546136c229fc13affad74044752cabbd66ccbd4535cbbde77da8f42243::funny::FUNNY": {
    "name": "FUNNY HOP",
    "symbol": "FUNNY",
    "description": "The Bunny is setting up the stage .... Next show: wen \n@hopaggregator\n  decides to launch HOP FUN |  TG:  https://t.me/funnythebunny",
    "iconUrl": "https://images.hop.ag/ipfs/QmWmcL1puULZohnKkJVwkbdUvrpUiNVXDs1SqfoyMWKfpm",
    "decimals": 6
  },
  "0x92716fb26f1c1f7a49de4bff22af52ac91fadd77ad7211221205532047c67139::bb::BB": {
    "name": "Blue Balls",
    "symbol": "BB",
    "description": "Blue balls on SUI",
    "iconUrl": "https://images.hop.ag/ipfs/QmWR9cLBYcE1MFRp3YyEvHjWLe5q7qP6KPkkbJsGtfZZB3",
    "decimals": 6
  },
  "0xef370a8773eb53367acb4af1d51b286c143aa57a49afe75651aaa9b5f10fd530::sdog::SDOG": {
    "name": "Sui Dog Wif Hat",
    "symbol": "SDOG",
    "description": "Shiba with hat",
    "iconUrl": "https://images.hop.ag/ipfs/Qmd4F8GBXDgwuWXmBoHAzSwncnbhnNGXevijZZ9QPEMdde",
    "decimals": 6
  },
  "0xd852c2a215253d5d4c1b5c444a08660cbb239c61ac601dc284cae249196d939d::sdog::SDOG": {
    "name": "SUI Dog wif haglt",
    "symbol": "SDOG",
    "description": "Shiba with hat on sui",
    "iconUrl": "https://images.hop.ag/ipfs/Qmd4F8GBXDgwuWXmBoHAzSwncnbhnNGXevijZZ9QPEMdde",
    "decimals": 6
  },
  "0xfacc7fa035b257b8102763d1d862a65020ea293680c7e15a17d9e75daf7d12c4::rachop::RACHOP": {
    "name": "RacHop",
    "symbol": "RacHop",
    "description": "With my bunny by my side, Rachop and I take the ride",
    "iconUrl": "https://images.hop.ag/ipfs/QmSGQUYn4MkWnYhmmQCaf7rWtNrd4zcsRt1UWr9HFZ1Mgp",
    "decimals": 6
  },
  "0xcd4ed875925cf784baf93d8a72dbeae0d03a845cc146570bb39ece644a67a517::suicat::SUICAT": {
    "name": "SUI CAT",
    "symbol": "SUICAT",
    "description": "SUICAT: The boss cat of SUI chain",
    "iconUrl": "https://images.hop.ag/ipfs/QmbXnWcnTmsPDuXdhcyfhxxXxKS4HKGaYNYHQhX3hqbQ6f",
    "decimals": 6
  },
  "0xb97cef2708ca6d923f5db39c438f1df49c1c055469c4b6558b26b0a71ad5725f::chop::CHOP": {
    "name": "Chop Sui",
    "symbol": "CHOP",
    "description": "Delicious. Sideways. Side of sum-yun gai.",
    "iconUrl": "https://images.hop.ag/ipfs/Qme4mtJfcXaxR8Cw7iYxSMW5UMd5pYykbbTyzZgEjCT6Wh",
    "decimals": 6
  },
  "0x223d8d4676acce50e57edb9c48aca2df1c871870d6633a1b7a8e5c59ebcb6e1f::hopj::HOPJ": {
    "name": "Hopthisjerks",
    "symbol": "HopJ",
    "description": "Took too long",
    "iconUrl": "https://images.hop.ag/ipfs/QmYM5UY4BWvkccQRTYwe5j38usBjGe7iGdZTBEVe4XRbA8",
    "decimals": 6
  },
  "0x25f94cc7689695cc6389fb1daefc28d418a07159a07de15421367141ac9e78c8::nuts::NUTS": {
    "name": "deeez",
    "symbol": "NUTS",
    "description": "for fun",
    "iconUrl": "https://images.hop.ag/ipfs/QmZwieh4F7LBqGYUE6HLqmC6HJPNtdpRjHoGjk5u7pMxW9",
    "decimals": 6
  },
  "0x5cfd8b2f1f10244d33373b07ce5b059ef5cb7af74b14d2f8dacf865775e50605::wow::WOW": {
    "name": "wow",
    "symbol": "wow",
    "description": "wow",
    "iconUrl": "https://images.hop.ag/ipfs/QmQcYx6XHXSg9tPncEtHH6rJc9PkDhPTdBZQnUgd8rLUVT",
    "decimals": 6
  },
  "0xb6ec6838d4e3c673ae78e16b528731fff6915873146810d3fe07e2891d44d6af::hehe::HEHE": {
    "name": "HEHE",
    "symbol": "HEHE",
    "description": "HEHE",
    "iconUrl": "https://images.hop.ag/ipfs/QmR1Rhbs6rBpoyHsnLvmdia7JfYF4TvdfT2uL3SWJxSZWb",
    "decimals": 6
  },
  "0xc94eb248239231afc218410057f52df73636e42b14f35e28b02f5b025f6d69b8::gandon::GANDON": {
    "name": "gandon",
    "symbol": "gandon",
    "description": "real GANDON is HERE!",
    "iconUrl": "https://images.hop.ag/ipfs/QmXv1HcSMWxh5g9foYs4yMrXup4VmAhrUwf8QLWbY1gXcV",
    "decimals": 6
  },
  "0xa155958ad5c9ca1632e75a0f9ae7458b636782eddffd20398a33f10bad825fea::bonkman::BONKMAN": {
    "name": "BONKMAN",
    "symbol": "BONKMAN",
    "description": "Bonkman the base coin",
    "iconUrl": "https://images.hop.ag/ipfs/Qmd9jqXeosv6cZU3orK2kw5EpMhZdR4Scuvwq3ds1rPjoV",
    "decimals": 6
  },
  "0xa476530e3c40f262b493812e56bcb253b5c085ca541f84d2cf45306235d45ac7::goat::GOAT": {
    "name": "Sui GOAT",
    "symbol": "GOAT",
    "description": "Sui GOAT 🐐",
    "iconUrl": "https://images.hop.ag/ipfs/QmbPR3XjpsTdzZN9i7Fa1Bp2acBmnxC5J4tK2Sf39MzEhc",
    "decimals": 6
  },
  "0xe3576a72f3173d09167348258a1f39ff267615fcb1ea9fe1a9a5d08301949934::fuckhop::FUCKHOP": {
    "name": "hop dick",
    "symbol": "FuckHop",
    "description": "rug time fuck hope",
    "iconUrl": "https://images.hop.ag/ipfs/QmUVtyA2K1SLwjvx43XoAvHFteiXbLC7RSr6ccMJbMyF1C",
    "decimals": 6
  },
  "0xcb709905655391e363e2010df55f74a12a632053e2ed48dd2d5e6b9748877712::dragon::DRAGON": {
    "name": "Totem-Dragon",
    "symbol": "Dragon",
    "description": "Dragon is a mythological creature with great symbolic significance, which is often regarded as a symbol of power, dignity and auspiciousness. Unlike Western dragons, Chinese dragons are usually depicted as long, snake-shaped, four-clawed and capable of controlling water, rain and wind.",
    "iconUrl": "https://images.hop.ag/ipfs/Qme4VYBTkrRKRD2qt29QTK5rg5h5jx9BGAcHzE2CgoUAWc",
    "decimals": 6
  },
  "0x5d4d21ddf1aa2a79089cc0fe5e15b96bd8d4fa0272ac991e909da1393042da2f::pnut::PNUT": {
    "name": "Peanut the Squirrel",
    "symbol": "Pnut",
    "description": "It’s our job to help save peanut the squirrel! Sign the petition!",
    "iconUrl": "https://images.hop.ag/ipfs/QmPncTECroMSa7Jxth4R69QNWwssKFFLrQ5Nvh3rFqBiJP",
    "decimals": 6
  },
  "0x71a8c329b4894f76c86c3aed0689e345033549d640a330670c46a10e2a70ae46::verse::VERSE": {
    "name": "SUIVERSE",
    "symbol": "VERSE",
    "description": "SUIVERSE",
    "iconUrl": "https://images.hop.ag/ipfs/QmP5yt8khqwvXUK2DDPw6joXk5L1BrXJcjrfnGfvAWXNBd",
    "decimals": 6
  },
  "0x4640c793cecc315c1e66d1adcc9825639f0274f4b609e02ac449273c0af85e7c::paws::PAWS": {
    "name": "PandaPaws",
    "symbol": "PAWS",
    "description": "PandaPaws is more than a fun meme coin; it inspires its community, the \"Panda Guardians,\" to embrace environmental stewardship.",
    "iconUrl": "https://images.hop.ag/ipfs/QmeDe8Cy2kSsUa9zWuz3wt5pbpBrdysEWZBm2aGnHCmj2H",
    "decimals": 6
  },
  "0xc52ae95fc0d0d3176998c2aace2ad76a251f98d34cd2f43e92ea33837f66abc7::sheikh::SHEIKH": {
    "name": "SHIEKH ZACK MORRIS",
    "symbol": "$SHEIKH",
    "description": "SHEIKH ZACK ON SUI",
    "iconUrl": "https://images.hop.ag/ipfs/QmUyRaShHAaRyZoKn2tkxgnnMa4QvQLfmkMu9bewBjaXkN",
    "decimals": 6
  },
  "0x09e66b9cb979692be470168c58a65138566ba6b1e3a06150ce40e481cdbe977c::nutsui::NUTSUI": {
    "name": "Peanut Squirrel on SUI",
    "symbol": "NUTSUI",
    "description": "Peanut Squirrel embodies resilience, rebellion, and charm on the SUI network! Once a scrappy survivor turned Instagram sensation, $NUTSUI honors Peanut's wild journey and spirit. He’s a symbol of tenacity for the crypto crowd, representing those who refuse to be tamed. With $NUTSUI, Peanut lives on as the network’s cheeky champion—ready to disrupt, inspire, and rally the underdogs.",
    "iconUrl": "https://images.hop.ag/ipfs/QmPPR5ywjHCa4jz9grD86sCymCxW8DNG87ywn4RnE4pcFL",
    "decimals": 6
  },
  "0x83ce7fd0a8819c815f06823e4435c572fcf31d36dbd5f5b9325954696df4bab6::nossy::NOSSY": {
    "name": "catwifnosejob",
    "symbol": "NOSSY",
    "description": "The cat has a fokin nose job and he is very cute",
    "iconUrl": "https://images.hop.ag/ipfs/QmdQ8fAoh7yTWaydLewGcRd3b9ykCz7t6UdfhqHhVLtLPP",
    "decimals": 6
  },
  "0xe0bb454911c926076f15fb8fcc0cb4ac16d1e3754e700d3abe867d39841abc24::trumpsui::TRUMPSUI": {
    "name": "TRUMPSUI",
    "symbol": "TRUMPSUI",
    "description": "We will make SUI great Again",
    "iconUrl": "https://images.hop.ag/ipfs/QmfQmCpSA9aZ1oDSBkcHYwJNZgNFUktLqxCzEBR2rHVkyv",
    "decimals": 6
  },
  "0x1ee9e24487d278887dafbe1bb0cd2ae617f6afc390a19d97aeb1551e3b827ef6::xsn::XSN": {
    "name": "XONO",
    "symbol": "XSN",
    "description": "I AM XONO",
    "iconUrl": "https://images.hop.ag/ipfs/QmXxcaytQghx8vJuhgJNPyCTc939NthS8zr8R4F1DFL7o4",
    "decimals": 6
  },
  "0x7ad0db94c830ccda11ab71a4a1ba56deca21faa43ba5f086a6d34e09b7a00d76::bonkman::BONKMAN": {
    "name": "bonkman",
    "symbol": "bonkman",
    "description": "fun",
    "iconUrl": "https://images.hop.ag/ipfs/QmWn6rXpcFnNjaZzGBaPKDz6yYqnW2H6AE82DHnUPDP4J6",
    "decimals": 6
  },
  "0xf2d0df7cc1c1f683a4d6276304ff859d6f7ce8bdead7d8854d22643102fd44d4::bunni::BUNNI": {
    "name": "Bunni",
    "symbol": "BUNNI",
    "description": "The first bonded token on Hop.ag/fun",
    "iconUrl": "https://images.hop.ag/ipfs/QmcyG4vT1pbjTcHHVEtyCN1s6G4Sd7UEQFA7ayLUYgmMyN",
    "decimals": 6
  },
  "0x3c8d27c982b063a32f2e7fa3cc663df336b21fbd5f26e8860c1b92a267dceccb::apnut::APNUT": {
    "name": "armoured pnut",
    "symbol": "apnut",
    "description": "armoured pnut would have never lost to those commies",
    "iconUrl": "https://images.hop.ag/ipfs/QmQF38oVRC6JLp1dZAKXi5EtiESe4jNy2L1G7Wm6mh8iRY",
    "decimals": 6
  },
  "0xebd3cc21316f77355cd704f11a28ded4572507e6b3de496e89f34d75f26ec3c4::hops::HOPS": {
    "name": "HopSoul",
    "symbol": "HopS",
    "description": "Swap on @SuiNetwork and Pray Hard for HopSoul",
    "iconUrl": "https://images.hop.ag/ipfs/QmTDKRGW9SF1yf5ShGVdV6Hr3AbraDP2t7oxPTEgAYgwSj",
    "decimals": 6
  },
  "0xd54a80ddb0490d135d05b326e4e08b747874047e5720d191b4febd634493fc83::yummy::YUMMY": {
    "name": "YummyOnSui",
    "symbol": "YUMMY",
    "description": "Yum the fck out $YUM put on your SUIglasses",
    "iconUrl": "https://images.hop.ag/ipfs/Qmd3rykaDny8QZhWrmFvRMcmyTGSJzbmJHAiqfKKBmZFQZ",
    "decimals": 6
  },
  "0x5f29db0672874096398206c241fe5b2b56140043d16db4713b4e909823572382::luv::LUV": {
    "name": "Missing Luv",
    "symbol": "LUV",
    "description": "When it’s all you got to give, give luv. Organic. Community. Purpose. ❤️",
    "iconUrl": "https://images.hop.ag/ipfs/QmUWzHyv7hnYZtyyyX8ne88huNuGuunNVKAdaXZn3t8m8S",
    "decimals": 6
  },
  "0x5332a42fea343f2000831653e898b626e36a03161fe85418d754e32d93756162::hop::HOP": {
    "name": "Hoppy",
    "symbol": "HOP",
    "description": "Hoppy the official hop frog of hop",
    "iconUrl": "https://images.hop.ag/ipfs/QmPVm7QNhstGvhyMnKoYHXVsmuwGRcyT4riAH2Z3cr1MgG",
    "decimals": 6
  },
  "0xd8fed11cfecf4d9fa648a81ddb82e2b800d8b674d2ca741be2dca4c477d34bec::goat::GOAT": {
    "name": "Goatseus Maximus",
    "symbol": "GOAT",
    "description": "Parody of Goatseus Maximus (GOAT) a new meme coin created by Pump.fun user @EZX7c1 and adopted by an AI bot, Truth Terminal. It was named after a tweet by @truth_terminal and has gained popularity among the crypto community.",
    "iconUrl": "https://images.hop.ag/ipfs/Qmc8m42xdJYWrXB4DKx8Mo7BmgFGxPALqQPSccoyh18JvN",
    "decimals": 6
  },
  "0x27df1672f5fd2696b0fe9ae536897f724c6377d71184f7ebd2ae7c81c6b216cf::hoppussy::HOPPUSSY": {
    "name": "hoppussy",
    "symbol": "HOPPUSSY",
    "description": "First meme",
    "iconUrl": "https://images.hop.ag/ipfs/QmPoRLCfJUDefzMeB8CRnfxMqDfH96BuEmmSpMXQqq8DKA",
    "decimals": 6
  },
  "0x4733f8b64bd26c06b3c479bd71f4152b38d85ad5b9599af3cd3f1128ea5d546d::poring::PORING": {
    "name": "Poring",
    "symbol": "PORING",
    "description": "$PORING - The true meme token on $SUI blockchain with fair distribution on the hop.fun platform",
    "iconUrl": "https://images.hop.ag/ipfs/QmUu4Yz7JM48n45Buj4aeW1kd5Ew6wKiNjfQgGsz7Z7gwK",
    "decimals": 6
  },
  "0x04fc571b6b280bc652aa14231178d6db1c3f9d1269b8642b8567b84f81e332f0::hopcat::HOPCAT": {
    "name": "HopCat",
    "symbol": "HOPCAT",
    "description": "First cat on Hop Fun",
    "iconUrl": "https://images.hop.ag/ipfs/QmWBx34wFTPcn8u2RbT74MgrKZFcbssptPYZWZhHQM3oVt",
    "decimals": 6
  },
  "0x1d61730a419d1cdc0ddd6bb345a76c9ba9853732918cedccc4505989be8f8490::hopfun::HOPFUN": {
    "name": "HopFun",
    "symbol": "HOPFUN",
    "description": "The First Meme ever launched on Hop Fun Memecoin Fair Launch Platform on the SUI Network",
    "iconUrl": "https://images.hop.ag/ipfs/QmasV1aD6bC7yGm3bqAEJWXT5uiSQsQqcZvx1PdA1jUoCU",
    "decimals": 6
  },
  "0xaf0af4aa4de8c5b2db97996946075ec62e2e38e055cc5f5c1791348194a9da0c::degen::DEGEN": {
    "name": "Sui degen Apes",
    "symbol": "$Degen",
    "description": "Degen Apes wanting to degen",
    "iconUrl": "https://images.hop.ag/ipfs/QmYeV5vyP5ePiWtDZAZWY2dUHBY3qc1FXtETUArGhY5uYi",
    "decimals": 6
  },
  "0xc29fd16450a13be5ac505515c5b569b33bdaa10f259a2eeae1a71a0d679d6670::hpf::HPF": {
    "name": "Hopfun",
    "symbol": "HPF",
    "description": "the first token on hop fun",
    "iconUrl": "https://images.hop.ag/ipfs/QmQvXjD3avEQMvYHXK9qGPA4ixefgNqQXLfyd5PiLz3y35",
    "decimals": 6
  },
  "0xbdae394b9adea9676ecf4c2fe098a322f06919e1c51d5327a6e0606ea3aaf719::pnut::PNUT": {
    "name": "Peanut the Squirrel",
    "symbol": "Pnut",
    "description": "RIP",
    "iconUrl": "https://images.hop.ag/ipfs/QmQUAM2io5bgcf1LU8ocEhwnp11p318nPxrQQcNBXHPq2d",
    "decimals": 6
  },
  "0x8ba7b4dc853d243d2118b395f1319a8e0de09cea487a1a74fef3f63173308a11::dik::DIK": {
    "name": "Dog In Kimono",
    "symbol": "$Dik",
    "description": "$Dik the dog is a Dog In a Kimono. He brings the luck. No Twitter, Just a ticker!\nDik the Dog isn't just another Dog in the crypto world. It's a symbol of hope and resilience, representing the potential for dogs to thrive in a dog-eat-cat world now on sui",
    "iconUrl": "https://images.hop.ag/ipfs/QmZsaDzbPs5NbGCRkN2HhnnBEavmDviESkxWYJbcVXPitY",
    "decimals": 6
  },
  "0x79ebdb4563c3de2d660899651fa2beea321dbc764af6caf1586e90d52e489271::suit::SUIT": {
    "name": "SUITARDIO",
    "symbol": "SUIT",
    "description": "World Domination",
    "iconUrl": "https://images.hop.ag/ipfs/Qman7tyZQTrW7fZDJ4xPxRtE6sEy7eipTKnGwAMgt56mAM",
    "decimals": 6
  },
  "0x744cdf261a4cb78d982b38d46a7e0d6900633a829a042651bfa43a5c86456629::hfish::HFISH": {
    "name": "HOP FISH",
    "symbol": "HFISH",
    "description": "They say fish don't jump, I'm an exception #sui",
    "iconUrl": "https://images.hop.ag/ipfs/QmWJJR8B1qXzo6FpAq4YGtPKYBNZWFA4AqZdrSPVZSBUJs",
    "decimals": 6
  },
  "0x1481f80d016da65f64a6008d6314cf918335e13c46c53adee48d4995b05e61bc::pnut::PNUT": {
    "name": "Obi PNut Kenobi",
    "symbol": "Pnut",
    "description": "f you strike me down, I will become more powerful than you could possibly imagine” Obi PNut Kenobi\n-Elon Musk",
    "iconUrl": "https://images.hop.ag/ipfs/QmbZT8enZzdJURSpWyxf6MipPM5UKRXdx93HbDQrfw1m9x",
    "decimals": 6
  },
  "0x0e6469f65bcdc68a1f822bd74a29de93adcf6585fe46524cb1ee122bab80a3aa::notfun::NOTFUN": {
    "name": "HOP_NOT_FUN",
    "symbol": "NOTFUN",
    "description": "Phase 3 coming",
    "iconUrl": "https://images.hop.ag/ipfs/QmfVjjkcGghv6gKJCqdib7dUcFDbvjx4n7inJaRMykaVdf",
    "decimals": 6
  },
  "0x2eca8de7f34403a6a279ce0719e8301f4416279a7bf0b0cb7a3c88b02a53770a::dragon::DRAGON": {
    "name": "Totem-Dragon",
    "symbol": "Dragon",
    "description": "Dragon is a mythological creature with great symbolic significance, which is often regarded as a symbol of power, dignity and auspiciousness. Unlike Western dragons, Chinese dragons are usually depicted as long, snake-shaped, four-clawed and capable of controlling water, rain and wind.",
    "iconUrl": "https://images.hop.ag/ipfs/Qme4VYBTkrRKRD2qt29QTK5rg5h5jx9BGAcHzE2CgoUAWc",
    "decimals": 6
  },
  "0x78fa5b1ae635dbdf6dec9fcf5e61869dbd9c3038399038a4aa0afe26a203d89f::papa::PAPA": {
    "name": "Papa on sui",
    "symbol": "PAPA",
    "description": "Meet papa, the mastermind ready to dominate the Sui blockchain scene. With street-smart strategy and unmatched influence, papa's about to take over and set new rules in the game of innovation and leadership.",
    "iconUrl": "https://images.hop.ag/ipfs/Qma4cHoUL3SkNopMHUeC52kgemZJCaoAUP4a7zjLDf7rH2",
    "decimals": 6
  },
  "0x37a8be46dfb5bb0bbe24ecfec81d815a879f2eb4bfdd87ac19d52daa57974b55::pnut::PNUT": {
    "name": "Obi PNut Kenobi",
    "symbol": "Pnut",
    "description": "f you strike me down, I will become more powerful than you could possibly imagine” Obi PNut Kenobi\n-Elon Musk",
    "iconUrl": "https://images.hop.ag/ipfs/QmbZT8enZzdJURSpWyxf6MipPM5UKRXdx93HbDQrfw1m9x",
    "decimals": 6
  },
  "0x0f3334f0844ece6df3a7c24252acf406cf57fbeb5616efca06c68d8c7124cdf2::mist::MIST": {
    "name": "MIST",
    "symbol": "MIST",
    "description": "MIST",
    "iconUrl": "https://images.hop.ag/ipfs/QmXfKoD1f6dLvD3iBJ4v5TnaB75bnDiuMutDnQyknmr6zE",
    "decimals": 6
  },
  "0x947fba1d02183f75de3299d90a99d5fff2ffe20a6bd8c83a161b853936aaa43a::pnut::PNUT": {
    "name": "Obi PNut Kenobi",
    "symbol": "Pnut",
    "description": "f you strike me down, I will become more powerful than you could possibly imagine” Obi PNut Kenobi on SUI\n-Elon musk",
    "iconUrl": "https://images.hop.ag/ipfs/QmbZT8enZzdJURSpWyxf6MipPM5UKRXdx93HbDQrfw1m9x",
    "decimals": 6
  },
  "0x6a554724498a49c35559453a7903cd7e3a522bf7ad74eac127311a1e73d679bb::dik::DIK": {
    "name": "Dog In Kimono",
    "symbol": "$Dik",
    "description": "$Dik the dog is a Dog In a Kimono. He brings the luck. No Twitter, Just a ticker!\nDik the Dog isn't just another Dog in the crypto world. It's a symbol of hope and resilience, representing the potential for dogs to thrive in a dog-eat-cat world now on sui",
    "iconUrl": "https://images.hop.ag/ipfs/QmZsaDzbPs5NbGCRkN2HhnnBEavmDviESkxWYJbcVXPitY",
    "decimals": 6
  },
  "0xc6c993b4a13c426f4430a4db3fa91c56f9ec594d326f253cdb4d98cadd16a38b::suidoku::SUIDOKU": {
    "name": "Suidoku",
    "symbol": "SUIDOKU",
    "description": "Suidoku\n\nSuidoku is a concept that combines the classic game of Sudoku with the Sui blockchain ecosystem. It could involve a puzzle game where users solve Sudoku grids to earn rewards or participate in a decentralized experience. In such a project, players might engage with blockchain-based assets or incentives, with rewards distributed through smart contracts on Sui.\n\nThe core idea of Suidoku could revolve around gamification in the crypto space, creating a fun and rewarding environment while promoting user interaction with the Sui network. This project could attract both puzzle enthusiasts and crypto enthusiasts, merging logic-based gameplay with the innovative blockchain features of Sui.",
    "iconUrl": "https://images.hop.ag/ipfs/QmeGwAY8xRRu4NDqCQFAjpXMckjS4HjjjjepuLjULHHyMg",
    "decimals": 6
  },
  "0xeb56b41ffe075460960c5a03ba6a5b5c210c7f58d0885a2749535efc847138e0::suina::SUINA": {
    "name": "SUINA The Pig",
    "symbol": "SUINA",
    "description": "Just a pig looking for truffles on SUI",
    "iconUrl": "https://images.hop.ag/ipfs/QmSAkqGpWYUMLTYGa2v6EPkGJVPM4DiMogBzUAsn6F21aG",
    "decimals": 6
  },
  "0x9c6a5d7d40218b5b1a8d5893075b544c47407374a1f800e0e4172b95e46b9e33::qeewwer::QEEWWER": {
    "name": "QEEWWER",
    "symbol": "QEEWWER",
    "description": "42",
    "iconUrl": "https://images.hop.ag/ipfs/QmaKAAxb9cVRwscXZhpBiQsmTjuqaVkLrXUs9jDV9txzRx",
    "decimals": 6
  },
  "0xee2a6e5e8a5413b71cd9cd90ccedf6a6e00217df02adf4a07bc65287ff2342cd::hophip::HOPHIP": {
    "name": "Hop Hip",
    "symbol": "HOPHIP",
    "description": "hophip.",
    "iconUrl": "https://images.hop.ag/ipfs/QmPwgfqwdxAPAh9kkRHG6ZQ4XY2xYJr8TEzxY8kmA4h3MM",
    "decimals": 6
  },
  "0xdad461789a33099cd20a610c49e40c13b3d431e68e8b8e04326c16ad60b8f3e8::pac::PAC": {
    "name": "PacMoon SUI",
    "symbol": "$PAC",
    "description": "Devs abandoned  PACMOON $PAC on blast. Community is bringing it back on SUI. Join community",
    "iconUrl": "https://images.hop.ag/ipfs/QmaRFQwFAUMGZM8pGLfDm9eLBZavpWtgxTBXK9x9e5ZzLQ",
    "decimals": 6
  },
  "0xefabcce7d0aae789f7daf945d4ab23a23b81d908199fe9299e1525156938e9f2::oink::OINK": {
    "name": "Sui Oink",
    "symbol": "OINK",
    "description": "Sui Oink",
    "iconUrl": "https://images.hop.ag/ipfs/QmSVLVKJgudnXQTyowHTJ8MTYV9SvtZrbVKk7QzcCFzRqh",
    "decimals": 6
  },
  "0x05ac5e4d6453c006d6ba0a6eb3936c457c55a4991058c9447a491abe5c40e184::snut::SNUT": {
    "name": "Sui Pnut",
    "symbol": "$SNUT",
    "description": "HAHAHAH YES !",
    "iconUrl": "https://images.hop.ag/ipfs/Qmbnpp7WVCfooUVPHWgQZAaPaHPDT2SpAuizdwiAiJjna4",
    "decimals": 6
  },
  "0x29efdd11e52333fcbb5536a5fc712bc75ac3194b5b95b4e274c6e34e730a13fe::gecko::GECKO": {
    "name": "GECKO on SUI",
    "symbol": "GECKO",
    "description": "just a gecko with legit community",
    "iconUrl": "https://images.hop.ag/ipfs/QmSff1TKR7VtLhDzUxCFS2xnLNCv8aiKeZoetM5a9dd8GR",
    "decimals": 6
  },
  "0x697fc97c0de6b64510e61b11ec60a9c4205f751ed94fc6d6838c9947e6dc328e::wtef::WTEF": {
    "name": "Wojack depressed",
    "symbol": "Wtef ",
    "description": "Why can’t I do anything on this site.\n\nHop.sad",
    "iconUrl": "https://images.hop.ag/ipfs/QmeSo8MBND3SfGGKpqYwCR7dktkQj92Fu3ezC2W5bToSLz",
    "decimals": 6
  },
  "0x25a5067fe76bd40c6cd2f7ba2a04d45477cf7a49e375e80def752ce125e624fd::lfog::LFOG": {
    "name": "Light Dog",
    "symbol": "LFOG",
    "description": "Light on the Innu Dog",
    "iconUrl": "https://images.hop.ag/ipfs/QmXy5wRB1PFZqGBJCigqXdww5izCsU6DG2AuUdqJds4m7R",
    "decimals": 6
  },
  "0x21dcbcd2ce9263efcd3696928a92373f9dd1b93bbaa06a86f04bdc0f35539f1d::moodeng::MOODENG": {
    "name": "Moo Deng",
    "symbol": "MOODENG",
    "description": "just a viral lil hippo",
    "iconUrl": "https://images.hop.ag/ipfs/Qmf1g7dJZNDJHRQru7E7ENwDjcvu7swMUB6x9ZqPXr4RV2",
    "decimals": 6
  },
  "0xd7defed8761e92af73394fcd804c7be0b679b6734e32895637fa479be26f995e::fkhopfun::FKHOPFUN": {
    "name": "FUCK HOPFUN",
    "symbol": "FKHOPFUN",
    "description": "FUCK THIS DELAYED BAD LAUNCH",
    "iconUrl": "https://images.hop.ag/ipfs/QmeJTNVE4rNjHtkpua8ELKt4PjD7Rhijf468bZcy8gqw9v",
    "decimals": 6
  },
  "0xb341a7ed07449a8f86a0378f3b833eda00d973567d51dbeed92c319d66e33434::maga::MAGA": {
    "name": "DonaldTrump",
    "symbol": "MAGA",
    "description": "TRUMP FOR PRESS",
    "iconUrl": "https://images.hop.ag/ipfs/QmdebTpGuM5H7CVH8Msy67PbgTdZ3Poz4vNbcgbwxmbuk5",
    "decimals": 6
  },
  "0x54a779744e32118a487d71025d063c35dfb1180041127638f5e3c264bf1cc709::sui::SUI": {
    "name": "Soviet Union International",
    "symbol": "SUI",
    "description": "Soviet Union International ($SUI) is here for the people and powered by memes! 💼 The motherland needs YOU! Step forward and shape history! 🌍",
    "iconUrl": "https://images.hop.ag/ipfs/QmbsLdRdxRWrMhrxebDqbasS4pDppP3u1THrfoEroVk6wQ",
    "decimals": 6
  },
  "0x4c31cf94bef22b75517b2ad72ef384b7f0cf64497607ce4e7309191aa8f79f1c::mother::MOTHER": {
    "name": "MOTHER IGGY",
    "symbol": "MOTHER",
    "description": "No you are… *$MOTHER is the only ticker I’m associated with. there are no derivatives. beware of scams*",
    "iconUrl": "https://images.hop.ag/ipfs/QmPeS4kL3Fns3714U5uZt4jQc8yeGXY1JGRHHDLATn2Ydn",
    "decimals": 6
  },
  "0xe8e02c6048c67dec3afa3726e40f8207635499fcbe3d5924c1df2dba28c748fc::jinbe::JINBE": {
    "name": "$JINBE",
    "symbol": "JINBE",
    "description": "$JINBE first son of the SUI.\n\nWhen it comes to the Sui ecosystem, $Jinbe’s got the currents on his side. The ocean’s deep, but his wallet's deeper!",
    "iconUrl": "https://images.hop.ag/ipfs/QmSChK2Nui2F9DjuDQASU1XAcgkcqoJ933zgbejrhohEpk",
    "decimals": 6
  },
  "0x187e8490e84627a8d57ddfdc090f260e3194f3a56369aa50484df69057eddb09::rex::REX": {
    "name": "suirex",
    "symbol": "rex",
    "description": "suirex",
    "iconUrl": "https://images.hop.ag/ipfs/QmdHT3ngjCqfn7wq8n668cnrqkaJ8YFQwt5GAghiQ9STuq",
    "decimals": 6
  },
  "0x8bf9579b40562dd897bbcd9ab584846948184ac9a02c386820e3940eaf0ca4df::rachop::RACHOP": {
    "name": "Rachopsui",
    "symbol": "Rachop",
    "description": "With my bunny by my side, Rachop and I take the ride,\nOn long journeys we glide, through #SUI, our dreams collide!",
    "iconUrl": "https://images.hop.ag/ipfs/QmSUPaCSxT5YU6vRgwnKte1VmxHzQVTHcazNx4kwQZpk6B",
    "decimals": 6
  },
  "0x6893472b18b487c8ea91ee6620689d7b8f7b132e70c8e43b1cf6e4659926963b::wst::WST": {
    "name": "We Stand Together",
    "symbol": "WST",
    "description": "We are the people of crypto\nWe represent a dream. A rebellion. A cause greater than any one person or entity",
    "iconUrl": "https://images.hop.ag/ipfs/QmQG9P7QDw7BufufZW8r16xNRKHvReYC3Cgwgngjq4ufDH",
    "decimals": 6
  },
  "0x045eb84651748cc77f5f57a0a99f8892b33947ccd2c43835fb5e2f9a5b27f834::fkhopfun::FKHOPFUN": {
    "name": "FUCK HOPFUN",
    "symbol": "FKHOPFUN",
    "description": "CLOWN LAUNCH",
    "iconUrl": "https://images.hop.ag/ipfs/QmeJTNVE4rNjHtkpua8ELKt4PjD7Rhijf468bZcy8gqw9v",
    "decimals": 6
  },
  "0xd4d7b1b2109d012a7d79d55408326b94dad86ed4f503ec00274657ad1320c5a7::wen::WEN": {
    "name": "Emperor Wen",
    "symbol": "Wen",
    "description": "Ruler of Sui Dynasty",
    "iconUrl": "https://images.hop.ag/ipfs/QmaFJxFzvkSa4JTJ63tTmWmSpPyhWhq9GJLUyve7NpeuV1",
    "decimals": 6
  },
  "0x97a17493acd0a87f86b3e6343d7677a812285b0ccfdd8831a57dc966a60a94b7::rainy::RAINY": {
    "name": "Rainy",
    "symbol": "RAINY",
    "description": "Let's make it $RAINY on sui 💧",
    "iconUrl": "https://images.hop.ag/ipfs/QmW1t9epgxG7zjUJJqUCwFhVppndPD8fR6QuSS21uUHsvz",
    "decimals": 6
  },
  "0x8dc05188d7452fcc30e34b4a7ad0154c99fc4b85808031aa6471dcba78e891be::idog::IDOG": {
    "name": "instadog",
    "symbol": "idog",
    "description": "Instadog (IDOG) 🐾✨ — The memecoin celebrating the first dog ever posted on Instagram! Instadog honors the iconic pup that started the social media revolution!",
    "iconUrl": "https://images.hop.ag/ipfs/QmawnaomKYfiWaVdV8tqnDQLh3WKBRFCPEbxNWLinYJATJ",
    "decimals": 6
  },
  "0x7617147aa3f5e1c86b329c5f945a6d735be5e3ee15dc1897d64bcc07d0d3a06e::boi::BOI": {
    "name": "Boi",
    "symbol": "$BOI",
    "description": "Boi on Sui",
    "iconUrl": "https://images.hop.ag/ipfs/QmVKNL7YQh94J1pUzwUqg8hdg84bkLjr2HXi6CEFk1zpK4",
    "decimals": 6
  },
  "0x2571d74caefb669e9f16db46dce095ca6771bb7b1dbb0edcd68969ab7a485cc0::supepe::SUPEPE": {
    "name": "SUI PEPE",
    "symbol": "$SUPEPE",
    "description": "Blue? or Green? You choose your pepe, biggest on chain token by....shhhh Founder reveal on the biggest twitter page of SUI!",
    "iconUrl": "https://images.hop.ag/ipfs/QmYZov3JKTf297uSHqgMevQZ4TJYzoUH83NhFxriq7yQWV",
    "decimals": 6
  },
  "0xe868a2dc6fcb745345f5b8c7618b56866c8659ced912b1d126d7aea7c0b257c4::amigos::AMIGOS": {
    "name": "Amigos",
    "symbol": "AMIGOS",
    "description": "The meme coin on sui that’s so laid-back, it’s practically horizontal.",
    "iconUrl": "https://images.hop.ag/ipfs/QmbX9NveFjtYhDuetVw1RAXQbQAMHpB7W5ygZ1RTTjYHed",
    "decimals": 6
  },
  "0x11c1d0cfe6da6a96d007cc2d39d1cdc3eca4c7313cbe652a2384b6dfc2539600::hopdog::HOPDOG": {
    "name": "HOPDOG",
    "symbol": "HOPDOG",
    "description": "First dog on sui CHAIN",
    "iconUrl": "https://images.hop.ag/ipfs/QmZZCMGkhJwLmFRrpQiLKSHzqnRP9kgvRqbEKxBLJkpoBB",
    "decimals": 6
  },
  "0xadd24186047e4018ba387d5f78854477f2c2bdb5be1416fff4007f7c3e3bba58::neiro::NEIRO": {
    "name": "Neiro",
    "symbol": "Neiro",
    "description": "Neiro in Sui",
    "iconUrl": "https://images.hop.ag/ipfs/QmYUJWWhREqjnXfuAtoeW66sEeiMUXBirbnNZ6RqK1XKsp",
    "decimals": 6
  },
  "0xe300dd5a63077e1ee302b8cb1f5ef1676a1ce45e21fefa3e54e38764df66dc8a::yarra::YARRA": {
    "name": "YARRAAA",
    "symbol": "YARRA",
    "description": "YARRRAAAAAAAAAAAA YARRRAAAAAAA",
    "iconUrl": "https://images.hop.ag/ipfs/QmP9GZWakzUY6sGazfQBwXoEJT4FAousRpL1RjpGoCYxiB",
    "decimals": 6
  },
  "0xe684d9e8c901c35030f4e7a9c0175f71eca65ac4689d97203556a623a54a62db::brett::BRETT": {
    "name": "sui mascot",
    "symbol": "BRETT",
    "description": "sui mascot",
    "iconUrl": "https://images.hop.ag/ipfs/QmWDYu5WFZDSfpKEFAaQw3BUDavKEoWA91E7kuTYHwNpPi",
    "decimals": 6
  },
  "0xfdd0ddc5096be4cbe753f5ffdb391c164992e00a1c21fedd028d2a038b2809a1::rachop::RACHOP": {
    "name": "RacHop",
    "symbol": "RacHop",
    "description": "RacHop",
    "iconUrl": "https://images.hop.ag/ipfs/QmSGQUYn4MkWnYhmmQCaf7rWtNrd4zcsRt1UWr9HFZ1Mgp",
    "decimals": 6
  },
  "0x24e040e6f0f269aade7207155d26cdb7da07a0f46289a64527e63267cb1c23ca::three34::THREE34": {
    "name": "FWOG",
    "symbol": "334",
    "description": "hi",
    "iconUrl": "https://images.hop.ag/ipfs/QmaSj7FAiStKkgRE1KZVmRYrN7wytKDWe3zZttUHAW6PPB",
    "decimals": 6
  },
  "0xb8145b6587f2b17a9b782d9eabc6001083068b34f641f88dfa0e577806c009ed::test::TEST": {
    "name": "test",
    "symbol": "test",
    "description": "test",
    "iconUrl": "https://images.hop.ag/ipfs/QmUkdBKYUp8MWVFb6RUPaZafYnvsKmrW47Aoab7J94Lz73",
    "decimals": 6
  },
  "0x3fe111cdd7bc13c7d6f48fdf0b84c5bab19b635702220d07c9ed175dfdaa1a19::hopcat::HOPCAT": {
    "name": "HopCAT",
    "symbol": "HopCAT",
    "description": "HopCAT",
    "iconUrl": "https://images.hop.ag/ipfs/QmVuuFiSKSdFKUNH2c3FgeHDqmGWqakRv6Tzo263R9wvbb",
    "decimals": 6
  },
  "0xa58539760a542fc4107adb73d2d91bdf0a9a75c2ef62544b343af7099b43281c::rachop_sui::RACHOP_SUI": {
    "name": "rachop",
    "symbol": "rachop_sui",
    "description": "Hop fun first racoon",
    "iconUrl": "https://images.hop.ag/ipfs/QmSGQUYn4MkWnYhmmQCaf7rWtNrd4zcsRt1UWr9HFZ1Mgp",
    "decimals": 6
  },
  "0x8113ffdb18c0fe9d81f5f72d478d25643c50397b4f7e4df73bf087b0fa273799::hopecat::HOPECAT": {
    "name": "HopeCat",
    "symbol": "Hopecat",
    "description": "First cat on Sui",
    "iconUrl": "https://images.hop.ag/ipfs/QmZypMAezu4x4ipm3i19AjGBChe8qxbovtTM7GJ8yVAqw8",
    "decimals": 6
  },
  "0x26c671063205375267d330ed58a2c62bfc0165a9dc8802a82b99698401a5bcf6::bean::BEAN": {
    "name": "Sui Bean",
    "symbol": "BEAN",
    "description": "BEAN me up, Scotty!",
    "iconUrl": "https://images.hop.ag/ipfs/Qmci1TMNqoHovfPdCm1mnEetb9S9XycQpT3yYmqGGCD3x1",
    "decimals": 6
  },
  "0xb66a3f06aaa5204eb71c04bd9ea73794411f9c386bde587c251ca9e80071d79a::lr::LR": {
    "name": "Landlord Ronald",
    "symbol": "LR",
    "description": "Ronald rules over the Boys Club by MATT FURIE, collecting rent from Brett, Pepe, Andy and Landwolf. Chaos is constant: the boys throw loud parties, the hallway reeks of marijuana, and neighbors are up in arms about Doogle, who’s desperate to join the in-crowd and is notorious for carrying weapons. Ronald, the wealthiest $LANDLORD on Sui",
    "iconUrl": "https://images.hop.ag/ipfs/QmWHuuMWwCXd2ubez9Q5u1QauN2XDbu2simgin54XaSsgf",
    "decimals": 6
  },
  "0x709d735de93025c2289536a4b0b5d4744937486e634c25c502652d59b7a2d750::funny::FUNNY": {
    "name": "$FUNNY Bunny",
    "symbol": "$FUNNY",
    "description": "The Bunny is setting up the stage",
    "iconUrl": "https://images.hop.ag/ipfs/QmWmcL1puULZohnKkJVwkbdUvrpUiNVXDs1SqfoyMWKfpm",
    "decimals": 6
  },
  "0xc98ce498cc15857665e95071e75966c48621af092413a5a7fc14fffaa50739a6::oink::OINK": {
    "name": "SuiOink",
    "symbol": "OINK",
    "description": "🐽🤿 $OINK is the most adventurous pig, scuba-diving into SUI Ocean 💧",
    "iconUrl": "https://images.hop.ag/ipfs/QmSVLVKJgudnXQTyowHTJ8MTYV9SvtZrbVKk7QzcCFzRqh",
    "decimals": 6
  },
  "0xedd5ad84bb24fcb045ca8a8ffb60503863c4977229fc4b8cb9cf1258d5180a05::dragon::DRAGON": {
    "name": "DragonBall",
    "symbol": "Dragon",
    "description": "A giant dragon is about to be born, look forward to it together!",
    "iconUrl": "https://images.hop.ag/ipfs/QmfV4hUT8yxWs4wotYEUYfAh4gyfcBBHUqYgvfzVAuGRii",
    "decimals": 6
  },
  "0xccf0256bd92e90df640fd754e5f4f07aab11c91477d460207932d6ec98ece252::hoptr::HOPTR": {
    "name": "HOPTRUMP",
    "symbol": "HOPTR",
    "description": "Next president of hop.fun",
    "iconUrl": "https://images.hop.ag/ipfs/QmcN9sisk3LacH4UnEE9RWcJFfoVMbjSqtnR5GUpYUWESZ",
    "decimals": 6
  },
  "0x8d4afd8bd1783ba08f02caa4f76aea118aa9febcdfe70103f21b8abae22bec85::hnf::HNF": {
    "name": "HopNOTfun",
    "symbol": "$HNF",
    "description": "This token was created to protest the stupid hop team. All profits will be distributed among the holders. FUCK OFf HOP!!!",
    "iconUrl": "https://images.hop.ag/ipfs/QmTiNDjZ4hd1uJchrb48jEDqaZDiRNFpbfzKL3seP3so4K",
    "decimals": 6
  },
  "0x488b22bacd390bf8663c4184d22ecdc2bdb3665abfd51636d2416907b3758e58::hopium::HOPIUM": {
    "name": "HOPIUM",
    "symbol": "HOPIUM",
    "description": "HOPIUM FOR HOP",
    "iconUrl": "https://images.hop.ag/ipfs/QmU9QrCTAucUpYtFY3uKkE9WyVSDqoJAJDh4H2pvxoaBeS",
    "decimals": 6
  },
  "0xa7a4516b13f1200f2596ae8c8c54cf66dce04d77ac9dc13ac8dd24c9ad41df87::test::TEST": {
    "name": "test",
    "symbol": "test",
    "description": "test",
    "iconUrl": "https://images.hop.ag/ipfs/QmUkdBKYUp8MWVFb6RUPaZafYnvsKmrW47Aoab7J94Lz73",
    "decimals": 6
  },
  "0x2ebaee2adeb35d69b67c4be97e0f7826a476e557119eb8150d8f48d788cb6f4d::sdg::SDG": {
    "name": "Suidogigog",
    "symbol": "SDG",
    "description": "lets gooo",
    "iconUrl": "https://images.hop.ag/ipfs/Qmf6Dv1zTqzoj8kJa4G7NX2VDxNmuXkRjrmPqnTPSWrACf",
    "decimals": 6
  },
  "0x10a20e3c7702f8695e2695bad4dae90686b3449e7757d3c839c52f62a05c3200::spinach::SPINACH": {
    "name": "Popeye",
    "symbol": "SPINACH",
    "description": "Popeye the Sailor Man",
    "iconUrl": "https://images.hop.ag/ipfs/QmSdG4LtSM11izmJqLGr96PkgUjwL6NeZATfm9TD85zbv5",
    "decimals": 6
  },
  "0x2e7a242a0fb88f15067663530c089e2aca3d0b735d77467b59cc38884f605423::sink::SINK": {
    "name": "Sinky",
    "symbol": "Sink",
    "description": "Sui water so sweet..",
    "iconUrl": "https://images.hop.ag/ipfs/QmVbNVJGkpyftqjBwuxEH8RLA4uPFsUQQr7ctTVnAERrQA",
    "decimals": 6
  },
  "0xf4a56139e94f2ea5211e18abe04955890fa6ad276a81fe8f57624ed0299280bc::phase::PHASE": {
    "name": "PHASE",
    "symbol": "PHASE",
    "description": "PHASE 420 69",
    "iconUrl": "https://images.hop.ag/ipfs/QmdkSLte4wYi4q4sQ8ZVTkRdPFoshigydPWA18ARdEFeQE",
    "decimals": 6
  },
  "0x8b579e9bfac3c80c4569f45fc6cb1b0963abc99c5bc29fca74743161ed57e7e3::hoptr::HOPTR": {
    "name": "HOPTRUMP",
    "symbol": "HOPTR",
    "description": "Next president of hop.fun",
    "iconUrl": "https://images.hop.ag/ipfs/QmcN9sisk3LacH4UnEE9RWcJFfoVMbjSqtnR5GUpYUWESZ",
    "decimals": 6
  },
  "0x273aa5fb15a5efe60c095501a0c4848b4248ef9a5b47de1591a7bfde6d1291ab::hf::HF": {
    "name": "Hopfun",
    "symbol": "HF",
    "description": "Swap on \n@SuiNetwork\n for the best price with zero fees.",
    "iconUrl": "https://images.hop.ag/ipfs/QmUdqYWffGtgH38hqcZEjiPjMtYKUKuwFba4yrDusbvowZ",
    "decimals": 6
  },
  "0x7ea88f012cd76fa4f3c355a093142349e6578800404398b3bda3747e196cc491::test::TEST": {
    "name": "test",
    "symbol": "test",
    "description": "test",
    "iconUrl": "https://images.hop.ag/ipfs/Qmb18A9c3jtfzThWbeA4YXvuW5vw4YEASTUsJzDGxCLCxq",
    "decimals": 6
  },
  "0x2fb62edccc779568e706c57f20df1e4491508ac459d88f8176a848c0dbc0cd9c::hoppy::HOPPY": {
    "name": "Hoppy The Kangaroo",
    "symbol": "HOPPY",
    "description": "The Unofficial Hop Fun Mascot!",
    "iconUrl": "https://images.hop.ag/ipfs/QmYwrdy1dzk4bt6EDLv6P37yQVApNzvT9L3e83N8xVvPT2",
    "decimals": 6
  },
  "0x014fcd3b7aef4aefd3bbea702cad21b9b067dace47948df66c2d2df85c604cc5::hopeless::HOPELESS": {
    "name": "Hopeless",
    "symbol": "Hopeless",
    "description": "Imagine creating a site with a team that is said to be professional, but losing to Rug Solana developers",
    "iconUrl": "https://images.hop.ag/ipfs/Qma9tD4Htzyd7iNoohVNcqwSpuLCn67SVwDHJGEANxwhop",
    "decimals": 6
  },
  "0xf221f685b78721105507881eb9e477ab3ee1f638095ab839b74db318e27000b1::cock::COCK": {
    "name": "Cock",
    "symbol": "COCK",
    "description": "The first visit in https://hop.ag/fun",
    "iconUrl": "https://images.hop.ag/ipfs/QmTMTAPxfEC9Fu9rgS7qGZsE3Bt49BnTkDc8yK11fbP1ig",
    "decimals": 6
  },
  "0x76bcf6691912aa12d6a4f78eb663caa0b9565841f6490391bc2335848f56d525::soil::SOIL": {
    "name": "Soil",
    "symbol": "SOIL",
    "description": "$SOIL will absorb all of Sui",
    "iconUrl": "https://images.hop.ag/ipfs/QmPs1DZQxpfav8AKpBSJtVuwx8JnF2LPb5XeAEYCQqdoJo",
    "decimals": 6
  },
  "0xd6a5c62104964d02f9067966c13fa9bad9a01bb2dea3112b4b0548c031d3263f::mutt::MUTT": {
    "name": "Mutt Coin",
    "symbol": "MUTT",
    "description": "Mutt Coin on Sui because everything better with a Mutt 🐶",
    "iconUrl": "https://images.hop.ag/ipfs/QmPoUkTqLhdZSVsV1FpUFaUPYVFPNyUcV6TSMdRr4Z2Hut",
    "decimals": 6
  },
  "0xfe9c4da07b85ddc57a55e24bc6f27bf745c08164abe1fd8cf7633701d4145515::suihui::SUIHUI": {
    "name": "SUIHUI",
    "symbol": "SUIHUI",
    "description": "The first $HUI on $SUI!",
    "iconUrl": "https://images.hop.ag/ipfs/Qmc7VRGYen2ZET1Vi1gf5punGbL2EQCH7vzYvwSiFcvKhn",
    "decimals": 6
  },
  "0x500414f02b1ab466fc6cedcd919b196fb16fb518ce8f41a30b67ffe0c271b45a::hoppy::HOPPY": {
    "name": "HOPPY official mascot",
    "symbol": "HOPPY",
    "description": "HOPPY official mascot of hopfun",
    "iconUrl": "https://images.hop.ag/ipfs/QmQPNiA8v8E3C2zZVd4FL6AE8WuRuGhSyyn9cpwr1Q2voM",
    "decimals": 6
  },
  "0xa1c3fd23b950a174bee3c7f4eae1ba827e0136664b56b3bd184080dd9e2c6a6d::lr::LR": {
    "name": "Landlord Ronald",
    "symbol": "LR",
    "description": "Ronald rules over the Boys Club by MATT FURIE, collecting rent from Brett, Pepe, Andy and Landwolf. Chaos is constant: the boys throw loud parties, the hallway reeks of marijuana, and neighbors are up in arms about Doogle, who’s desperate to join the in-crowd and is notorious for carrying weapons. Ronald, the wealthiest $LANDLORD on Sui",
    "iconUrl": "https://images.hop.ag/ipfs/QmWHuuMWwCXd2ubez9Q5u1QauN2XDbu2simgin54XaSsgf",
    "decimals": 6
  },
  "0x39f1f00cd5ed97b6325c77e1e39e80dc7d4da84e01c4c914b23528d6c0094af1::pepe::PEPE": {
    "name": "sui pepe",
    "symbol": "PEPE",
    "description": "meme token on sui",
    "iconUrl": "https://images.hop.ag/ipfs/QmZMTnCNYincJTTtNvxptHGEnB36F336C544Re5Zjo2QLj",
    "decimals": 6
  },
  "0xe5d4be4c98a6d7394e12a930cbd5c021c2891084947bdf04233b5f59977a59f3::soil::SOIL": {
    "name": "Soil",
    "symbol": "SOIL",
    "description": "$soil will absorb all of sui",
    "iconUrl": "https://images.hop.ag/ipfs/QmPs1DZQxpfav8AKpBSJtVuwx8JnF2LPb5XeAEYCQqdoJo",
    "decimals": 6
  },
  "0x9f2c6d3a517c6e0609efcc75188c5770c1a7460610d6236da65f49b4ddb055dd::dik::DIK": {
    "name": "Dog In Kimono",
    "symbol": "$Dik",
    "description": "$Dik the dog is a Dog In a Kimono. He brings the luck. No Twitter, Just a ticker!\nDik the Dog isn't just another Dog in the crypto world. It's a symbol of hope and resilience, representing the potential for dogs to thrive in a dog-eat-cat world now on sui",
    "iconUrl": "https://images.hop.ag/ipfs/QmZsaDzbPs5NbGCRkN2HhnnBEavmDviESkxWYJbcVXPitY",
    "decimals": 6
  },
  "0xe190779272ac22ef8b9b6a10990cff4a5662c2fdc686db71f1969553cf7868aa::one8::ONE8": {
    "name": "KEBAP",
    "symbol": "18",
    "description": "TURKISH KEBAB",
    "iconUrl": "https://images.hop.ag/ipfs/QmdrrsA2EygRRBGmVz91UUHbLmJ23Vn1SFnzw1nLGG5dat",
    "decimals": 6
  },
  "0x64af15d1f1c8501f60de4eb36cba8b6f357a2efd926a6d3aa68f1ec1de3c33be::rachop::RACHOP": {
    "name": "Rachop",
    "symbol": "RACHOP",
    "description": "With my bunny by my side, Rachop and I take the ride,\nOn long journeys we glide, through #SUI, our dreams collide!",
    "iconUrl": "https://images.hop.ag/ipfs/Qmc3C2biRPk3rnwDtvtF9k2HUwgUPf9DiELJJmqQCMqYhv",
    "decimals": 6
  },
  "0x78e1cc63cd28dc6c886976dc4263bd85add1c164b361470fe329d34df93844d8::suiyans::SUIYANS": {
    "name": "Super Suiyan",
    "symbol": "SUIYANS",
    "description": "It's the Super Suiyan cycle",
    "iconUrl": "https://images.hop.ag/ipfs/QmQHXY6CFB9iCvec79J5YqUMozWzcMFuEfqYK658MuBJ2S",
    "decimals": 6
  },
  "0xea6b409ac68337887ab57f920463444f1e399dbf577069cd381406ff92c947d2::one::ONE": {
    "name": "1",
    "symbol": "1",
    "description": "1",
    "iconUrl": "https://images.hop.ag/ipfs/QmPsK3ecsE6Jq6yZfQBciDJCfrSzA3a1Rgq1cRPDnyHZyt",
    "decimals": 6
  },
  "0x8eadff2506baf4b4c79e7326bb2650ca38ea4e0cb8063f2da665977bd4531f0c::pnut::PNUT": {
    "name": "PNUT",
    "symbol": "PNUT",
    "description": "PNUT",
    "iconUrl": "https://images.hop.ag/ipfs/QmRmSLrceFLKSQp84UqSvqbe8nJbdhQgSq3gJWQPKcb9aM",
    "decimals": 6
  },
  "0x1947dfcb062fb1dcd42c831ac24fdcf2383c9f5965873379469093e1e79e1b67::fun::FUN": {
    "name": "FUN Token",
    "symbol": "FUN",
    "description": "01st HOP fun",
    "iconUrl": "https://images.hop.ag/ipfs/QmQvXjD3avEQMvYHXK9qGPA4ixefgNqQXLfyd5PiLz3y35",
    "decimals": 6
  },
  "0xe43e92c53edeeb19fa01095f89de25340434f54e0f7bbb2d31332792f979744f::hpd::HPD": {
    "name": "HOPDOG",
    "symbol": "HPD",
    "description": "woof woof",
    "iconUrl": "https://images.hop.ag/ipfs/QmRKatJ6xRmwJ3NaK24UT4WwU7YqCXTWQbTf4AZGnHPn8R",
    "decimals": 6
  },
  "0xefda13da8caabe1146e55e2559693517e3220f537d379b4f711fc09a4ee530f7::degen::DEGEN": {
    "name": "Sui Degen Apes",
    "symbol": "$Degen",
    "description": "Degens wanting to degen",
    "iconUrl": "https://images.hop.ag/ipfs/QmYeV5vyP5ePiWtDZAZWY2dUHBY3qc1FXtETUArGhY5uYi",
    "decimals": 6
  },
  "0xf39dba2f6ccc84f5ebf5e81b463241464acde7a840a9b93b8a53c298248154d9::rabbit::RABBIT": {
    "name": "rabbit",
    "symbol": "rabbit",
    "description": "rabbit",
    "iconUrl": "https://images.hop.ag/ipfs/QmU2tKh2oSARzAn1YgtnhZac33qxDukZtu8SGQr9B1AwM8",
    "decimals": 6
  },
  "0x88035636d2787e6247d9ea5cd2ca6b8c5ff3d78fa09e32b37ac4208a2db64cf3::hpd::HPD": {
    "name": "HOPDOG",
    "symbol": "HPD",
    "description": "WOOF WOOF",
    "iconUrl": "https://images.hop.ag/ipfs/QmRKatJ6xRmwJ3NaK24UT4WwU7YqCXTWQbTf4AZGnHPn8R",
    "decimals": 6
  },
  "0x69e631f10afe33ce8a9dea104ec1d09a5e6ff2b06c1462d5d0aa4a3a6f6cae96::orcie::ORCIE": {
    "name": "ORCIE",
    "symbol": "orcie",
    "description": "orcie",
    "iconUrl": "https://images.hop.ag/ipfs/Qmbrtoc1nZpMbbRUGQbSXwMQkVN4GBtRysMFKZG26BYxpC",
    "decimals": 6
  },
  "0x34bc4064a21d6d01add70d473b9fac30816b74e2b7e0f701e6290f8dcb74b3cf::uturn::UTURN": {
    "name": "UTU",
    "symbol": "UTURN",
    "description": "@UTURN",
    "iconUrl": "https://images.hop.ag/ipfs/QmRk1tn7dbk7gkNYwMtDsvVJhQuCpz5aYBTcDqDLZzzuSQ",
    "decimals": 6
  },
  "0x3cd3a9681a2cef93e9d34dce6d4bdff4061bc9a13d6b67231d5cad2bf6af3a36::heppo::HEPPO": {
    "name": "suiheppo",
    "symbol": "heppo",
    "description": "Your heppo with a beautiful smile",
    "iconUrl": "https://images.hop.ag/ipfs/QmT9cAytovJL78nHHhq6cJJRiQPXNgaj9aLg8DjuRkiqKq",
    "decimals": 6
  },
  "0x0e35ce0e73f055567569b441413ff6ec3d9197220acccab5af45567af84dabd6::bubl::BUBL": {
    "name": "BUBL",
    "symbol": "BUBL",
    "description": "Bubbling on SuiNetwork to make frens.",
    "iconUrl": "https://images.hop.ag/ipfs/Qmbf752EhgUbtLCph7R9XuGYiWQZKdyCaA8Bo5afpwuNiY",
    "decimals": 6
  },
  "0x3523fcb1c7edccac316d0e0686dd367ca5fd4da756555a15590c4a65ca7dfe0b::angrybob::ANGRYBOB": {
    "name": "BOB",
    "symbol": "AngryBoB",
    "description": "Many things piss off Bob, He’s an angry dog 😤",
    "iconUrl": "https://images.hop.ag/ipfs/QmQQSmJMymLqNgnXRXqUmbaGaxWguh9AxnCahSZ8WADNVd",
    "decimals": 6
  },
  "0x15c5ab249affeaea72bf47e21fd680c86ac032d5f09fb60876fc06701ba9dc87::volo::VOLO": {
    "name": "Volo",
    "symbol": "Volo",
    "description": "DAO",
    "iconUrl": "https://images.hop.ag/ipfs/QmSazt7U4M5rYv7RLdf7YXwKAPVdYqSaADXnZmhGWribUu",
    "decimals": 6
  },
  "0x090ee54ad2b815ec7d56900ce4900d1922c1b885bee757c1fca0734a39580cc0::wizd::WIZD": {
    "name": "WIZARD OG",
    "symbol": "WIZD",
    "description": "FIRST WIZARD MEME",
    "iconUrl": "https://images.hop.ag/ipfs/QmYM5HoWMH7N65SmbqVJHDYkooNujto1bfkEzGoVb3oZxe",
    "decimals": 6
  },
  "0x88794d0af53179c1f76484261b668719d5d6d0c12ea5884ab11c620102c10485::fun::FUN": {
    "name": "F.U.N",
    "symbol": "FUN",
    "description": "F is for friends who do stuff together\nU is for U and Me\nN is for anywhere and anytime at all down here in the deep blue sea",
    "iconUrl": "https://images.hop.ag/ipfs/QmUF6Yv6fB9XpFSTdYx1yCayyJ6CanWVq39MQoumYnKn2x",
    "decimals": 6
  },
  "0x09d3832f62120537c716a86119e31f14b339072106bf0b617b6678cde75385d5::pant::PANT": {
    "name": "PEANUT",
    "symbol": "PANT",
    "description": "two the moon",
    "iconUrl": "https://images.hop.ag/ipfs/QmZBrpBA4jMMhRAQoiB5mGKC9xQVZYab2V167RHZpGgvKz",
    "decimals": 6
  },
  "0xdad0aa74735bb23eabf9611aa0f58f538b54c2828c0cc042248db73455400d5c::obi::OBI": {
    "name": "Obi PNut Kenobi",
    "symbol": "OBI",
    "description": "“If you strike me down, I will become more powerful than you could possibly imagine” Obi PNut Kenobi",
    "iconUrl": "https://images.hop.ag/ipfs/QmUgNB1eRwLHapVXTwZ1gib5DZRfXUwx1WYqgrBD43QoRF",
    "decimals": 6
  },
  "0x68686c7e954af2d8bfef2ab43447ae1534005d8a56b7ebd0340fe7cfda7f5d85::suiba::SUIBA": {
    "name": "Suibabot",
    "symbol": "Suiba",
    "description": "Suibabot - Sui's fastest bot to trade coins, and SUIBA's official Telegram trading bot.",
    "iconUrl": "https://images.hop.ag/ipfs/QmSwXgMoo5rydqLo5XZVXAuoRxCmBnAs5dJmWR3SscZibe",
    "decimals": 6
  },
  "0x1b87e5992c418340030ec22a3e49eb81e69ad63377f4f4c9c03658a31cb849d5::vio::VIO": {
    "name": "Avior Protocol",
    "symbol": "VIO",
    "description": "High Powered Multi-Chain Protocol",
    "iconUrl": "https://images.hop.ag/ipfs/QmPhNF5RNkBxNrR87TtMGtxsNiRnCs28rDFDDW5J2SXRiZ",
    "decimals": 6
  },
  "0x2fb79cde4a36a766816dc3e1b0ad88683e33756ebcb27004ebcc7a0472d81c95::neo::NEO": {
    "name": "Neo Whale",
    "symbol": "NEO",
    "description": "Neo Whale, an AI-integrated entity built on the SUI blockchain, is a digital guardian navigating the vast seas of decentralized data.",
    "iconUrl": "https://images.hop.ag/ipfs/QmW8tZQKxvxLXLkpzbWDwVAEEVc7ySFXhNgdQB1SS24h2p",
    "decimals": 6
  },
  "0x9d206c5ff2b1f2a345b07430a30dde901ae14403109cebf366ecef04a6eec2c4::mock::MOCK": {
    "name": "SpongeMock",
    "symbol": "MOCK",
    "description": "retardio",
    "iconUrl": "https://images.hop.ag/ipfs/QmP5Ei3RVwqCNNfnzrJWBMKCoyc25115HBNMoxWXNDiVfQ",
    "decimals": 6
  },
  "0xea398b97b3f729947fefba36ba10b72c068506576a24de4aac295473b281c42c::fun::FUN": {
    "name": "FUN",
    "symbol": "FUN",
    "description": "F is for Friends who do stuff together\nU is for You and Me\nN is for anywhere at anytime at all down here in the deep blue sea",
    "iconUrl": "https://images.hop.ag/ipfs/QmUF6Yv6fB9XpFSTdYx1yCayyJ6CanWVq39MQoumYnKn2x",
    "decimals": 6
  },
  "0x4c38555a991be7546cd234b80d7a55200abe7372574e685625fb7e2ededeb049::carts::CARTS": {
    "name": "CARROTS",
    "symbol": "CARTS",
    "description": "only carrots",
    "iconUrl": "https://images.hop.ag/ipfs/QmXBkNQMW6Wm3KQNAifBeY4FR984oxx4sNjnYTvpHDgeuk",
    "decimals": 6
  },
  "0x4ee15df8d94d2baffe9c97af966becad31260229553d4052b50a1d385aee4ae6::test::TEST": {
    "name": "test",
    "symbol": "test",
    "description": "test",
    "iconUrl": "https://images.hop.ag/ipfs/QmQvXjD3avEQMvYHXK9qGPA4ixefgNqQXLfyd5PiLz3y35",
    "decimals": 6
  },
  "0x5fb8214653416e84c3db4facb913614e2bdda5fda701a429651ae06b9bb395d5::suirrell::SUIRRELL": {
    "name": "Peanut the Suirrell",
    "symbol": "Suirrell",
    "description": "Most famous Suirrell!",
    "iconUrl": "https://images.hop.ag/ipfs/Qmepmeh8sXCpdwc3aUWLEYtAKf6FvJsFLRkB2YtuV5NcMt",
    "decimals": 6
  },
  "0x4802c573538c7eb1f68a891be873e65a9b7e8a54dbee80bcedf1f371672d399d::ronado::RONADO": {
    "name": "CRISTIANO RONADO",
    "symbol": "RONADO",
    "description": "I AM RONADO\nI WIN EUROS\nI AM PORTUGA\nSUUIIIII",
    "iconUrl": "https://images.hop.ag/ipfs/Qmf9wTVZEsGyiXDGzzjxgta3vtSuHPkfzoUxXZqdo5LQ8Q",
    "decimals": 6
  },
  "0x787b548acdec73eeb214faedd6dbf0b0f8db1178612bd503e3a0fe2ad0eb235b::bullrun::BULLRUN": {
    "name": "bull run",
    "symbol": "bullrun",
    "description": "To the moon",
    "iconUrl": "https://images.hop.ag/ipfs/QmRMX3Rvp9A1cM6uhBTRHa5oyfpyapUcW1ckZSDQvgPhQY",
    "decimals": 6
  },
  "0xbfbbd1227efb5625a0197b6311ced381aa26d84cdb92316e204a8c22e7e5d0a4::cat::CAT": {
    "name": "cat",
    "symbol": "cat",
    "description": "cat",
    "iconUrl": "https://images.hop.ag/ipfs/QmSBEtrrD5pQea3TesfXNiUf5UYxZme313KjsprnbpSvxo",
    "decimals": 6
  },
  "0xd40a1393dfc8371afe006c491c1301044c9198a4a5b978d2453657423e5cf886::plop::PLOP": {
    "name": "PlopSui",
    "symbol": "Plop",
    "description": "Official mascot of $SUI chain 💧 plop plop 💧",
    "iconUrl": "https://images.hop.ag/ipfs/QmTE4PwdbR5ffPETh6uTbrMBRp9TWvNtBjguG9Ky8NsqaM",
    "decimals": 6
  },
  "0x322b8f9589af8abc842335c63e3e3edb5a07ddd38f75c4b7237272e3e885385b::hophop::HOPHOP": {
    "name": "Hophop",
    "symbol": "hophop",
    "description": "hophophophophophophophophophophophophophop",
    "iconUrl": "https://images.hop.ag/ipfs/QmbuMR86qgMrUYzxQTVHbdT4o8Ugc2ZnyaUG8v4YSHCKnu",
    "decimals": 6
  },
  "0xf5f8bfb68c78a6dab9e42b7110cb72728864195747fe04af795713c02d82ec50::cult::CULT": {
    "name": "SUICULT",
    "symbol": "CULT",
    "description": "Secret societies and global influences...",
    "iconUrl": "https://images.hop.ag/ipfs/QmTuTjNCwKmvVYt4DfmU31vGCX3UhPfJ16vDTMmTazwuUs",
    "decimals": 6
  },
  "0xc4389723b8a5fd75810470aa21d1558cda50132a10cb82c62489ce4e82be8f2d::rachop::RACHOP": {
    "name": "RacHop",
    "symbol": "Rachop",
    "description": "SUIs #1 Racoon!\nWith my bunny by my side, Rachop and I take the ride..",
    "iconUrl": "https://images.hop.ag/ipfs/QmSGQUYn4MkWnYhmmQCaf7rWtNrd4zcsRt1UWr9HFZ1Mgp",
    "decimals": 6
  },
  "0x177f6ca5fc36e9db2c6578dee08b6f36b07d6d01f747c8d526c2539d65d2a7a3::ursa::URSA": {
    "name": "Ursa",
    "symbol": "Ursa",
    "description": "Ursa - a simple token about a friendly bear delivering some pizza. Choppy market has done him bad.",
    "iconUrl": "https://images.hop.ag/ipfs/QmQdcPJPTjijSd2SernGAANhTVrzZ5kfw55eUZiAedpj2u",
    "decimals": 6
  },
  "0x8998927eae653684f09ad972f154a05ce48d2521bbd41878c48411440babd244::pepepepe::PEPEPEPE": {
    "name": "SuiPepe",
    "symbol": "PePePePe",
    "description": "Suipepe for the Culture",
    "iconUrl": "https://images.hop.ag/ipfs/Qma5cG1SeSywxE5jWvRPr889mVtJVVa6QNHsHFBUpwutdk",
    "decimals": 6
  },
  "0xe73e47ab83ac74a9c20e0a50719cc4edf7a1f4762fba51d57cd6e74fed8b9984::hnf::HNF": {
    "name": "Hop not fun",
    "symbol": "HNF",
    "description": "Token to express dissatisfaction with hop fun launch",
    "iconUrl": "https://images.hop.ag/ipfs/QmdLRnEHuS5JR44BJE91Ucao1E7iCYoQ67z5ABvToXk82y",
    "decimals": 6
  },
  "0x5e621dbc4a3a2ae67cb1320e45720a0422c40d13fdad420b653b6e28edd99f79::hopdog::HOPDOG": {
    "name": "hopdog",
    "symbol": "HOPDOG",
    "description": "$HOPDOG, Sui's and Hop's best dog!",
    "iconUrl": "https://images.hop.ag/ipfs/QmRKatJ6xRmwJ3NaK24UT4WwU7YqCXTWQbTf4AZGnHPn8R",
    "decimals": 6
  },
  "0x8110c3175efd16ef3b9f9c0ddfc3cf8cfd138d2897263b05f94c14306dd9f3ff::gat::GAT": {
    "name": "GATTIVO",
    "symbol": "GAT",
    "description": "bad mothfuking son of a bitch cat but better than the mother of hopfun devs",
    "iconUrl": "https://images.hop.ag/ipfs/QmNrKaQozVsdD9Wj5z22WaKZCD2hq9wAAEaCSZEunuj7Lo",
    "decimals": 6
  },
  "0x47e827d7270ecc768a5fb0701ce91592137782377d9ce73f80052510828c1ac0::dinui::DINUI": {
    "name": "Dinosaurs on SUI",
    "symbol": "DINUI",
    "description": "Dinosaurs on SUI",
    "iconUrl": "https://images.hop.ag/ipfs/QmSoqSoLBNZKXXmA6KuLaSeJKr9LpVeqYkpJv1HFuNbg5M",
    "decimals": 6
  },
  "0x56f2cfc358d64c666d9cddfbf2baeb1cbd54fefb7db3ade5b97dfe33d6b7367e::hoppnut::HOPPNUT": {
    "name": "Peanut the Squirrel",
    "symbol": "HopPnut",
    "description": "Today we rally to help the plight of Peanut the Squirrel who is facing euthanisiation from the state",
    "iconUrl": "https://images.hop.ag/ipfs/QmPMsM4mssGg77fgKnE9mW58f6xaDFiPnm9YLfgc9HGay9",
    "decimals": 6
  },
  "0x1feaebfa9d9f82dee64441316eacb07a92d8ea189a105b15e45aa9d9d6d14aa9::suitor::SUITOR": {
    "name": "SUITOR",
    "symbol": "SUITOR",
    "description": "SUITOR",
    "iconUrl": "https://images.hop.ag/ipfs/QmVF94XX4638p9on4vVtdgGVUd5SvZ86accADaca4e42m4",
    "decimals": 6
  },
  "0xb3df40059f631bba20b03c11c51cbb8a7f353edc9132fee3a4cfd362354ae510::hopbunny::HOPBUNNY": {
    "name": "HOPBUNNY",
    "symbol": "HopBunny",
    "description": "1st meme of HopFun",
    "iconUrl": "https://images.hop.ag/ipfs/QmearEhnZ62uzEeHUHY3r4B1QQENYW8ykzBnHc6UGo1tfD",
    "decimals": 6
  },
  "0xba1dfab87087bf2411db8e6cbf3df9498a735a9f1b2d69a8e4cd8ecbbc932e2c::hopcat::HOPCAT": {
    "name": "Hop Cat",
    "symbol": "HOPCAT",
    "description": "The first cat on Hop Fun",
    "iconUrl": "https://images.hop.ag/ipfs/QmWBx34wFTPcn8u2RbT74MgrKZFcbssptPYZWZhHQM3oVt",
    "decimals": 6
  },
  "0x0c01d494aa6049ed87c5f7f2b81b9525446ad9a63c7bfd95d68bc8188a629190::era::ERA": {
    "name": "SUI ERA",
    "symbol": "ERA",
    "description": "This token only for FUN sui Community",
    "iconUrl": "https://images.hop.ag/ipfs/QmcayForoZDr6vagigwmag9nUmUhaM33qBJVngjdLb4j5v",
    "decimals": 6
  },
  "0x8aebe57c78013f64a2c5d5683d200663dde6553ce883c81b3cd7621519de9144::luce::LUCE": {
    "name": "Sui’s Mascot LUCE",
    "symbol": "$LUCE",
    "description": "HAHAHA YES !!!!\nWebsite and Telegram coming soon !!!",
    "iconUrl": "https://images.hop.ag/ipfs/QmeEerocjzoMJDiP2rNaSTzHvB5XkhLT9LyLKAkYsZ4Kqs",
    "decimals": 6
  },
  "0x732f1d3341851051be61c25713eb44f68896e1742df8f10cfdd5b939e65ebd61::chopsui::CHOPSUI": {
    "name": "CHOPSUI",
    "symbol": "CHOPSUI",
    "description": "Don't eat it.",
    "iconUrl": "https://images.hop.ag/ipfs/QmeyJbuLUBhj1nrV9kn3afsujZut4x97JU9bsqdNUQsbt4",
    "decimals": 6
  },
  "0x8fbdc608dbcec5ee18e49934baacef5becd2f5aa284d11b0857d9a485687ef1b::dos::DOS": {
    "name": "DogsOnSui",
    "symbol": "DOS",
    "description": "Dogs together strong",
    "iconUrl": "https://images.hop.ag/ipfs/QmT2M1dW8UfUJjxi8WZGD7w1FTzJnVekDSTKcuTjXFqEUe",
    "decimals": 6
  },
  "0xf278951a03feb79005ad981f47663cb268dc1a3cdc0ab053c2e8c347d1401aa3::jepe::JEPE": {
    "name": "JEPE",
    "symbol": "JEPE",
    "description": "the most memeable jellyfish on the internet",
    "iconUrl": "https://images.hop.ag/ipfs/QmXsmK5y1Tzd3D3qXAHQjd2LRXsZUXEoYBcbLUahTcg6vd",
    "decimals": 6
  },
  "0x0de6b7bd017a1e2fb43053abb08681f70356408f6c31ba1b6df2a5912b57d751::prts::PRTS": {
    "name": "Pirates",
    "symbol": "PRTS",
    "description": "First Pirates Token On SUI",
    "iconUrl": "https://images.hop.ag/ipfs/QmVypqKBU4pLBbDZxo9P1dcVEUHneLpf27DmfKS6Q8Mdnu",
    "decimals": 6
  },
  "0xd96e1f97176b2a1133ab41e0cc6e7bcc1b9f4d967bab7a18ae93780eda12c4b1::chopsui::CHOPSUI": {
    "name": "CHOPSUI",
    "symbol": "CHOPSUI",
    "description": "Don't eat it. It's not for you.",
    "iconUrl": "https://images.hop.ag/ipfs/QmeyJbuLUBhj1nrV9kn3afsujZut4x97JU9bsqdNUQsbt4",
    "decimals": 6
  },
  "0x4975e31d8bad38e7f2bc2c4fe60e867e8d03e8161d4d29e78cbfd853555ec1c1::pizza::PIZZA": {
    "name": "pizzasui",
    "symbol": "pizza",
    "description": "Fun for $Pizza Sui Hop Fun",
    "iconUrl": "https://images.hop.ag/ipfs/QmXCcVQRykHR9LJQi3a1jSLD7mT4aqDw2u418WqkRM7FFj",
    "decimals": 6
  },
  "0xf220961763b9243b16c802492549d6f4a718c70db746b1dead7a4e1f55a5bb77::fcksol::FCKSOL": {
    "name": "Fuck Sol",
    "symbol": "fcksol",
    "description": "Fuck Sol, SUI Supremacy",
    "iconUrl": "https://images.hop.ag/ipfs/QmTPxkdrNWGsdxYrFuQ2vVqWbq34UHkkfpH5VrKX3Sdc52",
    "decimals": 6
  },
  "0xe12542839dcb2df132368890996a4a01463ae7fff8ec90a9bdc96909c1d61f83::smn::SMN": {
    "name": "Suizle My Nizzle",
    "symbol": "SMN",
    "description": "Yo, it’s Suizle My Nizzle, where we flip coinzzles, makin’ stacks drizzle, ridin’ that Sui fizzle! With SMN, you know it’s the real puzzle, droppin’ crypto heat like a nozzle. So grab your wallet, my nizzle, and let’s make them gains sizzle!",
    "iconUrl": "https://images.hop.ag/ipfs/QmekntpGV7MVAymRKLBNrvUoEtABSyKoWV3sncw3n8UieX",
    "decimals": 6
  },
  "0x8578d346930d15f8786dc7ef67f7866b6d09035e42cc7408778a515535507878::meowlk::MEOWLK": {
    "name": "MEOWLK",
    "symbol": "MEOWLK",
    "description": "DRINK SOME MEOWLK",
    "iconUrl": "https://images.hop.ag/ipfs/QmaAF9uTKVwL5FhEjqoP6mNbxLi7jyWp2G7A7PWBx1zDPs",
    "decimals": 6
  },
  "0x510e9fedd07643e706214408fa4671db0ff0d962e6268844ba65440e2227d5bf::ssy::SSY": {
    "name": "supersuiyan",
    "symbol": "SSY",
    "description": "It's the Super Suiyan cycle",
    "iconUrl": "https://images.hop.ag/ipfs/QmQHXY6CFB9iCvec79J5YqUMozWzcMFuEfqYK658MuBJ2S",
    "decimals": 6
  },
  "0xcfd3096f8f90fed61e269a5da6157d097ab4648c0dfb9a4a3c3a8e7963e6531c::trumpnut::TRUMPNUT": {
    "name": "TRUMP THE SQUIRREL",
    "symbol": "$TRUMPNUT",
    "description": "The squirrel president.",
    "iconUrl": "https://images.hop.ag/ipfs/Qmc5Vdhnd32987yD6AuwEq9E17ZLyr71fPo5oNfTKuE4aZ",
    "decimals": 6
  },
  "0xa31adbab149e8641620ac948f548d1ee13b90622e241f914dcdf2773876d0810::nasdaq_4200::NASDAQ_4200": {
    "name": "Stonks",
    "symbol": "NASDAQ4200",
    "description": "Stonks",
    "iconUrl": "https://images.hop.ag/ipfs/QmVcjLaB2WgcqdtRFrni7ZNfi2RqM222WaG7T4ZH2FFZcD",
    "decimals": 6
  },
  "0x5c63b3c6b905f88cdb9bcc2569476b18de8f2be7cdfd89d62875ddbbf5638a47::fun::FUN": {
    "name": "FUN Token",
    "symbol": "FUN",
    "description": "FUN",
    "iconUrl": "https://images.hop.ag/ipfs/QmQvXjD3avEQMvYHXK9qGPA4ixefgNqQXLfyd5PiLz3y35",
    "decimals": 6
  },
  "0x57a59d2a0234c18da932f943ddeb42c2c1c8edd258da527df88340627a3c7b4b::luce::LUCE": {
    "name": "Sui’s Mascot LUCE",
    "symbol": "$LUCE",
    "description": "HAHAHAH YES !\nwebsite and tg coming soon",
    "iconUrl": "https://images.hop.ag/ipfs/QmeEerocjzoMJDiP2rNaSTzHvB5XkhLT9LyLKAkYsZ4Kqs",
    "decimals": 6
  },
  "0x8e8ed27534097bfad0395cf6048a2038d0596e2eaec00b836793aa993613c92f::zxfa::ZXFA": {
    "name": "SFKSA",
    "symbol": "ZXFA",
    "description": "1",
    "iconUrl": "https://images.hop.ag/ipfs/QmbsPuRTycWmNLvRLthrkMMDY6r4DAAkX5vJM7rU55Y5RR",
    "decimals": 6
  },
  "0xf2a267a05604f665960b0ca5c3dc22550ed9643d70549119202cf23fe68a6670::suipreme::SUIPREME": {
    "name": "Suipreme",
    "symbol": "SUIPREME",
    "description": "Suipreme",
    "iconUrl": "https://images.hop.ag/ipfs/QmWCLr7TLrhSR1XtdpX4ymDSkpAnkiwWQvGByB3H3m2Byy",
    "decimals": 6
  },
  "0x1d0798647403bd8286320158002cdc5ff13e5b99858e078496963a76f1a7d7de::suicam::SUICAM": {
    "name": "Sui Camel",
    "symbol": "SUICAM",
    "description": "The more camels you have, the richer you will be.",
    "iconUrl": "https://images.hop.ag/ipfs/QmWEFmaApMk8JNzfPBPjoa7jXdkX6TEXmMyvX5wKj9Fc3Z",
    "decimals": 6
  },
  "0xfc4ce3b927206c95e9596e56f79a8754cdb5099ef673fb468881d34973b655b4::hotter::HOTTER": {
    "name": "Hotter",
    "symbol": "Hotter",
    "description": "Hottest Otter on Sui Ocean",
    "iconUrl": "https://images.hop.ag/ipfs/QmdijSRgKiEJ3Ps1kifv28c4thzBKqdywHnz2CGTYceYT4",
    "decimals": 6
  },
  "0x317a7a090b6788654db4a66cd6161b83bc777e44ba9ef442a0208926f52efaf8::sdoge::SDOGE": {
    "name": "Sui Doge",
    "symbol": "Sdoge",
    "description": "From Community to community",
    "iconUrl": "https://images.hop.ag/ipfs/QmSPjQH4nom6qS7ZyV7eB76spSUDSnDCFyasvUbgutcaKx",
    "decimals": 6
  },
  "0x9f541a9a7fd929141df7397d0a055fb3fe95923e6213f82cfe80a269b63a38c6::pepe::PEPE": {
    "name": "Pepe",
    "symbol": "pepe",
    "description": "PeP PEp pEP",
    "iconUrl": "https://images.hop.ag/ipfs/QmeDbTg7ktsBoQwRU7t7oxhPrBXd56dPU7FEdkQh1qkoAT",
    "decimals": 6
  },
  "0x57144a76174e6e154a0035fd3d5579221c4741c7e877f4f597f5b19ff61df8c2::ban::BAN": {
    "name": "Comedian",
    "symbol": "Ban",
    "description": "The most significant meme of the art history by Mauricio Cattelan. Being auction at Sotheby’s November 20th 🍌",
    "iconUrl": "https://images.hop.ag/ipfs/QmXdp7yUDHKR9UMhwRkeMe681aKRwAJxBxappQPjuYw9dd",
    "decimals": 6
  },
  "0x27568cfb7c28c8f5bf838c2b9ffa52a023b241df21baa6f0ef19d3a1d030dc69::fubao::FUBAO": {
    "name": "Hop Fubao",
    "symbol": "FUBAO",
    "description": "I’m HopFubao, the cutest meme inspired by none other than Fubao Panda!\n\nI’m about to hop onto SUI and I can’t wait to bring a little bit of panda love and joy to your moon bag!\n\nIf you’re looking for something playful and fun, I’m here to make your portfolio a whole lot cuter!",
    "iconUrl": "https://images.hop.ag/ipfs/QmSn1MHFmeNtQGxoCSp5H2zQrNeqGmNUd98XmFRnh8n2qd",
    "decimals": 6
  },
  "0x44251553b776d40dd9bf47a21ef346ec1fc92744eaae138cae57a8f78a7758c2::trump::TRUMP": {
    "name": "TRUMP",
    "symbol": "TRUMP",
    "description": "$TRUMP",
    "iconUrl": "https://images.hop.ag/ipfs/QmNh9eDkt994Jus4zUKZejcYp3tSEjUndXUZ8KbPgkrfzT",
    "decimals": 6
  },
  "0x52a5242fabb7c1136cdc8b4480fab3a540207228f173255ed79b42d155577a2e::four20690::FOUR20690": {
    "name": "Pepe",
    "symbol": "420690",
    "description": "pePE PePe",
    "iconUrl": "https://images.hop.ag/ipfs/QmT3n9QXTeUQFVHGUGBeKCDT94S5HJJ6jHtYk99V5RyQ14",
    "decimals": 6
  },
  "0x9cd1754646c095963087261ed9a27b04292426f7f98a3e4923d5a1bfe29874f8::rkamala::RKAMALA": {
    "name": "Retardio Kamala",
    "symbol": "rKamala",
    "description": "Kamala has gone full retardio!",
    "iconUrl": "https://images.hop.ag/ipfs/QmPefXawJjkrfWEk2Pt8hQUHFmCEZCbCBYbaf3p7naQnea",
    "decimals": 6
  }
}`
