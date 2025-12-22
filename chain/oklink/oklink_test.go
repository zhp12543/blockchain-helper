package oklink

import "testing"

func TestOkLink(t *testing.T)  {
	txid := []string{
		"0x8f9fdb6901545b8a9f738567117c1f5d414bcef60316801461fa627b5a1d13b0",
		"0x7e7b8a1639d3d1e2eea731f0b5db338c3526a5a9bcb83305e43b47f19bb702f1",
		"0x2b0971ed39c91104c8b9497751add26a9d0929d4951e390f936b811ce3c3069f",
		"0x2d96c8adae86e668fd7ce19d721a62e680fe5c12d55fc541e7f7bc8f8e648015",
	}
	CrawlOklink("bsc", txid, "socks5://127.0.0.1:51837")
}
