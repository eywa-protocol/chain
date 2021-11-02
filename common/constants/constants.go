/*
* Copyright 2021 by EYWA chain <blockchain@digiu.ai>
*/

package constants

import (
	"time"
)

// genesis constants
var (
	//TODO: modify this when on mainnet
	GENESIS_BLOCK_TIMESTAMP = uint32(time.Date(2020, time.August, 10, 0, 0, 0, 0, time.UTC).Unix())
)

// multi-sig constants
const MULTI_SIG_MAX_PUBKEY_SIZE = 16

// transaction constants
const TX_MAX_SIG_SIZE = 16

// network magic number
const (
	NETWORK_MAGIC_MAINNET = 0x8c6077ab
	NETWORK_MAGIC_TESTNET = 0x2ddf8829
)

// extra info change height
const EXTRA_INFO_HEIGHT_MAINNET = 2917744
const EXTRA_INFO_HEIGHT_TESTNET = 1664798

// eth 1559 heigh
const ETH1559_HEIGHT_MAINNET = 12965000
const ETH1559_HEIGHT_TESTNET = 10499401

const POLYGON_SNAP_CHAINID_MAINNET = 16
