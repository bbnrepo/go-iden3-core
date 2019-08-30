package signedpacketsrv

import (
	"crypto/ecdsa"
	//"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	babykeystore "github.com/iden3/go-iden3-core/keystore"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/stretchr/testify/assert"

	"github.com/iden3/go-iden3-core/core"
	"github.com/iden3/go-iden3-core/services/discoverysrv"
	"github.com/iden3/go-iden3-core/services/nameresolversrv"
	"github.com/iden3/go-iden3-core/services/signsrv"
)

const debug = false

const passphrase = "secret"

// const relaySkHex = "4be5471a938bdf3606888472878baace4a6a64e14a153adf9a1333969e4e573c"
// const relaySkHex = "1cbb679f29fb298efd6e9e2d52405dd324ff9e75025e662683062b3db0dc06a9"
const relayIdHex = "113kyY52PSBr9oUqosmYkCavjjrQFuiuAw47FpZeUf"
const relaySkHex = "4be5471a938bdf3606888472878baace4a6a64e14a153adf9a1333969e4e573c"

const kSignSkHex = "9b3260823e7b07dd26ef357ccfed23c10bcef1c85940baa3d02bbf29461bbbbe"

//const idAddrHex = "0x970e8128ab834e8eac17ab8e3812f010678cf791"
// const idAddrHex = "0x308eff1357e7b5881c00ae22463b0f69a0d58adb"
const idHex = "11AVZrKNJVqDJoyKrdyaAgEynyBEjksV5z2NjZoPxf"

// root 0x1d9d41171c4b621ff279e2acb84d8ab45612fef53e37225bdf67e8ad761c3922
// (proofs generated in services/claimsrv/service_test.go)
const proofKSignJSON = `
{"proofs":[{"mtp0":"0x00030000000000000000000000000000000000000000000000000000000000061e7604b4b32e21da52f5f8a0ccf9709e378e033a9c1d458c4d426d57e53f629b2910a6fba42851f8282e0266c887e09db4fd84975a76a6c6ce468651683d2346","mtp1":"0x03030000000000000000000000000000000000000000000000000000000000061e7604b4b32e21da52f5f8a0ccf9709e378e033a9c1d458c4d426d57e53f629b2910a6fba42851f8282e0266c887e09db4fd84975a76a6c6ce468651683d234626c2f31d534b17f59353243b7d1d1a306602281bf07d5a50e4f6487af4ff27bb06d4571fb9634e4bed32e265f91a373a852c476656c5c13b09bc133ac61bc5a6","root":"0x278ddfbfb41d3862e164d6c38e8074ba90c4ba058bbe09520466595535eaa0cb","aux":{"version":0,"era":0,"id":"11AVZrKNJVqDJoyKrdyaAgEynyBEjksV5z2NjZoPxf"}},{"mtp0":"0x0000000000000000000000000000000000000000000000000000000000000000","mtp1":"0x030000000000000000000000000000000000000000000000000000000000000004c810bdf05e2786373c7ecfec6202a2b0891a382d2c0130ad8e798886cc8a571906be21558a5e6dc7ec8ee913fb33fc24df8de0a5cfe65903f4d79fb91a0c5e","root":"0x0c1e281a1dd5296ec25bebd1ba9017aa7a45a18a90e3c6f21902b4d4de201189","aux":null}],"leaf":"0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001e730fbf398dacc09c2bd5316f0a16563faa6d5750825e520655ceec440804340000000000000000000000000000000000000001000000000000000000000001","date":1567158430,"signature":"1e369372fab378873173a8abd2262ec80fd0c091ab67e31e667510037e8c2d23db9d91ea8c8713ca9fe5612e3305877b352d455ea3e74421f6ea6905c9692d03","signer":"113kyY52PSBr9oUqosmYkCavjjrQFuiuAw47FpZeUf"}
`

const ethName = "testName@iden3.eth"

