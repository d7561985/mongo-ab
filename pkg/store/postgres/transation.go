package postgres

type Balance struct {
	AccountID     uint64
	Balance       int64
	DepositAllSum int64
	DepositCount  int32

	PincoinBalance int64
	PincoinsAllSum int64
}
