package schema

import (
	"math/big"
	"time"
)

type Swap struct {
	Tx        string // 原始hash
	Token     string // 买卖的token
	Direction int    // 0买; 1卖

	Operator string // msg.sender0
	Receiver string //接收token的人，一般和operator相等, 但也可能为合约

	Price         *big.Int
	AmountInToken *big.Int
	CreatedAt     time.Time
}