//o0x0d1fadf720af488d10aaa4bdaf6a8d1163ad30b19624082b0e4403934ab57ff3
const proofAssignNameJSON = `
{"proofs":[{"mtp0":"0x00010000000000000000000000000000000000000000000000000000000000011807294329b1e53a4a9b23587ce40c7476e2476e969dd21f54d8164a08042d68","mtp1":"0x03010000000000000000000000000000000000000000000000000000000000012df1bac120f050facaebd732f32da77eac6d95c6e1189b9ba1a5b37e7d149e6204c810bdf05e2786373c7ecfec6202a2b0891a382d2c0130ad8e798886cc8a57022a1e2c3a59747c79b0cddee114e3bfb2d24777281ed568b364d43a6eea33a8","root":"0x003fe9e0b554e8330c0f939b468e2f459d79a22952d318f467b4183003d50066","aux":null}],"leaf":"0x0000000000000000000000000000000000000000000000000000000000000000000000d119a5c0b9fe1659620b8a635024d5ed0fed3cc9f5f20403a9ff480de400178118069763dbe18ad9c512b09b4f9a9b7ae14c4ead00200ceabdcbac85950000000000000000000000000000000000000000000000000000000000000003","date":1560761069,"signature":"ec632c79b2fdf12ab82e8d5a67af5001c22b33c030cce557530357863966651573f5c876622bcc7055bfb95ef86069f017f8b3ac16665b738ed83036609fd800","signer":"113kyY52PSBr9oUqosmYkCavjjrQFuiuAw47FpZeUf"}
`

const namesFileContent = `
{
  "iden3.eth": "113kyY52PSBr9oUqosmYkCavjjrQFuiuAw47FpZeUf"
}
`

const entititesFileContent = `
{
  "113kyY52PSBr9oUqosmYkCavjjrQFuiuAw47FpZeUf": {
    "name": "iden3-test-relay",
    "kOpPub": "117f0a278b32db7380b078cdb451b509a2ed591664d1bac464e8c35a90646796",
    "trusted": { "relay": true }
  }
}
`

var namesFilePath string
var entititesFilePath string

var signedPacketVerifier *SignedPacketVerifier
var signedPacketSigner *SignedPacketSigner

var dbDir string

// var keyStoreDir string
var keyStore *babykeystore.KeyStore
var relaySk babyjub.PrivateKey
var relayPkComp *babyjub.PublicKeyComp
var relayPk *babyjub.PublicKey

var kSignSk babyjub.PrivateKey
var kSignPkComp *babyjub.PublicKeyComp
var kSignPk *babyjub.PublicKey

var proofKSign core.ProofClaim

var relayId core.ID

var id core.ID

func genPrivateKey() {
	privateKeyECDSA, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	fmt.Println(hex.EncodeToString(crypto.FromECDSA(privateKeyECDSA)))
}

func setup() {
	//genPrivateKey()

	pass := []byte("my passphrase")
	storage := babykeystore.MemStorage([]byte{})
	keyStore, err := babykeystore.NewKeyStore(&storage, babykeystore.LightKeyStoreParams)
	if err != nil {
		panic(err)
	}

	if _, err := hex.Decode(relaySk[:], []byte(relaySkHex)); err != nil {
		panic(err)
	}
	if relayPkComp, err = keyStore.ImportKey(relaySk, pass); err != nil {
		panic(err)
	}
	if err := keyStore.UnlockKey(relayPkComp, pass); err != nil {
		panic(err)
	}
	if relayPk, err = relayPkComp.Decompress(); err != nil {
		panic(err)
	}

	if _, err := hex.Decode(kSignSk[:], []byte(kSignSkHex)); err != nil {
		panic(err)
	}
	if kSignPkComp, err = keyStore.ImportKey(kSignSk, pass); err != nil {
		panic(err)
	}
	if err := keyStore.UnlockKey(kSignPkComp, pass); err != nil {
		panic(err)
	}
	if kSignPk, err = kSignPkComp.Decompress(); err != nil {
		panic(err)
	}

	// common3.HexDecodeInto(idAddr[:], []byte(idAddrHex))
	id, err = core.IDFromString(idHex)
	if err != nil {
		panic(err)
	}

	namesFile, err := ioutil.TempFile("", "go-iden3-test-namesFile")
	if err != nil {
		panic(err)
	}
	namesFile.WriteString(namesFileContent)
	namesFilePath = namesFile.Name()
	namesFile.Close()

	nameResolverSrv, err := nameresolversrv.New(namesFilePath)
	if err != nil {
		panic(err)
	}

	entititesFile, err := ioutil.TempFile("", "go-iden3-test-entititesFile")
	if err != nil {
		panic(err)
	}
	entititesFile.WriteString(entititesFileContent)
	entititesFilePath = entititesFile.Name()
	entititesFile.Close()

	discoverySrv, err := discoverysrv.New(entititesFilePath)
	if err != nil {
		panic(err)
	}

	signSrv := signsrv.New(keyStore, *kSignPk)

	signedPacketVerifier = NewSignedPacketVerifier(discoverySrv, nameResolverSrv)

	if err := json.Unmarshal([]byte(proofKSignJSON), &proofKSign); err != nil {
		panic(err)
	}
	signedPacketSigner = NewSignedPacketSigner(*signSrv, proofKSign, id)
}

