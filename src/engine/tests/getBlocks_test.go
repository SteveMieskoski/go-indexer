package tests

import (
	"src/engine"
	"testing"
)

// In this example, our client wishes to track the latest 'block number'
// known to the server. The server supports two methods:
//
// eth_getBlockByNumber("latest", {})
//    returns the latest block object.
//
// eth_subscribe("newHeads")
//    creates a subscription which fires block objects when new blocks arrive.

func TestSubscribeToBlocks(t *testing.T) {
	// Connect the client.

	engine.GetBlocks("ws://localhost:8546")
	t.Deadline()

}
