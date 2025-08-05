package snowflake

import (
	"go-actor/common/pb"
	"testing"
)

func TestUUID(t *testing.T) {
	Init(&pb.Node{Type: 10, Id: 10})

	t.Log(GenUUID())
	t.Log(GenUUID())
	t.Log(GenUUID())
}