func teardown() {
	os.RemoveAll(dbDir)
	os.Remove(namesFilePath)
	os.Remove(entititesFilePath)
}

func TestSignedPacket(t *testing.T) {
	setup()
	defer teardown()

	t.Run("SignPacketV01", testSignPacketV01)
	t.Run("SignGenericSigV01", testSignGenericSigV01)
	t.Run("SignIdenAssertV01Name", testSignIdenAssertV01Name)
	t.Run("SignIdenAssertV01NoName", testSignIdenAssertV01NoName)
	t.Run("MarshalUnmarshal", testMarshalUnmarshal)

}

func BenchmarkSignedPacket(b *testing.B) {
	setup()
	defer teardown()

	b.Run("SignGenericSigV01", benchmarkSignGenericSigV01)
	b.Run("VerifySignGenericSigV01", benchmarkVerifySignGenericSigV01)
}

func testSignPacketV01(t *testing.T) {
	var proofKSign core.ProofClaim
	if err := json.Unmarshal([]byte(proofKSignJSON), &proofKSign); err != nil {
		panic(err)
	}
	if debug {
		fmt.Println(&proofKSign)
	}
	form := map[string]string{"foo": "baz"}
	signedPacket, err := signedPacketSigner.
		NewSignPacketV02(600, GENERICSIGV01, nil, form)
	assert.Nil(t, err)
	signedPacketStr, err := signedPacket.Marshal()
	assert.Nil(t, err)
	if debug {
		fmt.Println(signedPacketStr)
	}

	err = signedPacketVerifier.VerifySignedPacket(signedPacket)
	assert.Nil(t, err)
}

func testSignGenericSigV01(t *testing.T) {
	var proofKSign core.ProofClaim
	if err := json.Unmarshal([]byte(proofKSignJSON), &proofKSign); err != nil {
		panic(err)
	}
	form := map[string]string{"foo": "baz"}
	signedPacket, err := signedPacketSigner.
		NewSignGenericSigV01(600, form)
	assert.Nil(t, err)
	signedPacketStr, err := signedPacket.Marshal()
	assert.Nil(t, err)
	if debug {
		fmt.Println(signedPacketStr)
	}

	err = signedPacketVerifier.VerifySignedPacketGeneric(signedPacket)
	assert.Nil(t, err)
}

func benchmarkSignGenericSigV01(b *testing.B) {
	var proofKSign core.ProofClaim
	if err := json.Unmarshal([]byte(proofKSignJSON), &proofKSign); err != nil {
		panic(err)
	}
	form := map[string]string{"foo": "baz"}

	for n := 0; n < b.N; n++ {
		signedPacketSigner.
			NewSignGenericSigV01(600, form)
	}
}

// VerifySignedPacketGeneric is a bit slow right now.  The bottleneck resides
// in the VerifyProofClaim, in particular in the calculation of the mimc7.Hash.
// There is room for optimization in the mimc7.Hash
func benchmarkVerifySignGenericSigV01(b *testing.B) {
	var proofKSign core.ProofClaim
	if err := json.Unmarshal([]byte(proofKSignJSON), &proofKSign); err != nil {
		panic(err)
	}
	form := map[string]string{"foo": "baz"}
	signedPacket, err := signedPacketSigner.
		NewSignGenericSigV01(600, form)
	assert.Nil(b, err)

	for n := 0; n < b.N; n++ {
		signedPacketVerifier.VerifySignedPacketGeneric(signedPacket)
	}
}

