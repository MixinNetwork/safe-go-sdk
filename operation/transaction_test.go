package operation

import (
	"encoding/hex"
	"log"
	"testing"

	"github.com/MixinNetwork/go-safe-sdk/bitcoin"
	"github.com/stretchr/testify/assert"
)

func TestSignTx(t *testing.T) {
	assert := assert.New(t)
	SignSafeTx("raw transaction hash data", "private key", byte(1))

	raw := "70736274ff0100a402000000016daf0a2ca612879093698c5ab6dbcff372e893137d5dfda23615e1489f5e07210000000000ffffffff0310270000000000002200204a8f0888cc30695a20c71ae0d119f4c09743c0d03a7db52774d06c49a52d081a905f0100000000002200204a8f0888cc30695a20c71ae0d119f4c09743c0d03a7db52774d06c49a52d081a0000000000000000126a104525b641cc6e4ed1b2bb0713b786da6b000000000001007d0100000001a93da9d71875dac610ece59c07a6574211bba14adf01ecf6b601de93538b5f5f0000000000ffffffff02a0860100000000002200204a8f0888cc30695a20c71ae0d119f4c09743c0d03a7db52774d06c49a52d081a1dbd00000000000016001409eb71fab8358daa65e08d2e148568b83fe075880000000001012ba0860100000000002200204a8f0888cc30695a20c71ae0d119f4c09743c0d03a7db52774d06c49a52d081a22020208134c3bb3263598db7f28cb631b34f81d34bfdf3cee163da7c41b6434e92fad47304402206c9adbfea684f9dca42700db018a6aaebbee1f679f553e871351031ccdbff3510220064eeed0c51e0a018b4c275e6585fd81d686c106be1708507a1bd4813affb7a88101030481000000010578210208134c3bb3263598db7f28cb631b34f81d34bfdf3cee163da7c41b6434e92fadac7c2103e17978200e8961fc87358898db7b0d5686aa4f14935d418de9b533d14922a4b3ac937c8292632103c8f64e27a2f3ae961a57184841df19e7d8708ddbc998f0c5abc7197ead70931fad02b001b2926893528722060208134c3bb3263598db7f28cb631b34f81d34bfdf3cee163da7c41b6434e92fad04f5c895ac220603c8f64e27a2f3ae961a57184841df19e7d8708ddbc998f0c5abc7197ead70931f04d3886f84220603e17978200e8961fc87358898db7b0d5686aa4f14935d418de9b533d14922a4b30483f05fb600010178210208134c3bb3263598db7f28cb631b34f81d34bfdf3cee163da7c41b6434e92fadac7c2103e17978200e8961fc87358898db7b0d5686aa4f14935d418de9b533d14922a4b3ac937c8292632103c8f64e27a2f3ae961a57184841df19e7d8708ddbc998f0c5abc7197ead70931fad02b001b2926893528722020208134c3bb3263598db7f28cb631b34f81d34bfdf3cee163da7c41b6434e92fad04f5c895ac220203c8f64e27a2f3ae961a57184841df19e7d8708ddbc998f0c5abc7197ead70931f04d3886f84220203e17978200e8961fc87358898db7b0d5686aa4f14935d418de9b533d14922a4b30483f05fb600010178210208134c3bb3263598db7f28cb631b34f81d34bfdf3cee163da7c41b6434e92fadac7c2103e17978200e8961fc87358898db7b0d5686aa4f14935d418de9b533d14922a4b3ac937c8292632103c8f64e27a2f3ae961a57184841df19e7d8708ddbc998f0c5abc7197ead70931fad02b001b2926893528722020208134c3bb3263598db7f28cb631b34f81d34bfdf3cee163da7c41b6434e92fad04f5c895ac220203c8f64e27a2f3ae961a57184841df19e7d8708ddbc998f0c5abc7197ead70931f04d3886f84220203e17978200e8961fc87358898db7b0d5686aa4f14935d418de9b533d14922a4b30483f05fb60000"
	log.Println(CheckTransactionPartiallySignedBy(raw, "0208134c3bb3263598db7f28cb631b34f81d34bfdf3cee163da7c41b6434e92fad"))
	h := "70736274ff0100a40200000001d7fd88df22604c77503345ac0f9c81307406253b280a69b2f5d5c7818ec7c7bf0100000000ffffffff03a08601000000000022002097d984899b9916b84a13fcdbf0d101e19ada5146130ddc06bfc99194a03bef38e093040000000000220020f31c53a74dfd597480c0605ed13c048fc5ea2b19f051a136b8bf960e33fcda2d0000000000000000126a10ea0d8a3188194fe5ac22cdef9e61a4a0000000000001012b801a060000000000220020f31c53a74dfd597480c0605ed13c048fc5ea2b19f051a136b8bf960e33fcda2d220203221c2eeb196e913e5d890f33b110ae6af6fbddac2a8b9921aaa5d8a60a5cd20146304402206743c5c4e9e1cd4f67a50fda53f1bef8382baf7a70d288f7269ddd608d2c632102203826c26f8750566d309bff1452906b33a315a7bd94deef0b2fa8c74caf76b4b8010304810000000105782103221c2eeb196e913e5d890f33b110ae6af6fbddac2a8b9921aaa5d8a60a5cd201ac7c21035684fa60a93e85f940a2445ab7f2fb1786a4b415ef60cb8c53e1970b50b80324ac937c82926321032cb1c1eef0e43e19c9571c37ad5548289b4916ee898d48ef8b806397cd45da92ad02c006b2926893528700000000"
	rawb, err := hex.DecodeString(h)
	assert.Nil(err)
	_, err = bitcoin.UnmarshalPartiallySignedTransaction(rawb)
	assert.Nil(err)
}