func testSignIdenAssertV01Name(t *testing.T) {
	// Login Server
	nonceDb := core.NewNonceDb()
	requestIdenAssert := NewRequestIdenAssert(nonceDb, "example.com", 60)

	// Client
	var proofKSign core.ProofClaim
	if err := json.Unmarshal([]byte(proofKSignJSON), &proofKSign); err != nil {
		panic(err)
	}
	var proofAssignName core.ProofClaim
	if err := json.Unmarshal([]byte(proofAssignNameJSON), &proofAssignName); err != nil {
		panic(err)
	}
	signedPacket, err := signedPacketSigner.
		NewSignIdenAssertV01(requestIdenAssert,
			&IdenAssertForm{EthName: ethName, ProofAssignName: &proofAssignName}, 600)
	assert.Nil(t, err)
	signedPacketStr, err := signedPacket.Marshal()
	assert.Nil(t, err)
	if debug {
		fmt.Println(signedPacketStr)
	}

	// Login Server
	idenAssertResult, err := signedPacketVerifier.
		VerifySignedPacketIdenAssert(signedPacket, nonceDb, "example.com")
	assert.Nil(t, err)
	if debug {
		fmt.Println(idenAssertResult)
	}
}

func testSignIdenAssertV01NoName(t *testing.T) {
	// Login Server
	nonceDb := core.NewNonceDb()
	requestIdenAssert := NewRequestIdenAssert(nonceDb, "example.com", 60)

	// Client
	var proofKSign core.ProofClaim
	if err := json.Unmarshal([]byte(proofKSignJSON), &proofKSign); err != nil {
		panic(err)
	}
	signedPacket, err := signedPacketSigner.
		NewSignIdenAssertV01(requestIdenAssert, nil,
			600)
	assert.Nil(t, err)
	signedPacketStr, err := signedPacket.Marshal()
	assert.Nil(t, err)
	if debug {
		fmt.Println(signedPacketStr)
	}

	// Login Server
	idenAssertResult, err := signedPacketVerifier.
		VerifySignedPacketIdenAssert(signedPacket, nonceDb, "example.com")
	assert.Nil(t, err)
	if debug {
		fmt.Println(idenAssertResult)
	}
}

func testMarshalUnmarshal(t *testing.T) {
	var proofKSign core.ProofClaim
	if err := json.Unmarshal([]byte(proofKSignJSON), &proofKSign); err != nil {
		panic(err)
	}

	form := map[string]string{"foo": "baz", "bar": "biz"}
	signedPacket, err := signedPacketSigner.
		NewSignPacketV02(600,
			GENERICSIGV01, nil, form)
	assert.Nil(t, err)
	if debug {
		fmt.Println("\nSignedPacket:")
		fmt.Printf("Data: %#v\n", signedPacket.Payload.Data)
		fmt.Printf("Form: %#v\n", signedPacket.Payload.Form)
	}
	signedPacketStr, err := signedPacket.Marshal()
	assert.Nil(t, err)
	//if debug {
	//	fmt.Println("\nMarshal:")
	//	fmt.Println(signedPacketStr)
	//}
	var signedPacket2 SignedPacket
	err = signedPacket2.Unmarshal(signedPacketStr)
	assert.Nil(t, err)
	if debug {
		fmt.Println("\nUnmarshal:")
		fmt.Printf("Data: %#v\n", signedPacket2.Payload.Data)
		fmt.Printf("Form: %#v\n", signedPacket2.Payload.Form)
	}

	signedPacket3, err := signedPacketSigner.
		NewSignPacketV02(600, "invalid", nil, nil)
	assert.Nil(t, err)
	signedPacketStr2, err := signedPacket3.Marshal()
	assert.Nil(t, err)
	var signedPacket4 SignedPacket
	// "invalid" is not a valid signed packet type, so unmarshal must error
	err = signedPacket4.Unmarshal(signedPacketStr2)
	assert.Error(t, err)
}
